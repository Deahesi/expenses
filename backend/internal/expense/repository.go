package expense

import (
	"context"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/google/uuid"
)

// ExpensesRepository persists expenses.
type ExpensesRepository interface {
	ExpensesWithUserRepository

	Create(ctx context.Context, expense domain.Expense) (domain.Expense, error)
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.Expense, error)
	GetAll(ctx context.Context, filter ListFilter) ([]domain.Expense, error)
	List(ctx context.Context, filter ListFilter) ([]domain.Expense, error)
	Update(ctx context.Context, expense domain.Expense) (domain.Expense, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// ExpensesWithUserRepository reads expenses together with their owners.
type ExpensesWithUserRepository interface {
	GetWithUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.ExpenseWithUser, error)
	GetAllWithUser(ctx context.Context, filter ListFilter) ([]domain.ExpenseWithUser, error)
	ListWithUser(ctx context.Context, filter ListFilter) ([]domain.ExpenseWithUser, error)
}
