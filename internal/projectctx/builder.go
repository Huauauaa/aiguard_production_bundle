package projectctx

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aiguard/internal/model"
)

type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) Build(repoRoot string, diff *model.DiffSet, customRuleFile string) (model.ProjectBrief, error) {
	brief := model.ProjectBrief{}

	brief.Purpose = b.detectPurpose(repoRoot)
	brief.TechStack = b.detectTechStack(repoRoot)
	brief.Frameworks = b.detectFrameworks(repoRoot)
	brief.Modules = b.detectModules(repoRoot, diff)
	brief.Entrypoints = b.detectEntrypoints(repoRoot)
	brief.SensitiveAreas = b.detectSensitiveAreas(repoRoot, diff)
	brief.PerformanceHotspots = b.detectPerformanceHotspots(repoRoot, diff)
	brief.ChangeImpactPath = b.detectChangeImpact(diff)
	brief.RepoRules = b.readRepoRules(repoRoot, customRuleFile)
	brief.Notes = b.buildNotes(repoRoot, diff)

	return brief, nil
}

func (b *Builder) detectPurpose(repoRoot string) string {
	for _, name := range []string{"README.md", "README.MD", "readme.md"} {
		path := filepath.Join(repoRoot, name)
		if data, err := os.ReadFile(path); err == nil {
			text := strings.TrimSpace(stripMarkdown(string(data)))
			paragraphs := strings.Split(text, "\n")
			chunks := make([]string, 0, 3)
			for _, line := range paragraphs {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				chunks = append(chunks, line)
				if len(chunks) >= 3 {
					break
				}
			}
			if len(chunks) > 0 {
				purpose := strings.Join(chunks, " ")
				if len(purpose) > 220 {
					purpose = purpose[:220] + "..."
				}
				return purpose
			}
		}
	}
	return "未从 README 中识别出明确的项目用途，建议在仓库中补充背景说明。"
}

func (b *Builder) detectTechStack(repoRoot string) []string {
	items := []string{}
	if exists(filepath.Join(repoRoot, "go.mod")) {
		items = append(items, "Go")
	}
	if exists(filepath.Join(repoRoot, "package.json")) {
		items = append(items, "Node.js")
	}
	if exists(filepath.Join(repoRoot, "pnpm-lock.yaml")) || exists(filepath.Join(repoRoot, "yarn.lock")) {
		items = append(items, "JavaScript/TypeScript")
	}
	if exists(filepath.Join(repoRoot, "pom.xml")) || exists(filepath.Join(repoRoot, "build.gradle")) || exists(filepath.Join(repoRoot, "build.gradle.kts")) {
		items = append(items, "Java")
	}
	if exists(filepath.Join(repoRoot, "requirements.txt")) || exists(filepath.Join(repoRoot, "pyproject.toml")) {
		items = append(items, "Python")
	}
	if exists(filepath.Join(repoRoot, "Cargo.toml")) {
		items = append(items, "Rust")
	}
	if exists(filepath.Join(repoRoot, "Dockerfile")) {
		items = append(items, "Docker")
	}
	return unique(items)
}

func (b *Builder) detectFrameworks(repoRoot string) []string {
	frameworks := []string{}
	appendIfContains := func(path string, pairs map[string]string) {
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		lower := strings.ToLower(string(data))
		for keyword, name := range pairs {
			if strings.Contains(lower, keyword) {
				frameworks = append(frameworks, name)
			}
		}
	}

	appendIfContains(filepath.Join(repoRoot, "package.json"), map[string]string{
		"react":   "React",
		"vue":     "Vue",
		"next":    "Next.js",
		"nuxt":    "Nuxt",
		"nestjs":  "NestJS",
		"express": "Express",
	})
	appendIfContains(filepath.Join(repoRoot, "go.mod"), map[string]string{
		"github.com/gin-gonic/gin": "Gin",
		"github.com/labstack/echo": "Echo",
		"github.com/gofiber/fiber": "Fiber",
	})
	appendIfContains(filepath.Join(repoRoot, "pom.xml"), map[string]string{
		"spring-boot": "Spring Boot",
		"mybatis":     "MyBatis",
		"hibernate":   "Hibernate",
	})
	appendIfContains(filepath.Join(repoRoot, "requirements.txt"), map[string]string{
		"django":  "Django",
		"fastapi": "FastAPI",
		"flask":   "Flask",
	})
	appendIfContains(filepath.Join(repoRoot, "pyproject.toml"), map[string]string{
		"django":  "Django",
		"fastapi": "FastAPI",
		"flask":   "Flask",
	})
	return unique(frameworks)
}

