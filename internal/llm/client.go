package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aiguard/internal/config"
)

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
	maxRetry   int
}

func New(cfg config.Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(cfg.OpenAI.BaseURL), "/"),
		apiKey:  strings.TrimSpace(cfg.OpenAI.APIKey),
		model:   strings.TrimSpace(cfg.OpenAI.DefaultModel),
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Runtime.RequestTimeoutSec) * time.Second,
		},
		maxRetry: cfg.Runtime.MaxRetry,
	}
}

func (c *Client) Enabled() bool {
	return c.baseURL != "" && c.model != ""
}

func (c *Client) ChatJSON(ctx context.Context, systemPrompt, userPrompt string, maxTokens int, out any) error {
	if !c.Enabled() {
		return errors.New("LLM 未配置")
	}

	var lastErr error
	prompt := userPrompt
	for attempt := 0; attempt <= c.maxRetry; attempt++ {
		content, err := c.chat(ctx, systemPrompt, prompt, maxTokens)
		if err != nil {
			lastErr = err
			continue
		}
		payload, err := extractJSON(content)
		if err != nil {
			lastErr = err
			prompt = userPrompt + "\n\n请只输出合法 JSON，不要输出解释文字或 Markdown 代码块。"
			continue
		}
		if err := json.Unmarshal([]byte(payload), out); err != nil {
			lastErr = err
			prompt = userPrompt + "\n\n请只输出合法 JSON，不要输出解释文字或 Markdown 代码块。"
			continue
		}
		return nil
	}
	if lastErr == nil {
		lastErr = errors.New("模型未返回可解析的 JSON")
	}
	return lastErr
}

func (c *Client) chat(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	endpoint := c.baseURL
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint += "/chat/completions"
	}

	body := map[string]any{
		"model":       c.model,
		"temperature": 0.2,
		"max_tokens":  maxTokens,
		"stream":      false,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("模型服务返回错误: %s", string(data))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content any `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("模型未返回 choices")
	}

	switch content := parsed.Choices[0].Message.Content.(type) {
	case string:
		return content, nil
	case []any:
		parts := []string{}
		for _, item := range content {
			if obj, ok := item.(map[string]any); ok {
				if text, ok := obj["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "\n"), nil
	default:
		return "", errors.New("模型返回内容格式无法识别")
	}
}

func extractJSON(content string) (string, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	for _, openClose := range [][2]rune{{'{', '}'}, {'[', ']'}} {
		open, close := openClose[0], openClose[1]
		start := strings.IndexRune(content, open)
		if start < 0 {
			continue
		}
		depth := 0
		inString := false
		escaped := false
		for i, r := range content[start:] {
			if inString {
				if escaped {
					escaped = false
					continue
				}
				if r == '\\' {
					escaped = true
					continue
				}
				if r == '"' {
					inString = false
				}
				continue
			}
			if r == '"' {
				inString = true
				continue
			}
			if r == open {
				depth++
			}
			if r == close {
				depth--
				if depth == 0 {
					return content[start : start+i+1], nil
				}
			}
		}
	}
	return "", errors.New("未找到有效 JSON")
}
