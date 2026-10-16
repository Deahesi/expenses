package expenseHandlers

import (
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/invopop/validation"
	"github.com/invopop/validation/is"
)

type CreateExpenseRequest struct {
	UserID   string                 `json:"user_id"`
	Amount   int64                  `json:"amount"`
	Category domain.ExpenseCategory `json:"category"`
	SpentAt  string                 `json:"spent_at"`
	Comment  string                 `json:"comment,omitempty"`
}

func (r CreateExpenseRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UserID, validation.Required, is.UUID, validation.NotIn(uuid.Nil.String())),
		validation.Field(&r.Amount, validation.Required, validation.Min(int64(1)), validation.Max(int64(1e10)).Exclusive()),
		validation.Field(&r.Category, validation.Required, validation.In(domain.ExpenseCategories()...)),
		validation.Field(&r.SpentAt, validation.Required, validation.Date(time.DateOnly)),
		validation.Field(&r.Comment, validation.RuneLength(0, 500)),
	)
}

type UpdateExpenseRequest struct {
	Amount   int64                  `json:"amount"`
	Category domain.ExpenseCategory `json:"category"`
	SpentAt  string                 `json:"spent_at"`
	Comment  string                 `json:"comment,omitempty"`
}

func (r UpdateExpenseRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Amount, validation.Required, validation.Min(int64(1)), validation.Max(int64(1e10)).Exclusive()),
		validation.Field(&r.Category, validation.Required, validation.In(domain.ExpenseCategories()...)),
		validation.Field(&r.SpentAt, validation.Required, validation.Date(time.DateOnly)),
		validation.Field(&r.Comment, validation.RuneLength(0, 500)),
	)
}

type GetAllExpensesRequest struct {
	UserID   string                 `json:"user_id"`
	Category domain.ExpenseCategory `json:"category"`
	DateFrom string                 `json:"date_from"`
	DateTo   string                 `json:"date_to"`
}

func (r GetAllExpensesRequest) Validate() error {
	dateFrom, _ := time.Parse(time.DateOnly, r.DateFrom)

	return validation.ValidateStruct(&r,
		validation.Field(&r.UserID, is.UUID, validation.NotIn(uuid.Nil.String())),
		validation.Field(&r.Category, validation.In(domain.ExpenseCategories()...)),
		validation.Field(&r.DateFrom, validation.Date(time.DateOnly)),
		validation.Field(&r.DateTo, validation.Date(time.DateOnly).Min(dateFrom).RangeError("must not be before date_from")),
	)
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

type ExpenseWithUserResponse struct {
	ExpenseResponse
	User ExpenseUserResponse `json:"user"`
}

type ExpenseUserResponse struct {
	ID         uuid.UUID `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
}
