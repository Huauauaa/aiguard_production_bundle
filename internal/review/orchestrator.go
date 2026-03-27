package review

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"aiguard/internal/config"
	"aiguard/internal/findings"
	"aiguard/internal/gitops"
	"aiguard/internal/history"
	"aiguard/internal/llm"
	"aiguard/internal/model"
	"aiguard/internal/packer"
	"aiguard/internal/projectctx"
	"aiguard/internal/provider"
	"aiguard/internal/report"
	"aiguard/internal/scanner"
	"aiguard/internal/uiapi"
	"aiguard/internal/workspace"
)

type Orchestrator struct {
	git    *gitops.Manager
	locker *workspace.Locker
}

type PreparedRepo struct {
	Provider     string
	RepoIdentity string
	RepoURL      string
	RepoURLs     []string
	GitDir       string
	RepoKey      string
}

type RepoPrepareMode struct {
	CloneIfMissing bool
	Fetch          bool
}

type llmIssueResponse struct {
	Issues []model.Finding `json:"issues"`
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		git:    gitops.NewManager(),
		locker: workspace.NewLocker(),
	}
}

func (o *Orchestrator) Run(ctx context.Context, taskID string, req uiapi.StartReviewRequest, emit func(string, any)) (uiapi.ReviewDoneEvent, error) {
	cfg, err := config.Load(req.ConfigPath)
	if err != nil {
		return uiapi.ReviewDoneEvent{}, err
	}

	workspaceDir := strings.TrimSpace(req.WorkspaceDir)
	if workspaceDir == "" {
		workspaceDir = cfg.Review.WorkspaceDir
	}
	layout, err := workspace.Prepare(workspaceDir)
	if err != nil {
		return uiapi.ReviewDoneEvent{}, err
	}

	if strings.TrimSpace(req.SourceBranch) == "" || strings.TrimSpace(req.TargetBranch) == "" {
		return uiapi.ReviewDoneEvent{}, fmt.Errorf("源分支和目标分支不能为空")
	}

	emitProgress(emit, taskID, "初始化", 5, "准备工作区与任务上下文", model.Summary{})

	prepared, err := o.PrepareRepository(ctx, req, cfg, layout, RepoPrepareMode{CloneIfMissing: false, Fetch: false})
	if err != nil {
		return uiapi.ReviewDoneEvent{}, err
	}

	unlock := o.locker.Acquire(prepared.RepoKey)
	defer unlock()

	emitProgress(emit, taskID, "同步代码", 15, "准备本地 worktree", model.Summary{})
	worktreePath, sourceCommit, err := o.git.PrepareWorktree(ctx, prepared.GitDir, req.SourceBranch, layout.Worktrees, prepared.RepoKey)
	if err != nil {
		return uiapi.ReviewDoneEvent{}, err
	}

	mergeBase, err := o.git.MergeBase(ctx, prepared.GitDir, req.TargetBranch, req.SourceBranch)
	if err != nil {
		return uiapi.ReviewDoneEvent{}, err
	}

	diff, err := o.git.BuildDiff(ctx, worktreePath, mergeBase, sourceCommit, cfg.Review.MaxChangedFiles)
	if err != nil {
		return uiapi.ReviewDoneEvent{}, err
	}

	emitProgress(emit, taskID, "项目画像", 30, "构建项目背景、模块与风险热点画像", model.Summary{})
	brief, _ := projectctx.NewBuilder().Build(worktreePath, diff, cfg.Rules.CustomRuleFile)

	emitProgress(emit, taskID, "规则预扫", 45, "执行确定性规则预扫", model.Summary{})
	scanRes := scanner.New().Run(diff)
	preSummary := findings.BuildSummary(scanRes.Findings)
	emitProgress(emit, taskID, "规则预扫", 52, "规则预扫完成", preSummary)

	emitProgress(emit, taskID, "构建审计包", 60, "根据 diff 与上下文构建 Review Packs", preSummary)
	packs := packer.New(cfg.Runtime.SafeInputTokens).Build(diff, brief, scanRes.Hints)

	notes := []string{
		fmt.Sprintf("审计范围：%d 个变更文件，基于 merge-base 差异语义。", len(diff.Files)),
	}
	if len(packs) == 0 {
		notes = append(notes, "未生成可供模型审计的 Review Pack，可能是本次差异为空或仅包含二进制/无文本内容文件。")
	}

	llmClient := llm.New(cfg)
	llmFindings := []model.Finding{}
	if llmClient.Enabled() && len(packs) > 0 {
		emitProgress(emit, taskID, "AI审计", 68, "开始调用模型进行分片审计", preSummary)
		llmFindings, notes = o.runLLMReview(ctx, taskID, emit, llmClient, packs, notes, cfg.Runtime.Concurrency, preSummary)
	} else {
		notes = append(notes, "模型未启用或没有可审计的 pack，本次报告主要基于规则预扫与项目画像生成。")
	}

	emitProgress(emit, taskID, "结果裁决", 88, "合并规则预扫与 AI 审计结果", preSummary)
	allFindings := findings.Normalize(append(scanRes.Findings, llmFindings...))
	summary := findings.BuildSummary(allFindings)
	health := findings.BuildHealth(allFindings)

	previous, _ := history.FindLatest(layout.Reports, prepared.RepoIdentity, prepared.RepoURL, req.SourceBranch, req.TargetBranch, taskID)
	comparison := history.Compare(previous, allFindings)
	if previous != nil {
		notes = append(notes, fmt.Sprintf("已对比上一份同分支报告：新增 %d、已修复 %d、仍存在 %d。",
			len(comparison.Added), len(comparison.Fixed), len(comparison.Existing)))
	}

	rpt := model.Report{
		TaskID:    taskID,
		Title:     fmt.Sprintf("AI代码监视报告 - %s -> %s", req.SourceBranch, req.TargetBranch),
		CreatedAt: time.Now().Format(time.RFC3339),
		Scope: model.AuditScope{
			Provider:     prepared.Provider,
			RepoIdentity: prepared.RepoIdentity,
			RepoURL:      prepared.RepoURL,
			SourceBranch: req.SourceBranch,
			TargetBranch: req.TargetBranch,
			MergeBase:    mergeBase,
			SourceCommit: sourceCommit,
			ChangedFiles: len(diff.Files),
		},
		ProjectBrief: brief,
		Findings:     allFindings,
		Summary:      summary,
		Health:       health,
		Notes:        notes,
		Comparison:   comparison,
		CodeBrowseBaseURL: strings.TrimRight(strings.TrimSpace(cfg.Review.CodeBrowseBaseURL), "/"),
		ArtifactsHint: []string{
			"artifacts/diff.json",
			"artifacts/project_brief.json",
			"artifacts/prescan_findings.json",
			"artifacts/review_packs.json",
		},
	}

	emitProgress(emit, taskID, "生成报告", 95, "生成 HTML / Markdown / JSON 报告", summary)
	reportDir := filepath.Join(layout.Reports, taskID)
	paths, err := report.SaveAll(rpt, reportDir, diff, packs, scanRes.Findings)
	if err != nil {
		return uiapi.ReviewDoneEvent{}, err
	}

	emitProgress(emit, taskID, "完成", 100, "审计完成", summary)
	return uiapi.ReviewDoneEvent{
		TaskID:       taskID,
		ReportDir:    reportDir,
		HTMLPath:     paths.HTML,
		MarkdownPath: paths.Markdown,
		JSONPath:     paths.JSON,
		Report:       rpt,
	}, nil
}

