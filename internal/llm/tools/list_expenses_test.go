package tools_test

import (
	"strings"
	"testing"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm/tools"
	"github.com/google/uuid"
)

func TestListExpensesTool(t *testing.T) {
	owner := uuid.New()

	t.Run("returns expenses with string ids", func(t *testing.T) {
		repo := newFakeExpenseRepository()
		service := expense.NewExpenseService(repo)
		created := seedExpense(t, service, owner, 350)
		seedExpense(t, service, uuid.New(), 9999)

		tool := tools.NewListExpensesTool(service, owner)
		got, err := runTool(t, tool, tools.ListExpensesInput{
			Category: "еда",
			DateFrom: "2026-10-01",
			DateTo:   "2026-10-31",
		})
		if err != nil {
			t.Fatalf("RunRaw() error = %v", err)
		}

		if !got.Success {
			t.Fatalf("Success = false, message = %q", got.Message)
		}

		if len(got.Expenses) != 1 {
			t.Fatalf("Expenses = %d items, want 1", len(got.Expenses))
		}

		first := got.Expenses[0]
		if first.ID != created.ID.String() {
			t.Errorf("Expenses[0].ID = %q, want %q", first.ID, created.ID.String())
		}

		if first.Amount != 350 {
			t.Errorf("Expenses[0].Amount = %d, want 350", first.Amount)
		}

		if first.Category != domain.CategoryFood {
			t.Errorf("Expenses[0].Category = %q, want %q", first.Category, domain.CategoryFood)
		}
	})

	t.Run("invalid date filter", func(t *testing.T) {
		repo := newFakeExpenseRepository()
		service := expense.NewExpenseService(repo)
		seedExpense(t, service, owner, 350)

		tool := tools.NewListExpensesTool(service, owner)
		got, err := runTool(t, tool, tools.ListExpensesInput{
			Category: "еда",
			DateFrom: "01.10.2026",
			DateTo:   "31.10.2026",
		})
		if err != nil {
			t.Fatalf("RunRaw() error = %v", err)
		}

		if got.Success {
			t.Error("Success = true, want false for an invalid date filter")
		}

		if !strings.Contains(got.Message, "invalid date") {
			t.Errorf("Message = %q, want it to contain %q", got.Message, "invalid date")
		}
	})
}
