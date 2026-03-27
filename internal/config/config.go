package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type ProxyConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
	HTTP    string `yaml:"http"`
	HTTPS   string `yaml:"https"`
	NoProxy string `yaml:"no_proxy"`
}

type OpenAIConfig struct {
	BaseURL      string      `yaml:"base_url"`
	APIKey       string      `yaml:"api_key"`
	DefaultModel string      `yaml:"default_model"`
	Proxy        ProxyConfig `yaml:"proxy"`
}

type RuntimeConfig struct {
	RequestTimeoutSec int    `yaml:"request_timeout_sec"`
	Concurrency       int    `yaml:"concurrency"`
	MaxRetry          int    `yaml:"max_retry"`
	SafeInputTokens   int    `yaml:"safe_input_tokens"`
	ReservedOutput    int    `yaml:"reserved_output_tokens"`
	LogLevel          string `yaml:"log_level"`
}

type ReviewConfig struct {
	WorkspaceDir           string   `yaml:"workspace_dir"`
	DiffStrategy           string   `yaml:"diff_strategy"`
	MaxChangedFiles        int      `yaml:"max_changed_files"`
	MaxHunksPerFile        int      `yaml:"max_hunks_per_file"`
	ExportFormats          []string `yaml:"export_formats"`
	EnableProjectBrief     bool     `yaml:"enable_project_brief"`
	EnablePreScan          bool     `yaml:"enable_prescan"`
	RedactSecretsBeforeLLM bool     `yaml:"redact_secrets_before_llm"`
	// Base URL of the PR web portal. Used to generate code-jump hyperlinks in reports.
	// Example: https://pr.example.com
	CodeBrowseBaseURL string `yaml:"code_browse_base_url"`
}

type RulesConfig struct {
	CustomRuleFile string   `yaml:"custom_rule_file"`
	Ignore         []string `yaml:"ignore"`
}

type GitEndpointConfig struct {
	Scheme string `yaml:"scheme"`
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	User   string `yaml:"user"`
}

type GitProviderConfig struct {
	SSH   GitEndpointConfig `yaml:"ssh"`
	HTTPS GitEndpointConfig `yaml:"https"`
}

type GitConfig struct {
	PreferredProtocol string            `yaml:"preferred_protocol"`
	GitLab            GitProviderConfig `yaml:"gitlab"`
	GitHub            GitProviderConfig `yaml:"github"`
}

type Config struct {
	OpenAI  OpenAIConfig  `yaml:"openai"`
	Runtime RuntimeConfig `yaml:"runtime"`
	Review  ReviewConfig  `yaml:"review"`
	Rules   RulesConfig   `yaml:"rules"`
	Git     GitConfig     `yaml:"git"`
}

func Default() Config {
	return Config{
		Runtime: RuntimeConfig{
			RequestTimeoutSec: 180,
			Concurrency:       4,
			MaxRetry:          2,
			SafeInputTokens:   160000,
			ReservedOutput:    12000,
			LogLevel:          "info",
		},
		Review: ReviewConfig{
			WorkspaceDir:           "./workspace",
			DiffStrategy:           "merge_base",
			MaxChangedFiles:        200,
			MaxHunksPerFile:        40,
			ExportFormats:          []string{"html", "md", "json"},
			EnableProjectBrief:     true,
			EnablePreScan:          true,
			RedactSecretsBeforeLLM: true,
		},
		Rules: RulesConfig{
			Ignore: []string{
				"node_modules/**",
				"dist/**",
				"build/**",
				"*.min.js",
				"*.lock",
			},
		},
		Git: GitConfig{
			PreferredProtocol: "ssh",
			GitLab: GitProviderConfig{
				SSH: GitEndpointConfig{User: "git"},
				HTTPS: GitEndpointConfig{
					Scheme: "https",
				},
			},
			GitHub: GitProviderConfig{
				SSH: GitEndpointConfig{User: "git"},
				HTTPS: GitEndpointConfig{
					Scheme: "https",
				},
			},
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if strings.TrimSpace(path) != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return cfg, err
		}

		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, err
		}
	}

	cfg.applyEnvOverrides()
	cfg.normalize()
	return cfg, nil
}

