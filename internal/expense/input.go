package expense

import (
	"time"

	"github.com/google/uuid"
)

// CreateExpenseInput contains data for a new expense.
// Amount is in rubles; the service converts it to minor units (kopecks).
type CreateExpenseInput struct {
	UserID   uuid.UUID
	Amount   int64
	Category string
	SpentAt  time.Time
	Comment  string
}

// UpdateExpenseInput contains the editable expense fields.
// The owner of an expense cannot be changed.
type UpdateExpenseInput struct {
	Amount   int64
	Category string
	SpentAt  time.Time
	Comment  string
}

// ListFilter selects expenses by owner, category and inclusive date range.
// UserID is required for List and optional for GetAll; other zero values skip a filter.
type ListFilter struct {
	UserID   uuid.UUID
	Category string
	DateFrom time.Time
	DateTo   time.Time
}
