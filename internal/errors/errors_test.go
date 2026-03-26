package errors

import (
	"errors"
	"testing"
)

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrRepoNotFound", ErrRepoNotFound, "仓库未找到"},
		{"ErrBranchInvalid", ErrBranchInvalid, "分支无效"},
		{"ErrBranchEmpty", ErrBranchEmpty, "源分支和目标分支不能为空"},
		{"ErrRepoURLUnknown", ErrRepoURLUnknown, "无法识别仓库地址"},
		{"ErrLLMUnavailable", ErrLLMUnavailable, "LLM 服务不可用"},
		{"ErrConfigInvalid", ErrConfigInvalid, "配置无效"},
		{"ErrWorkspaceInvalid", ErrWorkspaceInvalid, "工作区无效"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	if !errors.Is(ErrRepoNotFound, ErrRepoNotFound) {
		t.Error("errors.Is should work with custom errors")
	}
}
