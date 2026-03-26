package findings

import (
	"fmt"
	"sort"
	"strings"

	"aiguard/internal/model"
)

var severityOrder = map[string]int{
	"高危": 4,
	"严重": 3,
	"一般": 2,
	"建议": 1,
}

func Normalize(items []model.Finding) []model.Finding {
	merged := map[string]model.Finding{}
	for _, item := range items {
		item = normalizeFinding(item)
		key := mergeKey(item)

		existing, ok := merged[key]
		if !ok {
			merged[key] = item
			continue
		}

		if severityRank(item.Severity) > severityRank(existing.Severity) {
			existing.Severity = item.Severity
		}
		if len(item.Description) > len(existing.Description) {
			existing.Description = item.Description
		}
		if len(item.Impact) > len(existing.Impact) {
			existing.Impact = item.Impact
		}
		if len(item.Evidence) > len(existing.Evidence) {
			existing.Evidence = item.Evidence
		}
		if len(item.Recommendation) > len(existing.Recommendation) {
			existing.Recommendation = item.Recommendation
		}
		merged[key] = existing
	}

	out := make([]model.Finding, 0, len(merged))
	for _, item := range merged {
		out = append(out, item)
	}

	sort.Slice(out, func(i, j int) bool {
		if severityRank(out[i].Severity) != severityRank(out[j].Severity) {
			return severityRank(out[i].Severity) > severityRank(out[j].Severity)
		}
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		if out[i].LineStart != out[j].LineStart {
			return out[i].LineStart < out[j].LineStart
		}
		return out[i].Title < out[j].Title
	})

	for i := range out {
		out[i].ID = fmt.Sprintf("F%03d", i+1)
	}
	return out
}

func BuildSummary(items []model.Finding) model.Summary {
	s := model.Summary{}
	for _, item := range items {
		switch item.Severity {
		case "高危":
			s.HighRisk++
		case "严重":
			s.Severe++
		case "一般":
			s.General++
		default:
			s.Suggestion++
		}
	}
	s.High = s.HighRisk + s.Severe
	s.Medium = s.General
	s.Low = s.Suggestion
	s.Total = len(items)
	return s
}

func BuildHealth(items []model.Finding) model.HealthScore {
	return model.HealthScore{
		Security:          scoreCategory(items, "安全"),
		Performance:       scoreCategory(items, "性能"),
		Robustness:        scoreCategory(items, "健壮性"),
		Maintainability:   scoreCategory(items, "规范"),
		FrameworkPractice: scoreCategory(items, "框架"),
	}
}

func scoreCategory(items []model.Finding, category string) int {
	score := 100
	for _, item := range items {
		if item.Category != category {
			continue
		}
		switch item.Severity {
		case "高危":
			score -= 28
		case "严重":
			score -= 18
		case "一般":
			score -= 8
		default:
			score -= 3
		}
	}
	if score < 0 {
		return 0
	}
	return score
}

func normalizeFinding(item model.Finding) model.Finding {
	item.Title = strings.TrimSpace(item.Title)
	item.Description = strings.TrimSpace(item.Description)
	item.Impact = strings.TrimSpace(item.Impact)
	item.Evidence = strings.TrimSpace(item.Evidence)
	item.Recommendation = strings.TrimSpace(item.Recommendation)
	item.File = strings.TrimSpace(item.File)
	item.Category = normalizeCategory(item.Category)
	item.Severity = normalizeSeverity(item.Severity)
	if item.Confidence == "" {
		item.Confidence = "medium"
	}
	if item.LineStart <= 0 {
		item.LineStart = 1
	}
	if item.LineEnd < item.LineStart {
		item.LineEnd = item.LineStart
	}
	if item.Title == "" {
		item.Title = "未命名问题"
	}
	if item.Category == "" {
		item.Category = "规范"
	}
	if item.Severity == "" {
		item.Severity = "建议"
	}
	return item
}

func normalizeSeverity(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	switch raw {
	case "critical", "blocker", "高危":
		return "高危"
	case "high", "severe", "严重":
		return "严重"
	case "medium", "warning", "一般", "中等":
		return "一般"
	case "low", "advice", "建议":
		return "建议"
	default:
		return "建议"
	}
}

func normalizeCategory(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	switch raw {
	case "security", "安全":
		return "安全"
	case "performance", "性能":
		return "性能"
	case "robustness", "reliability", "健壮性":
		return "健壮性"
	case "framework", "框架":
		return "框架"
	default:
		return "规范"
	}
}

func mergeKey(item model.Finding) string {
	return fmt.Sprintf("%s|%s|%d|%d|%s",
		strings.ToLower(item.File),
		strings.ToLower(item.Title),
		item.LineStart,
		item.LineEnd,
		strings.ToLower(item.Category),
	)
}

func severityRank(value string) int {
	if v, ok := severityOrder[value]; ok {
		return v
	}
	return 0
}
