package scanner

import (
	"regexp"
	"strings"

	"aiguard/internal/model"
)

type Result struct {
	Findings []model.Finding
	Hints    map[string][]string
}

type Scanner struct{}

func New() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Run(diff *model.DiffSet) Result {
	result := Result{
		Hints: map[string][]string{},
	}

	for _, file := range diff.Files {
		if strings.TrimSpace(file.SourceContent) == "" {
			continue
		}
		lines := strings.Split(file.SourceContent, "\n")
		s.scanSecrets(&result, file, lines)
		s.scanSQLConcat(&result, file, lines)
		s.scanCommandExec(&result, file, lines)
		s.scanTLS(&result, file, lines)
		s.scanExceptionSwallow(&result, file, lines)
		s.scanSensitiveLog(&result, file, lines)
		s.scanNoTimeoutHTTP(&result, file, lines)
		s.scanOpenWithoutClose(&result, file, lines)
		s.scanPathTraversal(&result, file, lines)
	}
	return result
}

func (s *Scanner) scanSecrets(result *Result, file model.ChangedFile, lines []string) {
	re := regexp.MustCompile(`(?i)(api[_-]?key|secret|password|token)\s*[:=]\s*["'][^"']{8,}["']`)
	s.scanLineRegex(result, file, lines, re, findingTemplate{
		Title:          "疑似硬编码敏感信息",
		Severity:       "严重",
		Category:       "安全",
		Confidence:     "medium",
		Description:    "发现疑似将密钥、密码或令牌直接硬编码在代码或配置中，存在敏感信息泄露风险。",
		Impact:         "一旦代码被泄露或日志暴露，可能导致凭证泄漏、横向移动或环境被接管。",
		Recommendation: "将密钥迁移到安全配置中心、环境变量或密钥管理系统，并清理历史提交中的敏感值。",
	})
}

func (s *Scanner) scanSQLConcat(result *Result, file model.ChangedFile, lines []string) {
	re := regexp.MustCompile(`(?i)(select|insert|update|delete).*(\+|fmt\.sprintf|\$\{)`)
	s.scanLineRegex(result, file, lines, re, findingTemplate{
		Title:          "疑似 SQL 字符串拼接",
		Severity:       "严重",
		Category:       "安全",
		Confidence:     "medium",
		Description:    "发现 SQL 语句可能通过字符串拼接构造，容易引入注入风险。",
		Impact:         "攻击者若可控制输入内容，可能篡改查询条件或执行未授权的数据操作。",
		Recommendation: "改用参数化查询、预编译语句或 ORM 参数绑定，并补充输入校验与安全测试。",
	})
}

func (s *Scanner) scanCommandExec(result *Result, file model.ChangedFile, lines []string) {
	re := regexp.MustCompile(`(?i)(exec\.command\("sh",\s*"-c"|runtime\.getruntime\(\)\.exec|subprocess\.(popen|run)\()`)
	s.scanLineRegex(result, file, lines, re, findingTemplate{
		Title:          "疑似危险命令执行",
		Severity:       "高危",
		Category:       "安全",
		Confidence:     "medium",
		Description:    "发现通过 shell 或系统命令执行外部输入的迹象，存在命令注入风险。",
		Impact:         "若参数未严格校验，攻击者可能借此执行任意命令或读取敏感文件。",
		Recommendation: "避免经由 shell 执行拼接命令，优先使用参数化 API，并对白名单参数做严格校验。",
	})
}

func (s *Scanner) scanTLS(result *Result, file model.ChangedFile, lines []string) {
	re := regexp.MustCompile(`(?i)insecureskipverify\s*:\s*true`)
	s.scanLineRegex(result, file, lines, re, findingTemplate{
		Title:          "关闭 TLS 证书校验",
		Severity:       "严重",
		Category:       "安全",
		Confidence:     "high",
		Description:    "检测到 TLS 证书校验被显式关闭，容易遭受中间人攻击。",
		Impact:         "攻击者可伪造服务端证书，窃取流量中的认证信息和业务数据。",
		Recommendation: "仅在测试环境局部使用，生产环境应恢复证书校验并正确配置可信 CA。",
	})
}

func (s *Scanner) scanExceptionSwallow(result *Result, file model.ChangedFile, lines []string) {
	re := regexp.MustCompile(`(?i)(catch\s*\([^)]*\)\s*\{\s*\}|except\s+exception\s*:\s*pass)`)
	s.scanLineRegex(result, file, lines, re, findingTemplate{
		Title:          "异常被吞掉",
		Severity:       "一般",
		Category:       "健壮性",
		Confidence:     "medium",
		Description:    "发现异常处理为空或直接忽略，可能掩盖真实故障。",
		Impact:         "线上问题可能无法被及时发现，导致数据状态不一致或错误传播。",
		Recommendation: "记录必要日志、返回明确错误并为关键路径补充监控与告警。",
	})
}

