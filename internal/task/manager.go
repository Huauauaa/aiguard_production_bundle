package task

import (
	"errors"
	"sync"

	"aiguard/internal/uiapi"
)

type TaskHandle struct {
	Cancel func()
	Req    uiapi.StartReviewRequest
}

type Manager struct {
	mu    sync.Mutex
	tasks map[string]*TaskHandle
}

func NewManager() *Manager {
	return &Manager{
		tasks: map[string]*TaskHandle{},
	}
}

func (m *Manager) Add(taskID string, cancel func(), req uiapi.StartReviewRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[taskID] = &TaskHandle{
		Cancel: cancel,
		Req:    req,
	}
}

func (m *Manager) Done(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tasks, taskID)
}

func (m *Manager) Cancel(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	handle, ok := m.tasks[taskID]
	if !ok {
		return errors.New("任务不存在或已结束")
	}

	handle.Cancel()
	delete(m.tasks, taskID)
	return nil
}
