package packer

import (
	"fmt"
	"strings"

	"aiguard/internal/model"
)

type Builder struct {
	maxChars int
}

func New(safeInputTokens int) *Builder {
	maxChars := 14000
	if safeInputTokens > 0 {
		candidate := safeInputTokens / 8
		if candidate < maxChars {
			maxChars = candidate
		}
		if maxChars < 6000 {
			maxChars = 6000
		}
	}
	return &Builder{maxChars: maxChars}
}

func (b *Builder) Build(diff *model.DiffSet, brief model.ProjectBrief, hints map[string][]string) []model.ReviewPack {
	packs := []model.ReviewPack{}
	briefText := brief.SummaryText()

	for _, file := range diff.Files {
		if strings.TrimSpace(file.Patch) == "" {
			continue
		}

		starts := file.HunkNewStarts
		if len(starts) == 0 {
			starts = []int{1}
		}

		windows := mergeWindows(starts, 40, 120)
		if len(windows) == 0 {
			windows = [][2]int{{1, 160}}
		}

		for idx, window := range windows {
			contextText := extractLines(file.SourceContent, window[0], window[1])
			pack := model.ReviewPack{
				ID:           fmt.Sprintf("%s#%d", file.Path, idx+1),
				FilePath:     file.Path,
				Language:     file.Language,
				DiffText:     truncate(file.Patch, b.maxChars/2),
				ContextText:  truncate(contextText, b.maxChars/2),
				ProjectBrief: truncate(briefText, 2500),
				PreScanHints: unique(hints[file.Path]),
			}
			pack.TokenEstimate = estimateTokens(pack.DiffText + "\n" + pack.ContextText + "\n" + pack.ProjectBrief)
			packs = append(packs, pack)
		}
	}

	return packs
}

func mergeWindows(starts []int, before, after int) [][2]int {
	merged := [][2]int{}
	for _, start := range starts {
		left := max(1, start-before)
		right := start + after
		if len(merged) == 0 {
			merged = append(merged, [2]int{left, right})
			continue
		}
		last := &merged[len(merged)-1]
		if left <= last[1]+20 {
			if right > last[1] {
				last[1] = right
			}
			continue
		}
		merged = append(merged, [2]int{left, right})
	}
	if len(merged) > 4 {
		merged = merged[:4]
	}
	return merged
}

func extractLines(content string, start, end int) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	lines := strings.Split(content, "\n")
	if start < 1 {
		start = 1
	}
	if end > len(lines) {
		end = len(lines)
	}
	if start > end || start > len(lines) {
		return ""
	}
	selected := lines[start-1 : end]
	var b strings.Builder
	for i, line := range selected {
		fmt.Fprintf(&b, "%d| %s\n", start+i, line)
	}
	return b.String()
}

func truncate(s string, maxChars int) string {
	if maxChars <= 0 || len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "\n...[truncated]"
}

func estimateTokens(s string) int {
	if s == "" {
		return 0
	}
	return len(s) / 4
}

func unique(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
