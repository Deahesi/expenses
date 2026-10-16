package messages

import (
	"context"
	"fmt"

	"github.com/Deahesi/expenses-backend/internal/domain"
)

// MessagesService manages the conversation history of a Telegram user;
// the bot uses it to give the LLM the context of the dialogue.
type MessagesService struct {
	repo MessagesRepository
}

func NewMessagesService(repo MessagesRepository) *MessagesService {
	return &MessagesService{repo: repo}
}

// AddMessage appends a message to the user's conversation history.
func (s *MessagesService) AddMessage(ctx context.Context, message *domain.Message) error {
	if err := s.repo.AddMessage(ctx, message); err != nil {
		return fmt.Errorf("add message: %w", err)
	}

	return nil
}

// GetMessages returns the conversation history of a Telegram user.
func (s *MessagesService) GetMessages(ctx context.Context, telegramID int64) ([]domain.Message, error) {
	messages, err := s.repo.GetMessages(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	return messages, nil
}
