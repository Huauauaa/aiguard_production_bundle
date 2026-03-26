package errors

import "errors"

var (
	ErrRepoNotFound     = errors.New("仓库未找到")
	ErrBranchInvalid    = errors.New("分支无效")
	ErrBranchEmpty      = errors.New("源分支和目标分支不能为空")
	ErrRepoURLUnknown   = errors.New("无法识别仓库地址")
	ErrLLMUnavailable   = errors.New("LLM 服务不可用")
	ErrConfigInvalid    = errors.New("配置无效")
	ErrWorkspaceInvalid = errors.New("工作区无效")
)
