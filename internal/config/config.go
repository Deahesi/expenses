package config

import (
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func Load() (*Config, error) {
	timezone, err := time.LoadLocation(getEnv("LLM_TIMEZONE", "UTC"))
	if err != nil {
		return nil, fmt.Errorf("LLM_TIMEZONE: %w", err)
	}

	cfg := &Config{
		HTTP: HTTPConfig{
			Addr:         getEnv("HTTP_ADDR", ":8080"),
			ReadTimeout:  getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
		},

		Postgres: PostgresConfig{
			URL:             os.Getenv("DATABASE_URL"),
			MaxConns:        int32(getEnvInt("PG_MAX_CONNS", 20)),
			MinConns:        int32(getEnvInt("PG_MIN_CONNS", 2)),
			ConnMaxLifetime: getEnvDuration("PG_CONN_MAX_LIFETIME", time.Hour),
		},

		LLM: LLMConfig{
			OpenRouterAPIKey: os.Getenv("OPENROUTER_API_KEY"),
			Model:            getEnv("LLM_MODEL", "openrouter/openai/gpt-4o-mini"),
			SystemPrompt:     systemPrompt,
			Timezone:         timezone,
		},
		Telegram: TelegramConfig{
			TelegramBotApiKey: os.Getenv("TELEGRAM_BOT_API_KEY"),
		},
		Redis: RedisConfig{
			RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		},
	}

	if cfg.Postgres.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.Telegram.TelegramBotApiKey == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_API_KEY is required")
	}

	if strings.TrimSpace(cfg.LLM.OpenRouterAPIKey) == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return def
}

func getEnvInt(key string, def int) int {
	value := os.Getenv(key)
	if value == "" {
		return def
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return def
	}

	return n
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return def
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return def
	}

	return duration
}

func getEnvBool(key string, def bool) bool {
	value := os.Getenv(key)
	if value == "true" {
		return true
	} else if value == "false" {
		return false
	} else {
		return def
	}
}

//go:embed llm_prompt.md
var systemPrompt string
