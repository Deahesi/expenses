package config

import (
	"strings"
	"testing"
)

func TestLoadLLMConfig(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		timezone string
		wantErr  string
	}{
		{name: "defaults", key: "test-key", timezone: "UTC"},
		{name: "configured timezone", key: "test-key", timezone: "Asia/Novosibirsk"},
		{name: "missing key", timezone: "UTC", wantErr: "OPENROUTER_API_KEY"},
		{name: "blank key", key: "  ", timezone: "UTC", wantErr: "OPENROUTER_API_KEY"},
		{name: "invalid timezone", key: "test-key", timezone: "invalid/timezone", wantErr: "LLM_TIMEZONE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://test")
			t.Setenv("TELEGRAM_BOT_API_KEY", "test-token")
			t.Setenv("OPENROUTER_API_KEY", tt.key)
			t.Setenv("LLM_MODEL", "")
			t.Setenv("LLM_TIMEZONE", tt.timezone)

			cfg, err := Load()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Load() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.LLM.Model != "openrouter/openai/gpt-4o-mini" {
				t.Errorf("Model = %q", cfg.LLM.Model)
			}
			if cfg.LLM.Timezone.String() != tt.timezone {
				t.Errorf("Timezone = %q, want %q", cfg.LLM.Timezone, tt.timezone)
			}
			if !strings.Contains(cfg.LLM.SystemPrompt, "{{CURRENT_DATE}}") {
				t.Error("system prompt must keep the date placeholder until generation")
			}
		})
	}
}
