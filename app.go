package main

import (
	"context"

	"aiguard/internal/review"
	"aiguard/internal/task"
	"aiguard/internal/uiapi"
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

func (a *App) StartReview(req uiapi.StartReviewRequest) (string, error) {
	taskID := uuid.NewString()
	runCtx, cancel := context.WithCancel(context.Background())
	a.tasks.Add(taskID, cancel, req)

	go func() {
		defer a.tasks.Done(taskID)

		emit := func(name string, payload any) {
			runtime.EventsEmit(a.ctx, name, payload)
		}

		done, err := a.orch.Run(runCtx, taskID, req, emit)
		if err != nil {
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

func (a *App) ListHistory() ([]uiapi.HistoryItem, error) {
	return a.orch.ListHistory()
}