func (s *Scanner) scanSensitiveLog(result *Result, file model.ChangedFile, lines []string) {
	re := regexp.MustCompile(`(?i)(log\.|logger\.|console\.log).*(password|token|secret|apikey)`)
	s.scanLineRegex(result, file, lines, re, findingTemplate{
		Title:          "日志中输出敏感信息",
		Severity:       "严重",
		Category:       "安全",
		Confidence:     "medium",
		Description:    "发现日志语句中可能输出口令、令牌或密钥等敏感字段。",
		Impact:         "敏感信息可能在日志平台、控制台或异常栈中扩散，增加泄露面。",
		Recommendation: "对敏感字段做脱敏或完全禁止输出，并检查下游日志系统的保留策略。",
	})
}

func (s *Scanner) scanNoTimeoutHTTP(result *Result, file model.ChangedFile, lines []string) {
	for i, line := range lines {
		if !strings.Contains(line, "http.Client{") {
			continue
		}
		end := min(len(lines), i+6)
		block := strings.Join(lines[i:end], "\n")
		if strings.Contains(block, "Timeout:") {
			continue
		}
		s.addFinding(result, model.Finding{
			Title:          "HTTP 客户端未设置超时",
			Severity:       "一般",
			Category:       "性能",
			Confidence:     "medium",
			File:           file.Path,
			LineStart:      i + 1,
			LineEnd:        i + 1,
			Description:    "发现 Go HTTP 客户端未设置 Timeout，网络异常时可能长时间阻塞。",
			Impact:         "在高并发或下游不稳定时，可能导致协程堆积、资源耗尽和响应超时扩散。",
			Evidence:       strings.TrimSpace(line),
			Recommendation: "为 HTTP 客户端显式设置 Timeout，并在调用链中配合 context 控制超时与取消。",
		})
	}
}

func (s *Scanner) scanOpenWithoutClose(result *Result, file model.ChangedFile, lines []string) {
	for i, line := range lines {
		if !(strings.Contains(line, "os.Open(") || strings.Contains(line, "OpenFile(")) {
			continue
		}
		end := min(len(lines), i+8)
		block := strings.Join(lines[i:end], "\n")
		if strings.Contains(block, ".Close(") || strings.Contains(block, "defer") {
			continue
		}
		s.addFinding(result, model.Finding{
			Title:          "文件句柄可能未关闭",
			Severity:       "一般",
			Category:       "性能",
			Confidence:     "medium",
			File:           file.Path,
			LineStart:      i + 1,
			LineEnd:        i + 1,
			Description:    "打开文件后未在相邻上下文中发现关闭逻辑，存在句柄泄露风险。",
			Impact:         "长时间运行后可能耗尽文件句柄，影响服务稳定性或触发 I/O 异常。",
			Evidence:       strings.TrimSpace(line),
			Recommendation: "在成功打开文件后及时 defer Close，并对异常路径做资源释放保障。",
		})
	}
}

func (s *Scanner) scanPathTraversal(result *Result, file model.ChangedFile, lines []string) {
	re := regexp.MustCompile(`(?i)(filepath\.join|path\.join).*(req\.|request\.|input|param|query|body)`)
	s.scanLineRegex(result, file, lines, re, findingTemplate{
		Title:          "疑似路径拼接风险",
		Severity:       "一般",
		Category:       "安全",
		Confidence:     "medium",
		Description:    "发现路径拼接直接使用外部输入，若缺少规范化与白名单校验，可能引入路径穿越。",
		Impact:         "攻击者可能借此读取或覆盖预期目录之外的文件。",
		Recommendation: "对输入路径做规范化、前缀校验和白名单限制，避免直接信任用户可控路径片段。",
	})
}

type findingTemplate struct {
	Title          string
	Severity       string
	Category       string
	Confidence     string
	Description    string
	Impact         string
	Recommendation string
}

func (s *Scanner) scanLineRegex(result *Result, file model.ChangedFile, lines []string, re *regexp.Regexp, tmpl findingTemplate) {
	for i, line := range lines {
		if !re.MatchString(line) {
			continue
		}
		s.addFinding(result, model.Finding{
			Title:          tmpl.Title,
			Severity:       tmpl.Severity,
			Category:       tmpl.Category,
			Confidence:     tmpl.Confidence,
			File:           file.Path,
			LineStart:      i + 1,
			LineEnd:        i + 1,
			Description:    tmpl.Description,
			Impact:         tmpl.Impact,
			Evidence:       strings.TrimSpace(line),
			Recommendation: tmpl.Recommendation,
		})
	}
}

func (s *Scanner) addFinding(result *Result, finding model.Finding) {
	result.Findings = append(result.Findings, finding)
	result.Hints[finding.File] = append(result.Hints[finding.File], finding.Title)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
