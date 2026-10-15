package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Expense is a single family expense record.
// Amount is stored in minor currency units (kopecks) to keep money math exact.
type Expense struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Amount    int64           `json:"amount"`
	Category  ExpenseCategory `json:"category"`
	SpentAt   time.Time       `json:"spent_at"`
	Comment   string          `json:"comment"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

var ErrExpenseNotFound = errors.New("expense not found")

// ExpenseCategory is one of the categories an expense can belong to.
type ExpenseCategory string

const (
	CategoryFood          ExpenseCategory = "еда"
	CategoryTransport     ExpenseCategory = "транспорт"
	CategoryHousing       ExpenseCategory = "жильё"
	CategoryEntertainment ExpenseCategory = "развлечения"
	CategoryHealth        ExpenseCategory = "здоровье"
	CategoryShopping      ExpenseCategory = "покупки"
	CategoryOther         ExpenseCategory = "другое"
)

// expenseCategories lists all valid categories in a stable order.
var expenseCategories = []ExpenseCategory{
	CategoryFood,
	CategoryTransport,
	CategoryHousing,
	CategoryEntertainment,
	CategoryHealth,
	CategoryShopping,
	CategoryOther,
}

// ExpenseCategories returns all valid categories; the web panel and the bot
// prompt use it to render the full list.
func ExpenseCategories() []ExpenseCategory {
	categories := make([]ExpenseCategory, len(expenseCategories))
	copy(categories, expenseCategories)

	return categories
}

// ParseExpenseCategory validates a raw category value.
func ParseExpenseCategory(value string) (ExpenseCategory, error) {
	category := ExpenseCategory(strings.TrimSpace(value))
	for _, valid := range expenseCategories {
		if category == valid {
			return category, nil
		}
	}

	return "", fmt.Errorf("%w: unknown category %q", ErrInvalidInput, value)
}
