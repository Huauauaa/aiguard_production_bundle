package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type OpenAIConfig struct {
	BaseURL      string `yaml:"base_url"`
	APIKey       string `yaml:"api_key"`
	DefaultModel string `yaml:"default_model"`
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
}

type RulesConfig struct {
	CustomRuleFile string   `yaml:"custom_rule_file"`
	Ignore         []string `yaml:"ignore"`
}

type Config struct {
	OpenAI  OpenAIConfig  `yaml:"openai"`
	Runtime RuntimeConfig `yaml:"runtime"`
	Review  ReviewConfig  `yaml:"review"`
	Rules   RulesConfig   `yaml:"rules"`
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
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
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
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.OpenAI.BaseURL) == "" {
		return fmt.Errorf("配置缺失: openai.base_url 或环境变量 OPENAI_BASE_URL / AIGUARD_OPENAI_BASE_URL")
	}
	if strings.TrimSpace(c.OpenAI.DefaultModel) == "" {
		return fmt.Errorf("配置缺失: openai.default_model 或环境变量 OPENAI_DEFAULT_MODEL / AIGUARD_OPENAI_DEFAULT_MODEL")
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

	setString(&c.OpenAI.BaseURL, "AIGUARD_OPENAI_BASE_URL", "OPENAI_BASE_URL")
	setString(&c.OpenAI.APIKey, "AIGUARD_OPENAI_API_KEY", "OPENAI_API_KEY")
	setString(&c.OpenAI.DefaultModel, "AIGUARD_OPENAI_DEFAULT_MODEL", "OPENAI_DEFAULT_MODEL")
	setString(&c.Review.WorkspaceDir, "AIGUARD_WORKSPACE_DIR")
	setString(&c.Runtime.LogLevel, "AIGUARD_LOG_LEVEL")
	setInt(&c.Runtime.Concurrency, "AIGUARD_CONCURRENCY")
	setInt(&c.Runtime.SafeInputTokens, "AIGUARD_SAFE_INPUT_TOKENS")
	setInt(&c.Runtime.ReservedOutput, "AIGUARD_RESERVED_OUTPUT_TOKENS")
}
