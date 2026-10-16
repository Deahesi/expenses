package messages

import (
	"context"

	"github.com/Deahesi/expenses-backend/internal/domain"
)

type MessagesRepository interface {
	AddMessage(ctx context.Context, message *domain.Message) error

	GetMessages(
		ctx context.Context,
		telegramID int64,
	) ([]domain.Message, error)
}