func (o *Orchestrator) PrepareRepository(
	ctx context.Context,
	req uiapi.StartReviewRequest,
	cfg config.Config,
	layout workspace.Layout,
	mode RepoPrepareMode,
) (PreparedRepo, error) {
	repoInfo := provider.Parse(req.MRURL, cfg.Git)
	providerName := repoInfo.Provider
	if providerName == "" {
		providerName = "generic"
	}

	manualRepoURL := strings.TrimSpace(req.RepoURL)
	repoURLs := uniqueStrings(compactStrings(append([]string{manualRepoURL}, repoInfo.RepoURLs...)))
	repoURL := firstNonEmpty(manualRepoURL, repoInfo.RepoURL)

	repoIdentity := strings.TrimSpace(req.LocalRepoPath)
	if repoIdentity == "" && strings.TrimSpace(repoInfo.Path) != "" {
		repoIdentity = providerName + ":" + strings.TrimSpace(repoInfo.Path)
	}
	if repoIdentity == "" {
		repoIdentity = repoURL
	}
	if repoIdentity == "" {
		return PreparedRepo{}, fmt.Errorf("无法识别仓库地址，请手动填写仓库地址或本地仓库路径")
	}

	prepared := PreparedRepo{
		Provider:     providerName,
		RepoIdentity: repoIdentity,
		RepoURL:      firstNonEmpty(repoURL, strings.TrimSpace(req.LocalRepoPath)),
		RepoURLs:     repoURLs,
		RepoKey:      workspace.RepoKey(repoIdentity),
	}

	if strings.TrimSpace(req.LocalRepoPath) != "" {
		gitDir, err := o.git.ResolveGitDir(ctx, req.LocalRepoPath)
		if err != nil {
			return PreparedRepo{}, err
		}
		prepared.GitDir = gitDir
		prepared.RepoURL = strings.TrimSpace(req.LocalRepoPath)
		if mode.Fetch {
			_ = o.git.FetchAll(ctx, gitDir)
		}
		return prepared, nil
	}

	if len(repoURLs) == 0 {
		return PreparedRepo{}, fmt.Errorf("无法识别仓库地址，请手动填写仓库地址或本地仓库路径")
	}

	gitDir, err := o.git.OpenCachedRepo(layout.Repos, repoIdentity)
	if err != nil {
		if !mode.CloneIfMissing {
			return PreparedRepo{}, err
		}
		var usedURL string
		gitDir, usedURL, err = o.git.PrepareRemoteRepo(ctx, repoURLs, layout.Repos, repoIdentity)
		if err != nil {
			return PreparedRepo{}, err
		}
		prepared.GitDir = gitDir
		prepared.RepoURL = firstNonEmpty(usedURL, prepared.RepoURL)
		return prepared, nil
	}

	prepared.GitDir = gitDir
	if mode.Fetch {
		usedURL, err := o.git.FetchWithFallback(ctx, gitDir, repoURLs)
		if err != nil {
			return PreparedRepo{}, err
		}
		prepared.RepoURL = firstNonEmpty(usedURL, prepared.RepoURL)
	}
	return prepared, nil
}