func (c *Config) normalize() {
	if c.Runtime.RequestTimeoutSec <= 0 {
		c.Runtime.RequestTimeoutSec = 180
	}
	if c.Runtime.Concurrency <= 0 {
		c.Runtime.Concurrency = 4
	}
	if c.Runtime.MaxRetry < 0 {
		c.Runtime.MaxRetry = 0
	}
	if c.Runtime.SafeInputTokens <= 0 {
		c.Runtime.SafeInputTokens = 160000
	}
	if c.Runtime.ReservedOutput <= 0 {
		c.Runtime.ReservedOutput = 12000
	}
	if strings.TrimSpace(c.Review.WorkspaceDir) == "" {
		c.Review.WorkspaceDir = "./workspace"
	}
	if c.Review.MaxChangedFiles <= 0 {
		c.Review.MaxChangedFiles = 200
	}
	if c.Review.MaxHunksPerFile <= 0 {
		c.Review.MaxHunksPerFile = 40
	}
	if len(c.Review.ExportFormats) == 0 {
		c.Review.ExportFormats = []string{"html", "md", "json"}
	}
	if strings.TrimSpace(c.Review.DiffStrategy) == "" {
		c.Review.DiffStrategy = "merge_base"
	}
	if strings.TrimSpace(c.Runtime.LogLevel) == "" {
		c.Runtime.LogLevel = "info"
	}
	if len(c.Rules.Ignore) == 0 {
		c.Rules.Ignore = Default().Rules.Ignore
	}

	c.OpenAI.BaseURL = strings.TrimSpace(c.OpenAI.BaseURL)
	c.OpenAI.APIKey = strings.TrimSpace(c.OpenAI.APIKey)
	c.OpenAI.DefaultModel = strings.TrimSpace(c.OpenAI.DefaultModel)
	c.OpenAI.Proxy.URL = strings.TrimSpace(c.OpenAI.Proxy.URL)
	c.OpenAI.Proxy.HTTP = strings.TrimSpace(c.OpenAI.Proxy.HTTP)
	c.OpenAI.Proxy.HTTPS = strings.TrimSpace(c.OpenAI.Proxy.HTTPS)
	c.OpenAI.Proxy.NoProxy = strings.TrimSpace(c.OpenAI.Proxy.NoProxy)
	c.Review.CodeBrowseBaseURL = strings.TrimSpace(c.Review.CodeBrowseBaseURL)

	c.Git.PreferredProtocol = strings.ToLower(strings.TrimSpace(c.Git.PreferredProtocol))
	if c.Git.PreferredProtocol != "ssh" && c.Git.PreferredProtocol != "https" {
		c.Git.PreferredProtocol = "ssh"
	}
	normalizeProvider := func(provider *GitProviderConfig) {
		provider.SSH.Scheme = strings.TrimSpace(provider.SSH.Scheme)
		provider.SSH.Host = strings.TrimSpace(provider.SSH.Host)
		provider.SSH.Port = strings.TrimSpace(provider.SSH.Port)
		provider.SSH.User = strings.TrimSpace(provider.SSH.User)
		provider.HTTPS.Scheme = strings.TrimSpace(provider.HTTPS.Scheme)
		provider.HTTPS.Host = strings.TrimSpace(provider.HTTPS.Host)
		provider.HTTPS.Port = strings.TrimSpace(provider.HTTPS.Port)
		provider.HTTPS.User = strings.TrimSpace(provider.HTTPS.User)
		if provider.SSH.User == "" {
			provider.SSH.User = "git"
		}
		if provider.HTTPS.Scheme == "" {
			provider.HTTPS.Scheme = "https"
		}
	}
	normalizeProvider(&c.Git.GitLab)
	normalizeProvider(&c.Git.GitHub)
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.OpenAI.BaseURL) == "" {
		return fmt.Errorf("配置缺失: openai.base_url 或环境变量 OPENAI_BASE_URL / AIGUARD_OPENAI_BASE_URL")
	}
	if strings.TrimSpace(c.OpenAI.DefaultModel) == "" {
		return fmt.Errorf("配置缺失: openai.default_model 或环境变量 OPENAI_DEFAULT_MODEL / AIGUARD_OPENAI_DEFAULT_MODEL")
	}
	if c.OpenAI.Proxy.Enabled {
		if strings.TrimSpace(c.OpenAI.Proxy.URL) == "" && strings.TrimSpace(c.OpenAI.Proxy.HTTP) == "" && strings.TrimSpace(c.OpenAI.Proxy.HTTPS) == "" {
			return fmt.Errorf("配置缺失: openai.proxy.url 或 openai.proxy.http / openai.proxy.https")
		}
		for _, item := range []struct {
			name  string
			value string
		}{
			{name: "openai.proxy.url", value: c.OpenAI.Proxy.URL},
			{name: "openai.proxy.http", value: c.OpenAI.Proxy.HTTP},
			{name: "openai.proxy.https", value: c.OpenAI.Proxy.HTTPS},
		} {
			if strings.TrimSpace(item.value) == "" {
				continue
			}
			if _, err := url.Parse(item.value); err != nil {
				return fmt.Errorf("%s 配置非法: %w", item.name, err)
			}
		}
	}
	return nil
}

