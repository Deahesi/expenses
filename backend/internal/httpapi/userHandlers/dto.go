package userHandlers

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
