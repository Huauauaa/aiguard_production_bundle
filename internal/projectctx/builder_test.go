package projectctx

import (
	"os"
	"path/filepath"
	"testing"

	"aiguard/internal/model"
)

func TestDetectTechStack(t *testing.T) {
	tmpDir := t.TempDir()
	builder := NewBuilder()

	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644)

	stack := builder.detectTechStack(tmpDir)
	if len(stack) != 2 {
		t.Errorf("expected 2 tech stacks, got %d", len(stack))
	}
}

func TestDetectFrameworks(t *testing.T) {
	tmpDir := t.TempDir()
	builder := NewBuilder()

	packageJSON := `{"dependencies": {"react": "^18.0.0", "vue": "^3.0.0"}}`
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644)

	frameworks := builder.detectFrameworks(tmpDir)
	if len(frameworks) < 2 {
		t.Errorf("expected at least 2 frameworks, got %d", len(frameworks))
	}
}

func TestBuildWithEmptyRepo(t *testing.T) {
	tmpDir := t.TempDir()
	builder := NewBuilder()
	diff := &model.DiffSet{}

	brief, err := builder.Build(tmpDir, diff, "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if brief.Purpose == "" {
		t.Error("expected non-empty purpose")
	}
}
