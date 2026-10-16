package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
)

// fakeExpenseRepository is an in-memory expense.ExpensesRepository
// implementation scoped by (id, user_id) like the generated SQL queries.
type fakeExpenseRepository struct {
	expenses map[uuid.UUID]domain.Expense
	listErr  error
}

func newFakeExpenseRepository() *fakeExpenseRepository {
	return &fakeExpenseRepository{expenses: make(map[uuid.UUID]domain.Expense)}
}

func (r *fakeExpenseRepository) Create(_ context.Context, e domain.Expense) (domain.Expense, error) {
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	r.expenses[e.ID] = e

	return e, nil
}

func (r *fakeExpenseRepository) Get(_ context.Context, id, userID uuid.UUID) (domain.Expense, error) {
	e, ok := r.expenses[id]
	if !ok || e.UserID != userID {
		return domain.Expense{}, domain.ErrExpenseNotFound
	}

	return e, nil
}

func (r *fakeExpenseRepository) GetAll(ctx context.Context, filter expense.ListFilter) ([]domain.Expense, error) {
	return r.List(ctx, filter)
}

func (r *fakeExpenseRepository) GetWithUser(context.Context, uuid.UUID, uuid.UUID) (domain.ExpenseWithUser, error) {
	return domain.ExpenseWithUser{}, errors.New("unexpected GetWithUser call")
}

func (r *fakeExpenseRepository) GetAllWithUser(context.Context, expense.ListFilter) ([]domain.ExpenseWithUser, error) {
	return nil, errors.New("unexpected GetAllWithUser call")
}

func (r *fakeExpenseRepository) ListWithUser(context.Context, expense.ListFilter) ([]domain.ExpenseWithUser, error) {
	return nil, errors.New("unexpected ListWithUser call")
}

func (r *fakeExpenseRepository) List(_ context.Context, filter expense.ListFilter) ([]domain.Expense, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}

	expenses := make([]domain.Expense, 0, len(r.expenses))
	for _, e := range r.expenses {
		if filter.UserID != uuid.Nil && e.UserID != filter.UserID {
			continue
		}
		if filter.Category != "" && string(e.Category) != filter.Category {
			continue
		}
		if !filter.DateFrom.IsZero() && e.SpentAt.Before(filter.DateFrom) {
			continue
		}
		if !filter.DateTo.IsZero() && e.SpentAt.After(filter.DateTo) {
			continue
		}
		expenses = append(expenses, e)
	}

	return expenses, nil
}

func (r *fakeExpenseRepository) Update(_ context.Context, e domain.Expense) (domain.Expense, error) {
	existing, ok := r.expenses[e.ID]
	if !ok || existing.UserID != e.UserID {
		return domain.Expense{}, domain.ErrExpenseNotFound
	}

	existing.Amount = e.Amount
	existing.Category = e.Category
	existing.SpentAt = e.SpentAt
	existing.Comment = e.Comment
	existing.UpdatedAt = time.Now()
	r.expenses[e.ID] = existing

	return existing, nil
}

func (r *fakeExpenseRepository) Delete(_ context.Context, id, userID uuid.UUID) error {
	e, ok := r.expenses[id]
	if !ok || e.UserID != userID {
		return domain.ErrExpenseNotFound
	}

	delete(r.expenses, id)

	return nil
}

// runTool invokes a genkit tool action and decodes its JSON output into the
// typed output struct.
func runTool[I any, O any](t *testing.T, tool *ai.ToolAction[I, O], input I) (O, error) {
	t.Helper()

	raw, err := tool.RunRaw(context.Background(), input)
	if err != nil {
		var zero O
		return zero, err
	}

	data, err := json.Marshal(raw)
	if err != nil {
		var zero O
		return zero, err
	}

	var out O
	if err := json.Unmarshal(data, &out); err != nil {
		var zero O
		return zero, err
	}

	return out, nil
}

// seedExpense creates an expense owned by owner through the service.
func seedExpense(t *testing.T, service *expense.ExpenseService, owner uuid.UUID, amount int64) domain.Expense {
	t.Helper()

	created, err := service.Create(context.Background(), expense.CreateExpenseInput{
		UserID:   owner,
		Amount:   amount,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
		Comment:  "кофе",
	})
	if err != nil {
		t.Fatalf("seed Create() error = %v", err)
	}

	return created
}
