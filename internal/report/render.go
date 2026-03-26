package report

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"aiguard/internal/model"
)

//go:embed templates/report.html.tmpl
var templateFS embed.FS

type Paths struct {
	JSON     string
	Markdown string
	HTML     string
}

func SaveAll(rpt model.Report, dir string, diff *model.DiffSet, packs []model.ReviewPack, prescan []model.Finding) (Paths, error) {
	if err := os.MkdirAll(filepath.Join(dir, "artifacts"), 0o755); err != nil {
		return Paths{}, err
	}

	paths := Paths{
		JSON:     filepath.Join(dir, "report.json"),
		Markdown: filepath.Join(dir, "report.md"),
		HTML:     filepath.Join(dir, "report.html"),
	}

	if err := writeJSON(paths.JSON, rpt); err != nil {
		return Paths{}, err
	}
	if err := os.WriteFile(paths.Markdown, []byte(renderMarkdown(rpt)), 0o644); err != nil {
		return Paths{}, err
	}
	htmlContent, err := renderHTML(rpt)
	if err != nil {
		return Paths{}, err
	}
	if err := os.WriteFile(paths.HTML, []byte(htmlContent), 0o644); err != nil {
		return Paths{}, err
	}

	_ = writeJSON(filepath.Join(dir, "artifacts", "diff.json"), diff)
	_ = writeJSON(filepath.Join(dir, "artifacts", "project_brief.json"), rpt.ProjectBrief)
	_ = writeJSON(filepath.Join(dir, "artifacts", "review_packs.json"), packs)
	_ = writeJSON(filepath.Join(dir, "artifacts", "prescan_findings.json"), prescan)

	return paths, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func renderMarkdown(rpt model.Report) string {
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(rpt.Title)
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("- 生成时间：%s\n", rpt.CreatedAt))
	b.WriteString(fmt.Sprintf("- 仓库：%s\n", rpt.Scope.RepoURL))
	b.WriteString(fmt.Sprintf("- 分支：%s -> %s\n", rpt.Scope.SourceBranch, rpt.Scope.TargetBranch))
	b.WriteString(fmt.Sprintf("- 变更文件数：%d\n\n", rpt.Scope.ChangedFiles))

	b.WriteString("## 1. 问题清单\n\n")
	for i, item := range rpt.Findings {
		b.WriteString(fmt.Sprintf("### %03d：%s（%s）\n", i+1, item.Title, item.Severity))
		b.WriteString(fmt.Sprintf("- 文件：`%s`\n", item.File))
		b.WriteString(fmt.Sprintf("- 行号：%d-%d\n", item.LineStart, item.LineEnd))
		b.WriteString(fmt.Sprintf("- 分类：%s\n", item.Category))
		b.WriteString(fmt.Sprintf("- 详细描述：%s\n", item.Description))
		b.WriteString(fmt.Sprintf("- 影响分析：%s\n", item.Impact))
		b.WriteString(fmt.Sprintf("- 修复建议：%s\n\n", item.Recommendation))
	}

	b.WriteString("## 2. 整体统计\n\n")
	b.WriteString(fmt.Sprintf("- 高（高危+严重）：%d\n", rpt.Summary.High))
	b.WriteString(fmt.Sprintf("- 中（一般）：%d\n", rpt.Summary.Medium))
	b.WriteString(fmt.Sprintf("- 低（建议）：%d\n", rpt.Summary.Low))
	b.WriteString(fmt.Sprintf("- 总计：%d\n\n", rpt.Summary.Total))

	b.WriteString("## 3. 代码质量与健康度分析\n\n")
	b.WriteString(fmt.Sprintf("- 安全性：%d\n", rpt.Health.Security))
	b.WriteString(fmt.Sprintf("- 性能：%d\n", rpt.Health.Performance))
	b.WriteString(fmt.Sprintf("- 健壮性：%d\n", rpt.Health.Robustness))
	b.WriteString(fmt.Sprintf("- 可维护性：%d\n", rpt.Health.Maintainability))
	b.WriteString(fmt.Sprintf("- 框架最佳实践：%d\n\n", rpt.Health.FrameworkPractice))

	b.WriteString("## 4. 其他待补充\n\n")
	for _, note := range rpt.Notes {
		b.WriteString("- ")
		b.WriteString(note)
		b.WriteString("\n")
	}
	return b.String()
}

func renderHTML(rpt model.Report) (string, error) {
	data, err := templateFS.ReadFile("templates/report.html.tmpl")
	if err != nil {
		return "", err
	}
	tpl, err := template.New("report").Parse(string(data))
	if err != nil {
		return "", err
	}
	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, rpt); err != nil {
		return "", err
	}
	return buf.String(), nil
}
