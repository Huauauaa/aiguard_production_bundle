package scanner

import (
	"testing"

	"aiguard/internal/model"
)

func TestScanSecrets(t *testing.T) {
	s := New()
	diff := &model.DiffSet{
		Files: []model.ChangedFile{
			{
				Path:          "config.go",
				SourceContent: `api_key = "sk-1234567890abcdef"`,
			},
		},
	}

	result := s.Run(diff)
	if len(result.Findings) == 0 {
		t.Error("expected to find hardcoded secret")
	}
	if result.Findings[0].Title != "疑似硬编码敏感信息" {
		t.Errorf("unexpected finding title: %s", result.Findings[0].Title)
	}
}

func TestScanSQLConcat(t *testing.T) {
	s := New()
	diff := &model.DiffSet{
		Files: []model.ChangedFile{
			{
				Path:          "db.go",
				SourceContent: `query := "SELECT * FROM users WHERE id=" + userId`,
			},
		},
	}

	result := s.Run(diff)
	if len(result.Findings) == 0 {
		t.Error("expected to find SQL concatenation")
	}
}

func TestScanCommandExec(t *testing.T) {
	s := New()
	diff := &model.DiffSet{
		Files: []model.ChangedFile{
			{
				Path:          "exec.go",
				SourceContent: `cmd := exec.Command("sh", "-c", userInput)`,
			},
		},
	}

	result := s.Run(diff)
	if len(result.Findings) == 0 {
		t.Error("expected to find command execution risk")
	}
}
