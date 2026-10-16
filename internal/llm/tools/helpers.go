package tools

import (
	"errors"
	"fmt"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
)

type LLMExpense struct {
	domain.Expense
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

func domainExpenseToLLMType(expense domain.Expense) LLMExpense {

	return LLMExpense{
		ID:        expense.ID.String(),
		UserID:    expense.UserID.String(),
		Amount:    expense.Amount,
		Category:  expense.Category,
		SpentAt:   expense.SpentAt,
		Comment:   expense.Comment,
		CreatedAt: expense.CreatedAt,
		UpdatedAt: expense.UpdatedAt,
	}
}

func mapExpenses(expenses []domain.Expense) []LLMExpense {
	llmExpenses := make([]LLMExpense, 0, len(expenses))

	for _, expense := range expenses {
		llmExpenses = append(llmExpenses, domainExpenseToLLMType(expense))
	}

	return llmExpenses
}

var (
	ErrEmptyDate = errors.New("empty date")
)

func parseDate(value string, loc *time.Location) (time.Time, error) {
	if value == "" {
		return time.Time{}, ErrEmptyDate
	}

	parsed, err := time.ParseInLocation("2006-01-02", value, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: %w", value, err)
	}

	return parsed, nil
}
