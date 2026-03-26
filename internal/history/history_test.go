package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestList(t *testing.T) {
	tmpDir := t.TempDir()

	reports, err := List(tmpDir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(reports) != 0 {
		t.Errorf("expected 0 reports, got %d", len(reports))
	}
}

func TestListWithReport(t *testing.T) {
	tmpDir := t.TempDir()
	reportDir := filepath.Join(tmpDir, "test-report")
	os.MkdirAll(reportDir, 0755)

	data := []byte(`{"taskId":"test-123","scope":{"sourceBranch":"feature","targetBranch":"main"}}`)
	os.WriteFile(filepath.Join(reportDir, "report.json"), data, 0644)

	reports, err := List(tmpDir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(reports) != 1 {
		t.Errorf("expected 1 report, got %d", len(reports))
	}
}
