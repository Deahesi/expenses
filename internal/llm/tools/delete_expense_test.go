package tools_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm/tools"
	"github.com/google/uuid"
)

func TestDeleteExpenseTool(t *testing.T) {
	owner := uuid.New()

	tests := []struct {
		name      string
		toolUser  uuid.UUID
		expenseID func(created domain.Expense) string
		wantMsg   string
		wantGone  bool
	}{
		{
			name:      "owner deletes own expense",
			toolUser:  owner,
			expenseID: func(created domain.Expense) string { return created.ID.String() },
			wantMsg:   "",
			wantGone:  true,
		},
		{
			name:      "another user cannot delete",
			toolUser:  uuid.New(),
			expenseID: func(created domain.Expense) string { return created.ID.String() },
			wantMsg:   domain.ErrExpenseNotFound.Error(),
			wantGone:  false,
		},
		{
			name:      "invalid expense id",
			toolUser:  owner,
			expenseID: func(domain.Expense) string { return "not-a-uuid" },
			wantMsg:   "invalid UUID length",
			wantGone:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeExpenseRepository()
			service := expense.NewExpenseService(repo)
			created := seedExpense(t, service, owner, 350)

			tool := tools.NewDeleteExpenseTool(service, tt.toolUser)
			got, err := runTool(t, tool, tools.DeleteExpenseInput{ExpenseID: tt.expenseID(created)})
			if err != nil {
				t.Fatalf("RunRaw() error = %v", err)
			}

			assertOutput(t, got, tt.wantGone, tt.wantMsg)

			_, getErr := service.Get(context.Background(), created.ID, owner)
			if tt.wantGone && !errors.Is(getErr, domain.ErrExpenseNotFound) {
				t.Errorf("expense still exists after delete: %v", getErr)
			}

			if !tt.wantGone && getErr != nil {
				t.Errorf("expense disappeared without a successful delete: %v", getErr)
			}
		})
	}
}

// assertOutput checks a tool output against the expected success flag and
// message substring ("" means the message must be empty).
func assertOutput(t *testing.T, got tools.DeleteExpenseOutput, wantSuccess bool, wantMsg string) {
	t.Helper()

	if got.Success != wantSuccess {
		t.Errorf("Success = %v, want %v (message = %q)", got.Success, wantSuccess, got.Message)
	}

	if wantMsg == "" {
		if got.Message != "" {
			t.Errorf("Message = %q, want an empty message", got.Message)
		}

		return
	}

	if !strings.Contains(got.Message, wantMsg) {
		t.Errorf("Message = %q, want it to contain %q", got.Message, wantMsg)
	}
}
