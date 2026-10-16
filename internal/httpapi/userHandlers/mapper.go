package userHandlers

import "github.com/Deahesi/expenses-backend/internal/domain"

func newUserListResponse(users []domain.User) []UserResponse {
	result := make([]UserResponse, 0, len(users))
	for _, record := range users {
		result = append(result, UserResponse{
			ID:         record.ID,
			TelegramID: record.TelegramID,
			FirstName:  record.FirstName,
			LastName:   record.LastName,
			CreatedAt:  record.CreatedAt,
			UpdatedAt:  record.UpdatedAt,
		})
	}
	return result
}
