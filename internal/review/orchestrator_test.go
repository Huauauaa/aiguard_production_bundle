package review

import (
	"testing"

	"aiguard/internal/model"
)

func TestNewOrchestrator(t *testing.T) {
	orch := NewOrchestrator()
	if orch == nil {
		t.Error("expected non-nil orchestrator")
	}
	if orch.git == nil {
		t.Error("expected non-nil git manager")
	}
	if orch.locker == nil {
		t.Error("expected non-nil locker")
	}
}

func TestEmitProgress(t *testing.T) {
	called := false
	emit := func(name string, payload any) {
		called = true
	}
	emitProgress(emit, "task1", "测试", 50, "测试消息", model.Summary{})
	if !called {
		t.Error("emit function not called")
	}
}
