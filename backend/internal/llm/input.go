package llm

import "github.com/Deahesi/expenses-backend/internal/domain"

type SendMessageInput struct {
	Text     string           `json:"text"`
	UserID   string           `json:"user_id"`
	Messages []domain.Message `json:"messages,omitempty"`
}

type SendMessageOutput struct {
	Text string `json:"text"`
}
