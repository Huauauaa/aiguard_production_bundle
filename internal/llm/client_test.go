package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"aiguard/internal/config"
)

func TestClientEnabled(t *testing.T) {
	cfg := config.Config{
		OpenAI: config.OpenAIConfig{
			BaseURL:      "http://localhost",
			DefaultModel: "test-model",
		},
		Runtime: config.RuntimeConfig{
			RequestTimeoutSec: 30,
			MaxRetry:          2,
		},
	}
	client := New(cfg)
	if !client.Enabled() {
		t.Error("client should be enabled")
	}
}

func TestClientDisabled(t *testing.T) {
	cfg := config.Config{}
	client := New(cfg)
	if client.Enabled() {
		t.Error("client should be disabled")
	}
}

func TestPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"OK"}}]}`))
	}))
	defer server.Close()

	cfg := config.Config{
		OpenAI: config.OpenAIConfig{
			BaseURL:      server.URL,
			DefaultModel: "test",
		},
		Runtime: config.RuntimeConfig{
			RequestTimeoutSec: 30,
			MaxRetry:          1,
		},
	}
	client := New(cfg)
	err := client.Ping(context.Background())
	if err != nil {
		t.Errorf("ping failed: %v", err)
	}
}
