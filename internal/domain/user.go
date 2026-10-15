package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User is a family member known to the system.
// TelegramID is the user's Telegram account ID; 0 means the user
// is not linked to Telegram yet.
type User struct {
	ID         uuid.UUID `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)
