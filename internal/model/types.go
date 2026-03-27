package model

import (
	"fmt"
	"strings"
)

type AuditScope struct {
	Provider     string `json:"provider"`
	RepoIdentity string `json:"repoIdentity,omitempty"`
	RepoURL      string `json:"repoUrl"`
	SourceBranch string `json:"sourceBranch"`
	TargetBranch string `json:"targetBranch"`
	MergeBase    string `json:"mergeBase"`
	SourceCommit string `json:"sourceCommit"`
	ChangedFiles int    `json:"changedFiles"`
}

type ProjectBrief struct {
	Purpose             string   `json:"purpose"`
	TechStack           []string `json:"techStack"`
	Frameworks          []string `json:"frameworks"`
	Modules             []string `json:"modules"`
	Entrypoints         []string `json:"entrypoints"`
	SensitiveAreas      []string `json:"sensitiveAreas"`
	PerformanceHotspots []string `json:"performanceHotspots"`
	ChangeImpactPath    []string `json:"changeImpactPath"`
	RepoRules           []string `json:"repoRules"`
	Notes               []string `json:"notes"`
}

func (p ProjectBrief) SummaryText() string {
	var b strings.Builder
	writeList := func(title string, items []string) {
		if len(items) == 0 {
			return
		}
		b.WriteString(title)
		b.WriteString(": ")
		b.WriteString(strings.Join(items, ", "))
		b.WriteString("\n")
	}

	if strings.TrimSpace(p.Purpose) != "" {
		b.WriteString("项目用途: ")
		b.WriteString(p.Purpose)
		b.WriteString("\n")
	}
	writeList("技术栈", p.TechStack)
	writeList("框架", p.Frameworks)
	writeList("模块", p.Modules)
	writeList("入口", p.Entrypoints)
	writeList("高敏感区域", p.SensitiveAreas)
	writeList("性能热点", p.PerformanceHotspots)
	writeList("本次改动影响链路", p.ChangeImpactPath)
	writeList("仓库规则", p.RepoRules)
	writeList("补充说明", p.Notes)
	return strings.TrimSpace(b.String())
}

type ChangedFile struct {
	Path          string `json:"path"`
	Status        string `json:"status"`
	OldPath       string `json:"oldPath,omitempty"`
	Language      string `json:"language"`
	Patch         string `json:"patch"`
	SourceContent string `json:"sourceContent"`
	BaseContent   string `json:"baseContent,omitempty"`
	HunkNewStarts []int  `json:"hunkNewStarts,omitempty"`
}

type DiffSet struct {
	MergeBase    string        `json:"mergeBase"`
	SourceCommit string        `json:"sourceCommit"`
	Files        []ChangedFile `json:"files"`
}

type ReviewPack struct {
	ID            string   `json:"id"`
	FilePath      string   `json:"filePath"`
	Language      string   `json:"language"`
	DiffText      string   `json:"diffText"`
	ContextText   string   `json:"contextText"`
	ProjectBrief  string   `json:"projectBrief"`
	PreScanHints  []string `json:"preScanHints"`
	TokenEstimate int      `json:"tokenEstimate"`
}

type Finding struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Severity       string `json:"severity"`
	Category       string `json:"category"`
	Confidence     string `json:"confidence"`
	File           string `json:"file"`
	LineStart      int    `json:"lineStart"`
	LineEnd        int    `json:"lineEnd"`
	Description    string `json:"description"`
	Impact         string `json:"impact"`
	Evidence       string `json:"evidence"`
	Recommendation string `json:"recommendation"`
}

func (f Finding) Fingerprint() string {
	return fmt.Sprintf("%s|%s|%d|%d|%s",
		strings.ToLower(strings.TrimSpace(f.File)),
		strings.ToLower(strings.TrimSpace(f.Title)),
		f.LineStart,
		f.LineEnd,
		strings.ToLower(strings.TrimSpace(f.Category)),
	)
}

type Summary struct {
	HighRisk   int `json:"highRisk"`
	Severe     int `json:"severe"`
	General    int `json:"general"`
	Suggestion int `json:"suggestion"`

	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
	Total  int `json:"total"`
}

type HealthScore struct {
	Security          int `json:"security"`
	Performance       int `json:"performance"`
	Robustness        int `json:"robustness"`
	Maintainability   int `json:"maintainability"`
	FrameworkPractice int `json:"frameworkPractice"`
}

type ComparisonResult struct {
	Added    []string `json:"added"`
	Fixed    []string `json:"fixed"`
	Existing []string `json:"existing"`
}

type Report struct {
	TaskID        string           `json:"taskId"`
	Title         string           `json:"title"`
	CreatedAt     string           `json:"createdAt"`
	Scope         AuditScope       `json:"scope"`
	ProjectBrief  ProjectBrief     `json:"projectBrief"`
	Findings      []Finding        `json:"findings"`
	Summary       Summary          `json:"summary"`
	Health        HealthScore      `json:"health"`
	Notes         []string         `json:"notes"`
	Comparison    ComparisonResult `json:"comparison"`
	SkippedFiles  []string         `json:"skippedFiles,omitempty"`
	ArtifactsHint []string         `json:"artifactsHint,omitempty"`
	CodeBrowseBaseURL string       `json:"codeBrowseBaseUrl,omitempty"`
}
