package gitops

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"aiguard/internal/model"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) EnsureBareRepo(ctx context.Context, repoURL, reposDir string) (string, error) {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return "", errors.New("仓库地址为空")
	}
	gitDir, _, err := m.PrepareRemoteRepo(ctx, []string{repoURL}, reposDir, repoURL)
	return gitDir, err
}

func (m *Manager) OpenCachedRepo(reposDir, repoIdentity string) (string, error) {
	repoIdentity = strings.TrimSpace(repoIdentity)
	if repoIdentity == "" {
		return "", errors.New("仓库标识为空")
	}
	target := filepath.Join(reposDir, repoKey(repoIdentity)+".git")
	if _, err := os.Stat(target); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("本地缓存中未找到仓库，请先点击“拉取代码”")
		}
		return "", err
	}
	return target, nil
}

func validateRepoURL(repoURL string) error {
	if strings.Contains(repoURL, ";") || strings.Contains(repoURL, "|") || strings.Contains(repoURL, "&") {
		return errors.New("仓库地址包含非法字符")
	}
	return nil
}

func (m *Manager) PrepareRemoteRepo(ctx context.Context, repoURLs []string, reposDir, repoIdentity string) (string, string, error) {
	repoIdentity = strings.TrimSpace(repoIdentity)
	if repoIdentity == "" {
		return "", "", errors.New("仓库标识为空")
	}
	candidates := compactStrings(repoURLs)
	if len(candidates) == 0 {
		return "", "", errors.New("仓库地址为空")
	}
	for _, url := range candidates {
		if err := validateRepoURL(url); err != nil {
			return "", "", fmt.Errorf("仓库地址校验失败: %w", err)
		}
	}

	target := filepath.Join(reposDir, repoKey(repoIdentity)+".git")
	if _, err := os.Stat(target); err == nil {
		return target, "", nil
	}
	if err := os.MkdirAll(reposDir, 0o755); err != nil {
		return "", "", err
	}

	failures := []string{}
	for _, repoURL := range candidates {
		_ = os.RemoveAll(target)
		cmd := commandContext(ctx, "git", "clone", "--bare", repoURL, target)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			failures = append(failures, fmt.Sprintf("%s => %v: %s", repoURL, err, strings.TrimSpace(stderr.String())))
			continue
		}
		return target, repoURL, nil
	}

	_ = os.RemoveAll(target)
	return "", "", fmt.Errorf("clone 失败，已尝试所有仓库地址: %s", strings.Join(uniqueStrings(failures), " | "))
}

