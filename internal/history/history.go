package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aiguard/internal/model"
)

func List(reportsRoot string) ([]model.Report, error) {
	entries, err := os.ReadDir(reportsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	reports := []model.Report{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(reportsRoot, entry.Name(), "report.json")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var rpt model.Report
		if err := json.Unmarshal(data, &rpt); err != nil {
			continue
		}
		reports = append(reports, rpt)
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].CreatedAt > reports[j].CreatedAt
	})
	return reports, nil
}

func FindLatest(reportsRoot, repoURL, source, target, exceptTaskID string) (*model.Report, error) {
	reports, err := List(reportsRoot)
	if err != nil {
		return nil, err
	}
	for _, rpt := range reports {
		if rpt.TaskID == exceptTaskID {
			continue
		}
		if strings.TrimSpace(repoURL) != "" && strings.TrimSpace(rpt.Scope.RepoURL) != strings.TrimSpace(repoURL) {
			continue
		}
		if strings.TrimSpace(source) != "" && strings.TrimSpace(rpt.Scope.SourceBranch) != strings.TrimSpace(source) {
			continue
		}
		if strings.TrimSpace(target) != "" && strings.TrimSpace(rpt.Scope.TargetBranch) != strings.TrimSpace(target) {
			continue
		}
		copy := rpt
		return &copy, nil
	}
	return nil, nil
}

func Compare(previous *model.Report, current []model.Finding) model.ComparisonResult {
	if previous == nil {
		added := []string{}
		for _, finding := range current {
			added = append(added, finding.Title+" @ "+finding.File)
		}
		return model.ComparisonResult{Added: unique(added)}
	}

	prevSet := map[string]string{}
	for _, finding := range previous.Findings {
		prevSet[finding.Fingerprint()] = finding.Title + " @ " + finding.File
	}

	currSet := map[string]string{}
	for _, finding := range current {
		currSet[finding.Fingerprint()] = finding.Title + " @ " + finding.File
	}

	cmp := model.ComparisonResult{}
	for fp, desc := range currSet {
		if _, ok := prevSet[fp]; ok {
			cmp.Existing = append(cmp.Existing, desc)
		} else {
			cmp.Added = append(cmp.Added, desc)
		}
	}
	for fp, desc := range prevSet {
		if _, ok := currSet[fp]; !ok {
			cmp.Fixed = append(cmp.Fixed, desc)
		}
	}

	cmp.Added = unique(cmp.Added)
	cmp.Fixed = unique(cmp.Fixed)
	cmp.Existing = unique(cmp.Existing)
	sort.Strings(cmp.Added)
	sort.Strings(cmp.Fixed)
	sort.Strings(cmp.Existing)
	return cmp
}

func unique(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
