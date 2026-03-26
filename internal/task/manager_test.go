package task

import (
	"testing"

	"aiguard/internal/uiapi"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Error("expected non-nil manager")
	}
	if m.HasRunning() {
		t.Error("new manager should have no running tasks")
	}
}

func TestAddAndDone(t *testing.T) {
	m := NewManager()
	cancel := func() {}
	req := uiapi.StartReviewRequest{}

	m.Add("task1", cancel, req)
	if !m.HasRunning() {
		t.Error("expected running task")
	}

	m.Done("task1")
	if m.HasRunning() {
		t.Error("expected no running tasks")
	}
}

func TestCancel(t *testing.T) {
	m := NewManager()
	cancelled := false
	cancel := func() { cancelled = true }
	req := uiapi.StartReviewRequest{}

	m.Add("task1", cancel, req)
	err := m.Cancel("task1")
	if err != nil {
		t.Errorf("cancel failed: %v", err)
	}
	if !cancelled {
		t.Error("cancel function not called")
	}
	if m.HasRunning() {
		t.Error("task should be removed")
	}
}

func TestCancelNonExistent(t *testing.T) {
	m := NewManager()
	err := m.Cancel("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}
