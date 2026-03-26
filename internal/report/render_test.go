package report

import (
	"os"
	"testing"

	"aiguard/internal/model"
)

func TestSaveAll(t *testing.T) {
	tmpDir := t.TempDir()
	rpt := model.Report{
		TaskID: "test-123",
		Summary: model.Summary{
			Total: 5,
		},
	}

	paths, err := SaveAll(rpt, tmpDir, &model.DiffSet{}, []model.ReviewPack{}, []model.Finding{})
	if err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}

	if _, err := os.Stat(paths.JSON); err != nil {
		t.Error("JSON file not created")
	}
	if _, err := os.Stat(paths.Markdown); err != nil {
		t.Error("Markdown file not created")
	}
	if _, err := os.Stat(paths.HTML); err != nil {
		t.Error("HTML file not created")
	}
}

func TestRenderMarkdown(t *testing.T) {
	rpt := model.Report{
		TaskID: "test",
		Summary: model.Summary{Total: 1},
		Findings: []model.Finding{
			{Title: "Test Issue", Severity: "高危"},
		},
	}

	md := renderMarkdown(rpt)
	if len(md) == 0 {
		t.Error("expected non-empty markdown")
	}
}
