package user

import (
	"context"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/google/uuid"
)

// Repository persists users.
type UsersRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	Get(ctx context.Context, id uuid.UUID) (domain.User, error)
	Upsert(
		ctx context.Context,
		user domain.User,
	) (domain.User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, user domain.User) (domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
