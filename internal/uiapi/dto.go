package uiapi

import "aiguard/internal/model"

type StartReviewRequest struct {
	MRURL         string `json:"mrUrl"`
	RepoURL       string `json:"repoUrl"`
	LocalRepoPath string `json:"localRepoPath"`
	SourceBranch  string `json:"sourceBranch"`
	TargetBranch  string `json:"targetBranch"`
	ConfigPath    string `json:"configPath"`
	WorkspaceDir  string `json:"workspaceDir"`
}

type ProgressEvent struct {
	TaskID  string `json:"taskId"`
	Stage   string `json:"stage"`
	Percent int    `json:"percent"`
	Message string `json:"message"`
	High    int    `json:"high"`
	Medium  int    `json:"medium"`
	Low     int    `json:"low"`
}

type ReviewDoneEvent struct {
	TaskID       string       `json:"taskId"`
	ReportDir    string       `json:"reportDir"`
	HTMLPath     string       `json:"htmlPath"`
	MarkdownPath string       `json:"markdownPath"`
	JSONPath     string       `json:"jsonPath"`
	Report       model.Report `json:"report"`
}

type HistoryItem struct {
	TaskID      string        `json:"taskId"`
	Title       string        `json:"title"`
	RepoURL     string        `json:"repoUrl"`
	SourceRef   string        `json:"sourceRef"`
	TargetRef   string        `json:"targetRef"`
	CreatedAt   string        `json:"createdAt"`
	ReportDir   string        `json:"reportDir"`
	TotalIssues int           `json:"totalIssues"`
	Summary     model.Summary `json:"summary"`
}
