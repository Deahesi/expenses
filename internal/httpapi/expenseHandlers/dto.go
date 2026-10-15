package expenseHandlers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

type CreateExpenseRequest struct {
	UserID   string `json:"user_id"`
	Amount   int64  `json:"amount"`
	Category string `json:"category"`
	SpentAt  string `json:"spent_at"`
	Comment  string `json:"comment,omitempty"`
}

func (r CreateExpenseRequest) toInput() (expense.CreateExpenseInput, error) {
	userID, err := parseUUID(r.UserID)
	if err != nil {
		return expense.CreateExpenseInput{}, fmt.Errorf("user_id: %w", err)
	}

	spentAt, err := parseDate(r.SpentAt)
	if err != nil {
		return expense.CreateExpenseInput{}, fmt.Errorf("spent_at: %w", err)
	}

	return expense.CreateExpenseInput{
		UserID:   userID,
		Amount:   r.Amount,
		Category: r.Category,
		SpentAt:  spentAt,
		Comment:  r.Comment,
	}, nil
}

type UpdateExpenseRequest struct {
	Amount   int64  `json:"amount"`
	Category string `json:"category"`
	SpentAt  string `json:"spent_at"`
	Comment  string `json:"comment,omitempty"`
}

func (r UpdateExpenseRequest) toInput() (expense.UpdateExpenseInput, error) {
	spentAt, err := parseDate(r.SpentAt)
	if err != nil {
		return expense.UpdateExpenseInput{}, fmt.Errorf("spent_at: %w", err)
	}

	return expense.UpdateExpenseInput{
		Amount:   r.Amount,
		Category: r.Category,
		SpentAt:  spentAt,
		Comment:  r.Comment,
	}, nil
}

type ExpenseResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	SpentAt   string    `json:"spent_at"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newExpenseResponse(expense domain.Expense) ExpenseResponse {
	return ExpenseResponse{
		ID:        expense.ID,
		UserID:    expense.UserID,
		Amount:    float64(expense.Amount) / 100,
		Category:  string(expense.Category),
		SpentAt:   expense.SpentAt.Format(dateLayout),
		Comment:   expense.Comment,
		CreatedAt: expense.CreatedAt,
		UpdatedAt: expense.UpdatedAt,
	}
}

func newExpenseListResponse(expenses []domain.Expense) []ExpenseResponse {
	result := make([]ExpenseResponse, 0, len(expenses))
	for _, expense := range expenses {
		result = append(result, newExpenseResponse(expense))
	}

	return result
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return uuid.Nil, errors.New("must be a valid uuid")
	}

	return id, nil
}

func parseDate(value string) (time.Time, error) {
	parsed, err := time.Parse(dateLayout, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("must be in %s format", dateLayout)
	}

	return parsed, nil
}
