package config

import "time"

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	LLM      LLMConfig
	Telegram TelegramConfig
}
type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type PostgresConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	ConnMaxLifetime time.Duration
}

type LLMConfig struct {
	OpenRouterAPIKey string
	Model            string
	SystemPrompt     string
}

type TelegramConfig struct {
	TelegramBotApiKey string
}