func (m *Manager) ResolveGitDir(ctx context.Context, repoPath string) (string, error) {
	out, err := m.run(ctx, repoPath, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (m *Manager) FetchAll(ctx context.Context, gitDir string) error {
	_, err := m.runGitDir(ctx, gitDir, "fetch", "--all", "--prune")
	return err
}

func (m *Manager) FetchWithFallback(ctx context.Context, gitDir string, repoURLs []string) (string, error) {
	currentURL, _ := m.currentOriginURL(ctx, gitDir)
	candidates := compactStrings(append([]string{currentURL}, repoURLs...))
	if len(candidates) == 0 {
		if err := m.FetchAll(ctx, gitDir); err != nil {
			return "", err
		}
		return "", nil
	}

	var failures []string
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		if currentURL != candidate {
			if _, err := m.runGitDir(ctx, gitDir, "remote", "set-url", "origin", candidate); err != nil {
				failures = append(failures, err.Error())
				continue
			}
			currentURL = candidate
		}
		if err := m.FetchAll(ctx, gitDir); err == nil {
			return candidate, nil
		} else {
			failures = append(failures, err.Error())
		}
	}

	if len(failures) == 0 {
		return "", errors.New("fetch 失败")
	}
	return "", fmt.Errorf("fetch 失败，已尝试所有仓库地址: %s", strings.Join(uniqueStrings(failures), " | "))
}

func (m *Manager) currentOriginURL(ctx context.Context, gitDir string) (string, error) {
	out, err := m.runGitDir(ctx, gitDir, "remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (m *Manager) ResolveCommit(ctx context.Context, gitDir, ref string) (string, error) {
	for _, candidate := range candidateRefs(ref) {
		out, err := m.runGitDir(ctx, gitDir, "rev-parse", candidate+"^{commit}")
		if err == nil {
			return strings.TrimSpace(out), nil
		}
	}
	return "", fmt.Errorf("无法解析引用: %s", ref)
}

func (m *Manager) BranchExists(ctx context.Context, gitDir, ref string) error {
	_, err := m.ResolveCommit(ctx, gitDir, ref)
	if err != nil {
		return fmt.Errorf("分支不存在: %s", ref)
	}
	return nil
}

func (m *Manager) MergeBase(ctx context.Context, gitDir, targetRef, sourceRef string) (string, error) {
	target, err := m.ResolveCommit(ctx, gitDir, targetRef)
	if err != nil {
		return "", err
	}
	source, err := m.ResolveCommit(ctx, gitDir, sourceRef)
	if err != nil {
		return "", err
	}
	out, err := m.runGitDir(ctx, gitDir, "merge-base", target, source)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (m *Manager) PrepareWorktree(ctx context.Context, gitDir, sourceRef, worktreesDir, repoKey string) (string, string, error) {
	sourceCommit, err := m.ResolveCommit(ctx, gitDir, sourceRef)
	if err != nil {
		return "", "", err
	}

	short := sourceCommit
	if len(short) > 12 {
		short = sourceCommit[:12]
	}
	wt := filepath.Join(worktreesDir, repoKey, short)
	_ = os.RemoveAll(wt)

	if err := os.MkdirAll(filepath.Dir(wt), 0o755); err != nil {
		return "", "", err
	}

	_, err = m.runGitDir(ctx, gitDir, "worktree", "add", "--force", "--detach", wt, sourceCommit)
	if err != nil {
		return "", "", err
	}
	return wt, sourceCommit, nil
}

func (m *Manager) BuildDiff(ctx context.Context, worktreePath, mergeBase, sourceCommit string, maxChangedFiles int) (*model.DiffSet, error) {
	nameStatus, err := m.run(ctx, worktreePath, "diff", "--name-status", mergeBase+"..."+sourceCommit)
	if err != nil {
		return nil, err
	}

	lines := splitNonEmptyLines(nameStatus)
	files := make([]model.ChangedFile, 0, len(lines))
	hunkPattern := regexp.MustCompile(`@@ -\d+(?:,\d+)? \+(\d+)`)

	for i, line := range lines {
		if maxChangedFiles > 0 && i >= maxChangedFiles {
			break
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		status := parts[0]
		path := parts[len(parts)-1]
		oldPath := ""
		if len(parts) >= 3 {
			oldPath = parts[1]
		}

		patch, err := m.run(ctx, worktreePath, "diff", "--unified=3", mergeBase+"..."+sourceCommit, "--", path)
		if err != nil {
			patch = ""
		}

		sourceContent := ""
		fullPath := filepath.Join(worktreePath, path)
		if data, err := os.ReadFile(fullPath); err == nil {
			sourceContent = string(data)
		}

		baseContent, err := m.run(ctx, worktreePath, "show", mergeBase+":"+path)
		if err != nil {
			baseContent = ""
		}

		hunkStarts := []int{}
		for _, match := range hunkPattern.FindAllStringSubmatch(patch, -1) {
			if len(match) != 2 {
				continue
			}
			var n int
			fmt.Sscanf(match[1], "%d", &n)
			if n > 0 {
				hunkStarts = append(hunkStarts, n)
			}
		}

		files = append(files, model.ChangedFile{
			Path:          path,
			OldPath:       oldPath,
			Status:        status,
			Language:      detectLanguage(path),
			Patch:         patch,
			SourceContent: sourceContent,
			BaseContent:   baseContent,
			HunkNewStarts: hunkStarts,
		})
	}

	return &model.DiffSet{
		MergeBase:    mergeBase,
		SourceCommit: sourceCommit,
		Files:        files,
	}, nil
}

func (m *Manager) ListBranchesByActivity(ctx context.Context, gitDir string) ([]string, error) {
	branches := []string{}
	seen := map[string]struct{}{}

	collect := func(scope, prefix string) {
		out, err := m.runGitDir(ctx, gitDir, "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)", scope)
		if err != nil {
			return
		}
		for _, branch := range splitNonEmptyLines(out) {
			branch = normalizeBranchName(branch, prefix)
			if branch == "" {
				continue
			}
			if _, ok := seen[branch]; ok {
				continue
			}
			seen[branch] = struct{}{}
			branches = append(branches, branch)
		}
	}

	collect("refs/remotes/origin", "origin/")
	collect("refs/heads", "")
	if len(branches) == 0 {
		return nil, errors.New("未找到任何分支")
	}
	return branches, nil
}

func (m *Manager) SuggestBranches(ctx context.Context, gitDir string) (string, string, []string, error) {
	branches, err := m.ListBranchesByActivity(ctx, gitDir)
	if err != nil {
		return "", "", nil, err
	}
	target := pickTargetBranch(branches)
	source := pickSourceBranch(branches, target)
	return source, target, branches, nil
}

func (m *Manager) run(ctx context.Context, dir string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", dir}, args...)
	cmd := commandContext(ctx, "git", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s 失败: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}

func (m *Manager) runGitDir(ctx context.Context, gitDir string, args ...string) (string, error) {
	cmdArgs := append([]string{"--git-dir", gitDir}, args...)
	cmd := commandContext(ctx, "git", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s 失败: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}

func splitNonEmptyLines(s string) []string {
	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func candidateRefs(ref string) []string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil
	}
	candidates := []string{ref}
	if !strings.HasPrefix(ref, "origin/") && !strings.HasPrefix(ref, "refs/") {
		candidates = append(candidates,
			"origin/"+ref,
			"refs/heads/"+ref,
			"refs/remotes/origin/"+ref,
		)
	}
	return uniqueStrings(candidates)
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".java":
		return "java"
	case ".kt":
		return "kotlin"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".jsx":
		return "jsx"
	case ".py":
		return "python"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".cs":
		return "csharp"
	case ".sql":
		return "sql"
	case ".vue":
		return "vue"
	case ".yml", ".yaml":
		return "yaml"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	default:
		return strings.TrimPrefix(ext, ".")
	}
}

func normalizeBranchName(branch, prefix string) string {
	branch = strings.TrimSpace(branch)
	if branch == "" || branch == "HEAD" || branch == "origin/HEAD" {
		return ""
	}
	if prefix != "" && strings.HasPrefix(branch, prefix) {
		branch = strings.TrimPrefix(branch, prefix)
	}
	if branch == "HEAD" {
		return ""
	}
	return branch
}

func pickTargetBranch(branches []string) string {
	set := map[string]struct{}{}
	for _, branch := range branches {
		set[strings.TrimSpace(branch)] = struct{}{}
	}
	for _, candidate := range []string{"master", "main", "develop"} {
		if _, ok := set[candidate]; ok {
			return candidate
		}
	}
	if len(branches) > 0 {
		return branches[0]
	}
	return ""
}

func pickSourceBranch(branches []string, target string) string {
	skip := map[string]struct{}{
		strings.TrimSpace(target): {},
		"master":                  {},
		"main":                    {},
		"develop":                 {},
	}
	for _, branch := range branches {
		if _, ok := skip[strings.TrimSpace(branch)]; ok {
			continue
		}
		return branch
	}
	for _, branch := range branches {
		if strings.TrimSpace(branch) != strings.TrimSpace(target) {
			return branch
		}
	}
	if len(branches) > 0 {
		return branches[0]
	}
	return ""
}

func compactStrings(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return uniqueStrings(out)
}

func uniqueStrings(items []string) []string {
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

func repoKey(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func sortedStrings(items []string) []string {
	out := append([]string(nil), items...)
	sort.Strings(out)
	return out
}