func (o *Orchestrator) SuggestBranches(ctx context.Context, gitDir string) (string, string, []string, error) {
	return o.git.SuggestBranches(ctx, gitDir)
}

func (o *Orchestrator) ValidateBranches(ctx context.Context, gitDir, source, target string) error {
	if err := o.git.BranchExists(ctx, gitDir, source); err != nil {
		return err
	}
	if err := o.git.BranchExists(ctx, gitDir, target); err != nil {
		return err
	}
	return nil
}

func (o *Orchestrator) runLLMReview(
	ctx context.Context,
	taskID string,
	emit func(string, any),
	client *llm.Client,
	packs []model.ReviewPack,
	notes []string,
	concurrency int,
	baseSummary model.Summary,
) ([]model.Finding, []string) {
	if concurrency <= 0 {
		concurrency = 4
	}

	var completed int32
	sem := make(chan struct{}, concurrency)
	outCh := make(chan []model.Finding, len(packs))
	errCh := make(chan error, len(packs))
	wg := sync.WaitGroup{}

	for _, pack := range packs {
		pack := sanitizePack(pack)
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			resp := llmIssueResponse{}
			if err := client.ChatJSON(ctx, CodeReviewSystem(), BuildCodeReviewUser(pack), 2200, &resp); err != nil {
				errCh <- fmt.Errorf("%s: %w", pack.FilePath, err)
				return
			}
			for i := range resp.Issues {
				if strings.TrimSpace(resp.Issues[i].File) == "" {
					resp.Issues[i].File = pack.FilePath
				}
				if resp.Issues[i].LineStart <= 0 {
					resp.Issues[i].LineStart = 1
				}
				if resp.Issues[i].LineEnd < resp.Issues[i].LineStart {
					resp.Issues[i].LineEnd = resp.Issues[i].LineStart
				}
			}
			outCh <- resp.Issues

			done := atomic.AddInt32(&completed, 1)
			percent := 68 + int(float64(done)/float64(max(1, len(packs)))*16)
			emitProgress(emit, taskID, "AI审计", percent, fmt.Sprintf("已完成 %d / %d 个审计包", done, len(packs)), baseSummary)
		}()
	}

	wg.Wait()
	close(outCh)
	close(errCh)

	results := []model.Finding{}
	for batch := range outCh {
		results = append(results, batch...)
	}

	errorsList := []string{}
	for err := range errCh {
		errorsList = append(errorsList, err.Error())
	}
	if len(errorsList) > 0 {
		notes = append(notes, fmt.Sprintf("AI 审计阶段有 %d 个 pack 调用失败，报告已保留其余成功结果。", len(errorsList)))
	}
	return results, notes
}

func (o *Orchestrator) ListHistory(workspaceDir string) ([]uiapi.HistoryItem, error) {
	layout, err := workspace.Prepare(workspaceDir)
	if err != nil {
		return nil, err
	}
	reports, err := history.List(layout.Reports)
	if err != nil {
		return nil, err
	}
	items := make([]uiapi.HistoryItem, 0, len(reports))
	for _, rpt := range reports {
		items = append(items, uiapi.HistoryItem{
			TaskID:      rpt.TaskID,
			Title:       rpt.Title,
			RepoURL:     rpt.Scope.RepoURL,
			SourceRef:   rpt.Scope.SourceBranch,
			TargetRef:   rpt.Scope.TargetBranch,
			CreatedAt:   rpt.CreatedAt,
			ReportDir:   filepath.Join(layout.Reports, rpt.TaskID),
			TotalIssues: rpt.Summary.Total,
			Summary:     rpt.Summary,
		})
	}
	return items, nil
}

func sanitizePack(pack model.ReviewPack) model.ReviewPack {
	replacer := strings.NewReplacer(
		"Authorization: Bearer ", "Authorization: Bearer [REDACTED]",
		"authorization: bearer ", "authorization: bearer [REDACTED]",
	)
	pack.DiffText = replacer.Replace(pack.DiffText)
	pack.ContextText = replacer.Replace(pack.ContextText)
	return pack
}

func emitProgress(emit func(string, any), taskID, stage string, percent int, message string, summary model.Summary) {
	emit("review:progress", uiapi.ProgressEvent{
		TaskID:  taskID,
		Stage:   stage,
		Percent: percent,
		Message: message,
		High:    summary.High,
		Medium:  summary.Medium,
		Low:     summary.Low,
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func compactStrings(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return uniqueStrings(out)
}

func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
