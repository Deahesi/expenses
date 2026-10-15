package expense

import (
	"context"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/google/uuid"
)

// ExpensesRepository persists expenses.
type ExpensesRepository interface {
	Create(ctx context.Context, expense domain.Expense) (domain.Expense, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Expense, error)
	List(ctx context.Context, filter ListFilter) ([]domain.Expense, error)
	Update(ctx context.Context, expense domain.Expense) (domain.Expense, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
