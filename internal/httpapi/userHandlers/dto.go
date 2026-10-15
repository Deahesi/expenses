package userHandlers

import (
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/user"
	"github.com/google/uuid"
)

type CreateUserRequest struct {
	TelegramID int64  `json:"telegram_id,omitempty"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
}

func (r CreateUserRequest) toInput() user.CreateUserInput {
	return user.CreateUserInput{
		TelegramID: r.TelegramID,
		FirstName:  r.FirstName,
		LastName:   r.LastName,
	}
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (r UpdateUserRequest) toInput() user.UpdateUserInput {
	return user.UpdateUserInput{
		FirstName: r.FirstName,
		LastName:  r.LastName,
	}
}

type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	TelegramID int64     `json:"telegram_id,omitempty"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func newUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

func newUserListResponse(users []domain.User) []UserResponse {
	result := make([]UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, newUserResponse(user))
	}

	return result
}
