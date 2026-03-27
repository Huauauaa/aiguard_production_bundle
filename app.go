package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aiguard/internal/config"
	"aiguard/internal/llm"
	"aiguard/internal/logging"
	"aiguard/internal/review"
	"aiguard/internal/task"
	"aiguard/internal/uiapi"
	"aiguard/internal/workspace"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx   context.Context
	tasks *task.Manager
	orch  *review.Orchestrator
}

func NewApp() *App {
	return &App{
		tasks: task.NewManager(),
		orch:  review.NewOrchestrator(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) PullCode(req uiapi.StartReviewRequest) (uiapi.PrepareRepositoryResponse, error) {
	cfg, layout, logPath, _, err := a.resolveRuntime(req.ConfigPath, req.WorkspaceDir)
	if err != nil {
		return uiapi.PrepareRepositoryResponse{}, err
	}

	logging.Infof(logPath, "Pull code requested. mr_url=%s repo_url=%s local_repo_path=%s", req.MRURL, req.RepoURL, req.LocalRepoPath)
	prepared, err := a.orch.PrepareRepository(context.Background(), req, cfg, layout, review.RepoPrepareMode{CloneIfMissing: true, Fetch: true})
	if err != nil {
		logging.Errorf(logPath, "Pull code failed: %v", err)
		return uiapi.PrepareRepositoryResponse{}, err
	}

	source, target, branches, err := a.orch.SuggestBranches(context.Background(), prepared.GitDir)
	if err != nil {
		logging.Errorf(logPath, "Failed to detect active branches: %v", err)
		return uiapi.PrepareRepositoryResponse{}, err
	}

	logging.Infof(logPath, "Repository prepared successfully. repo=%s source=%s target=%s branch_count=%d", prepared.RepoURL, source, target, len(branches))
	return uiapi.PrepareRepositoryResponse{
		RepoURL:           prepared.RepoURL,
		SourceBranch:      source,
		TargetBranch:      target,
		AvailableBranches: branches,
		LogPath:           logPath,
		Message:           "代码已拉取完成，已自动识别推荐分支。",
	}, nil
}

func (a *App) StartReview(req uiapi.StartReviewRequest) (string, error) {
	cfg, layout, logPath, _, err := a.resolveRuntime(req.ConfigPath, req.WorkspaceDir)
	if err != nil {
		return "", err
	}
	if err := cfg.Validate(); err != nil {
		logging.Errorf(logPath, "Review precheck failed during config validation: %v", err)
		return "", err
	}
	if strings.TrimSpace(req.SourceBranch) == "" || strings.TrimSpace(req.TargetBranch) == "" {
		return "", errors.New("请先点击“拉取代码”或手动填写源分支和目标分支")
	}

	logging.Infof(logPath, "Start review requested. source_branch=%s target_branch=%s", req.SourceBranch, req.TargetBranch)
	prepared, err := a.orch.PrepareRepository(context.Background(), req, cfg, layout, review.RepoPrepareMode{CloneIfMissing: false, Fetch: false})
	if err != nil {
		logging.Errorf(logPath, "Review precheck failed while loading repository cache: %v", err)
		return "", err
	}
	if err := a.orch.ValidateBranches(context.Background(), prepared.GitDir, req.SourceBranch, req.TargetBranch); err != nil {
		logging.Errorf(logPath, "Branch validation failed: %v", err)
		return "", err
	}
	logging.Infof(logPath, "Branch validation passed.")

	llmClient := llm.New(cfg)
	if err := llmClient.Ping(context.Background()); err != nil {
		logging.Errorf(logPath, "LLM connectivity check failed: %v", err)
		return "", fmt.Errorf("LLM 接口联通性检查失败: %w", err)
	}
	logging.Infof(logPath, "LLM connectivity check passed.")

	taskID := uuid.NewString()
	runCtx, cancel := context.WithCancel(context.Background())
	a.tasks.Add(taskID, cancel, req)

	go func() {
		defer a.tasks.Done(taskID)

		emit := func(name string, payload any) {
			switch name {
			case "review:progress":
				if progress, ok := payload.(uiapi.ProgressEvent); ok {
					logging.Infof(logPath, "[%s %d%%] %s", progress.Stage, progress.Percent, progress.Message)
				}
			case "review:error":
				if info, ok := payload.(map[string]any); ok {
					if message, ok := info["message"].(string); ok {
						logging.Errorf(logPath, message)
					}
				}
			case "review:done":
				if done, ok := payload.(uiapi.ReviewDoneEvent); ok {
					logging.Infof(logPath, "Review finished successfully. report_dir=%s", done.ReportDir)
				}
			}
			runtime.EventsEmit(a.ctx, name, payload)
		}

		logging.Infof(logPath, "Review task started. task_id=%s", taskID)
		done, err := a.orch.Run(runCtx, taskID, req, emit)
		if err != nil {
			logging.Errorf(logPath, "Review task failed: %v", err)
			emit("review:error", map[string]any{
				"taskId":  taskID,
				"message": err.Error(),
			})
			return
		}

		emit("review:done", done)
	}()

	return taskID, nil
}

func (a *App) CancelReview(taskID string) error {
	return a.tasks.Cancel(taskID)
}

func (a *App) ListHistory(req uiapi.RuntimeContextRequest) ([]uiapi.HistoryItem, error) {
	_, _, _, workspaceDir, err := a.resolveRuntime(req.ConfigPath, req.WorkspaceDir)
	if err != nil {
		return nil, err
	}
	return a.orch.ListHistory(workspaceDir)
}

func (a *App) GetLogState(req uiapi.RuntimeContextRequest) (uiapi.LogState, error) {
	_, layout, logPath, _, err := a.resolveRuntime(req.ConfigPath, req.WorkspaceDir)
	if err != nil {
		return uiapi.LogState{}, err
	}
	content, err := logging.ReadTail(logging.ResolvePath(layout.Logs), 128*1024)
	if err != nil {
		return uiapi.LogState{}, err
	}
	return uiapi.LogState{
		LogPath:   logPath,
		Content:   content,
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func (a *App) ClearCache(req uiapi.RuntimeContextRequest) (uiapi.CacheClearResult, error) {
	if a.tasks.HasRunning() {
		return uiapi.CacheClearResult{}, errors.New("当前存在正在运行的任务，请先取消任务后再清理缓存")
	}

	_, _, _, workspaceDir, err := a.resolveRuntime(req.ConfigPath, req.WorkspaceDir)
	if err != nil {
		return uiapi.CacheClearResult{}, err
	}
	layout, removed, err := workspace.Clear(workspaceDir)
	if err != nil {
		return uiapi.CacheClearResult{}, err
	}
	logPath := logging.ResolvePath(layout.Logs)
	logging.Infof(logPath, "Workspace cache cleared successfully.")
	return uiapi.CacheClearResult{
		WorkspaceDir: workspaceDir,
		LogPath:      logPath,
		Removed:      removed,
		Message:      "缓存、仓库镜像、报告与日志已清理完成。",
	}, nil
}

func (a *App) resolveRuntime(configPath, workspaceOverride string) (config.Config, workspace.Layout, string, string, error) {
	cfg, _ := config.Load("")
	if strings.TrimSpace(configPath) != "" {
		cleanPath := filepath.Clean(configPath)
		loaded, err := config.Load(cleanPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return cfg, workspace.Layout{}, "", "", err
			}
		} else {
			cfg = loaded
		}
	}

	workspaceDir := strings.TrimSpace(workspaceOverride)
	if workspaceDir == "" {
		workspaceDir = cfg.Review.WorkspaceDir
	}
	workspaceDir = filepath.Clean(workspaceDir)
	layout, err := workspace.Prepare(workspaceDir)
	if err != nil {
		return cfg, layout, "", "", err
	}
	logPath := logging.ResolvePath(layout.Logs)
	if err := logging.EnsureFile(logPath); err != nil {
		return cfg, layout, "", "", err
	}
	return cfg, layout, logPath, workspaceDir, nil
}