func (c *Config) applyEnvOverrides() {
	setString := func(target *string, keys ...string) {
		for _, key := range keys {
			if value := strings.TrimSpace(os.Getenv(key)); value != "" {
				*target = value
				return
			}
		}
	}

	setInt := func(target *int, keys ...string) {
		for _, key := range keys {
			if raw := strings.TrimSpace(os.Getenv(key)); raw != "" {
				if value, err := strconv.Atoi(raw); err == nil {
					*target = value
					return
				}
			}
		}
	}

	setBool := func(target *bool, keys ...string) {
		for _, key := range keys {
			if raw := strings.TrimSpace(os.Getenv(key)); raw != "" {
				if value, err := strconv.ParseBool(raw); err == nil {
					*target = value
					return
				}
			}
		}
	}

	setString(&c.OpenAI.BaseURL, "AIGUARD_OPENAI_BASE_URL", "OPENAI_BASE_URL")
	setString(&c.OpenAI.APIKey, "AIGUARD_OPENAI_API_KEY", "OPENAI_API_KEY")
	setString(&c.OpenAI.DefaultModel, "AIGUARD_OPENAI_DEFAULT_MODEL", "OPENAI_DEFAULT_MODEL")
	setBool(&c.OpenAI.Proxy.Enabled, "AIGUARD_OPENAI_PROXY_ENABLED", "OPENAI_PROXY_ENABLED")
	setString(&c.OpenAI.Proxy.URL, "AIGUARD_OPENAI_PROXY_URL", "OPENAI_PROXY_URL")
	setString(&c.OpenAI.Proxy.HTTP, "AIGUARD_OPENAI_PROXY_HTTP", "OPENAI_PROXY_HTTP")
	setString(&c.OpenAI.Proxy.HTTPS, "AIGUARD_OPENAI_PROXY_HTTPS", "OPENAI_PROXY_HTTPS")
	setString(&c.OpenAI.Proxy.NoProxy, "AIGUARD_OPENAI_PROXY_NO_PROXY", "OPENAI_PROXY_NO_PROXY")
	setString(&c.Review.WorkspaceDir, "AIGUARD_WORKSPACE_DIR")
	setString(&c.Review.CodeBrowseBaseURL, "AIGUARD_CODE_BROWSE_BASE_URL")
	setString(&c.Runtime.LogLevel, "AIGUARD_LOG_LEVEL")
	setInt(&c.Runtime.Concurrency, "AIGUARD_CONCURRENCY")
	setInt(&c.Runtime.SafeInputTokens, "AIGUARD_SAFE_INPUT_TOKENS")
	setInt(&c.Runtime.ReservedOutput, "AIGUARD_RESERVED_OUTPUT_TOKENS")
}
