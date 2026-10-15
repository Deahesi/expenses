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

// ListFilter selects expenses for the web panel;
// zero values mean the filter is not applied.
type ListFilter struct {
	UserID   uuid.UUID
	Category string
	DateFrom time.Time
	DateTo   time.Time
}