func (b *Builder) detectModules(repoRoot string, diff *model.DiffSet) []string {
	modules := []string{}
	for _, file := range diff.Files {
		parts := splitPath(file.Path)
		if len(parts) == 0 {
			continue
		}
		if len(parts) >= 2 {
			modules = append(modules, parts[0]+"/"+parts[1])
		} else {
			modules = append(modules, parts[0])
		}
	}

	for _, dir := range []string{"cmd", "internal", "pkg", "src", "app", "service", "services"} {
		path := filepath.Join(repoRoot, dir)
		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				modules = append(modules, filepath.ToSlash(filepath.Join(dir, entry.Name())))
			}
		}
	}
	sort.Strings(modules)
	return unique(limitStrings(modules, 20))
}

func (b *Builder) detectEntrypoints(repoRoot string) []string {
	entries := []string{}
	candidates := []string{
		"main.go",
		"cmd",
		"src/main.ts",
		"src/main.js",
		"src/main/java",
		"server.js",
		"server.ts",
		"app.py",
		"manage.py",
	}
	for _, candidate := range candidates {
		if exists(filepath.Join(repoRoot, candidate)) {
			entries = append(entries, candidate)
		}
	}
	return unique(entries)
}

func (b *Builder) detectSensitiveAreas(repoRoot string, diff *model.DiffSet) []string {
	keywords := []string{"auth", "login", "permission", "role", "security", "payment", "billing", "export", "admin", "token", "session"}
	areas := []string{}
	for _, file := range diff.Files {
		lower := strings.ToLower(file.Path)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				areas = append(areas, file.Path)
				break
			}
		}
	}
	return unique(limitStrings(areas, 12))
}

func (b *Builder) detectPerformanceHotspots(repoRoot string, diff *model.DiffSet) []string {
	keywords := []string{"query", "repository", "dao", "cache", "batch", "stream", "export", "sync", "worker", "queue"}
	items := []string{}
	for _, file := range diff.Files {
		lower := strings.ToLower(file.Path)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				items = append(items, file.Path)
				break
			}
		}
	}
	return unique(limitStrings(items, 12))
}

func (b *Builder) detectChangeImpact(diff *model.DiffSet) []string {
	items := []string{}
	for _, file := range diff.Files {
		parts := splitPath(file.Path)
		if len(parts) >= 3 {
			items = append(items, parts[0]+"/"+parts[1]+"/"+parts[2])
		} else if len(parts) >= 2 {
			items = append(items, parts[0]+"/"+parts[1])
		} else if len(parts) == 1 {
			items = append(items, parts[0])
		}
	}
	return unique(limitStrings(items, 15))
}

func (b *Builder) readRepoRules(repoRoot string, customRuleFile string) []string {
	paths := []string{}
	if strings.TrimSpace(customRuleFile) != "" {
		if filepath.IsAbs(customRuleFile) {
			paths = append(paths, customRuleFile)
		} else {
			paths = append(paths, filepath.Join(repoRoot, customRuleFile))
		}
	}
	paths = append(paths, filepath.Join(repoRoot, "AIGUARD.md"))

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		items := []string{}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
				line = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
				if line != "" {
					items = append(items, line)
				}
			}
		}
		return unique(limitStrings(items, 20))
	}
	return nil
}

func (b *Builder) buildNotes(repoRoot string, diff *model.DiffSet) []string {
	notes := []string{}
	if exists(filepath.Join(repoRoot, ".github", "workflows")) || exists(filepath.Join(repoRoot, ".gitlab-ci.yml")) {
		notes = append(notes, "仓库包含 CI/CD 配置，可在后续版本中纳入流水线与部署风险分析。")
	}
	if len(diff.Files) > 30 {
		notes = append(notes, "本次变更文件较多，建议优先关注高敏感目录与配置变更。")
	}
	if exists(filepath.Join(repoRoot, "Dockerfile")) {
		notes = append(notes, "仓库包含容器构建配置，可在后续版本增加容器安全与镜像最佳实践检查。")
	}
	return notes
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func splitPath(path string) []string {
	path = filepath.ToSlash(path)
	parts := strings.Split(path, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func stripMarkdown(s string) string {
	replacer := strings.NewReplacer("#", "", "`", "", "*", "", "_", "", ">", "", "|", " ")
	return replacer.Replace(s)
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

func limitStrings(items []string, limit int) []string {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}
