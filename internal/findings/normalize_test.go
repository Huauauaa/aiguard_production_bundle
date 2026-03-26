package findings

import (
	"testing"

	"aiguard/internal/model"
)

func TestNormalizeFindings(t *testing.T) {
	findings := []model.Finding{
		{File: "test.go", LineStart: 10, Title: "问题1", Severity: "严重"},
		{File: "test.go", LineStart: 10, Title: "问题1", Severity: "严重"},
		{File: "test.go", LineStart: 20, Title: "问题2", Severity: "高危"},
	}

	result := Normalize(findings)
	if len(result) != 2 {
		t.Errorf("expected 2 unique findings, got %d", len(result))
	}
}

func TestNormalizeEmpty(t *testing.T) {
	result := Normalize([]model.Finding{})
	if len(result) != 0 {
		t.Error("expected empty result")
	}
}
