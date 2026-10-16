package tools_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm/tools"
	"github.com/google/uuid"
)

func TestEditExpenseTool(t *testing.T) {
	owner := uuid.New()
	validDate := "2026-10-16"

	tests := []struct {
		name      string
		toolUser  uuid.UUID
		input     func(created domain.Expense) tools.EditExpenseInput
		wantMsg   string
		wantSaved *struct {
			amount   int64
			category domain.ExpenseCategory
			comment  string
		}
	}{
		{
			name:     "owner edits own expense",
			toolUser: owner,
			input: func(created domain.Expense) tools.EditExpenseInput {
				return tools.EditExpenseInput{
					ExpenseID: created.ID.String(),
					Amount:    450,
					Category:  "транспорт",
					SpentAt:   validDate,
					Comment:   "такси",
				}
			},
			wantMsg: "",
			wantSaved: &struct {
				amount   int64
				category domain.ExpenseCategory
				comment  string
			}{amount: 450, category: domain.CategoryTransport, comment: "такси"},
		},
		{
			name:     "another user cannot edit",
			toolUser: uuid.New(),
			input: func(created domain.Expense) tools.EditExpenseInput {
				return tools.EditExpenseInput{
					ExpenseID: created.ID.String(),
					Amount:    1,
					Category:  "еда",
					SpentAt:   validDate,
					Comment:   "взлом",
				}
			},
			wantMsg:   domain.ErrExpenseNotFound.Error(),
			wantSaved: nil,
		},
		{
			name:     "invalid spent_at",
			toolUser: owner,
			input: func(created domain.Expense) tools.EditExpenseInput {
				return tools.EditExpenseInput{
					ExpenseID: created.ID.String(),
					Amount:    450,
					Category:  "еда",
					SpentAt:   "16.10.2026",
					Comment:   "кофе",
				}
			},
			wantMsg:   `invalid date "16.10.2026"`,
			wantSaved: nil,
		},
		{
			name:     "invalid expense id",
			toolUser: owner,
			input: func(created domain.Expense) tools.EditExpenseInput {
				return tools.EditExpenseInput{
					ExpenseID: "not-a-uuid",
					Amount:    450,
					Category:  "еда",
					SpentAt:   validDate,
					Comment:   "кофе",
				}
			},
			wantMsg:   "invalid UUID length",
			wantSaved: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeExpenseRepository()
			service := expense.NewExpenseService(repo)
			created := seedExpense(t, service, owner, 350)

			tool := tools.NewEditExpenseTool(service, tt.toolUser)
			got, err := runTool(t, tool, tt.input(created))
			if err != nil {
				t.Fatalf("RunRaw() error = %v", err)
			}

			if got.Success != (tt.wantSaved != nil) {
				t.Errorf("Success = %v, want %v (message = %q)", got.Success, tt.wantSaved != nil, got.Message)
			}

			if tt.wantMsg == "" {
				if got.Message != "" {
					t.Errorf("Message = %q, want an empty message", got.Message)
				}
			} else if !strings.Contains(got.Message, tt.wantMsg) {
				t.Errorf("Message = %q, want it to contain %q", got.Message, tt.wantMsg)
			}

			stored, getErr := service.Get(context.Background(), created.ID, owner)
			if getErr != nil {
				t.Fatalf("Get() error = %v", getErr)
			}

			if tt.wantSaved != nil {
				if got.Expense.Amount != tt.wantSaved.amount {
					t.Errorf("output Expense.Amount = %d, want %d", got.Expense.Amount, tt.wantSaved.amount)
				}

				if stored.Amount != tt.wantSaved.amount {
					t.Errorf("stored amount = %d, want %d", stored.Amount, tt.wantSaved.amount)
				}

				if stored.Category != tt.wantSaved.category {
					t.Errorf("stored category = %q, want %q", stored.Category, tt.wantSaved.category)
				}

				if stored.Comment != tt.wantSaved.comment {
					t.Errorf("stored comment = %q, want %q", stored.Comment, tt.wantSaved.comment)
				}

				if !stored.SpentAt.Equal(time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)) {
					t.Errorf("stored spent_at = %v", stored.SpentAt)
				}
			} else {
				if stored.Amount != 350 || stored.Comment != "кофе" {
					t.Errorf("failed edit modified the record: %+v", stored)
				}
			}
		})
	}
}
