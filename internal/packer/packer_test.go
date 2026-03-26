package packer

import (
	"testing"

	"aiguard/internal/model"
)

func TestNew(t *testing.T) {
	p := New(100000)
	if p == nil {
		t.Error("expected non-nil builder")
	}
}

func TestBuild(t *testing.T) {
	p := New(100000)
	diff := &model.DiffSet{
		Files: []model.ChangedFile{
			{
				Path:          "file1.go",
				SourceContent: "package main\nfunc main() {}",
				Patch:         "+package main\n+func main() {}",
			},
		},
	}
	brief := model.ProjectBrief{}
	hints := map[string][]string{}

	packs := p.Build(diff, brief, hints)
	if len(packs) == 0 {
		t.Error("expected at least one pack")
	}
}
