package tools_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm/tools"
	"github.com/google/uuid"
)

func TestCreateExpenseTool(t *testing.T) {
	owner := uuid.New()

	tests := []struct {
		name       string
		input      tools.CreateExpenseInput
		wantMsg    string
		wantStored bool
	}{
		{
			name: "creates an expense for the tool user",
			input: tools.CreateExpenseInput{
				Amount:   350,
				Category: "еда",
				SpentAt:  "2026-10-15",
				Comment:  "кофе",
			},
			wantMsg:    "",
			wantStored: true,
		},
		{
			name: "unknown category",
			input: tools.CreateExpenseInput{
				Amount:   350,
				Category: "путешествия",
				SpentAt:  "2026-10-15",
				Comment:  "поездка",
			},
			wantMsg:    `unknown category "путешествия"`,
			wantStored: false,
		},
		{
			name: "invalid date",
			input: tools.CreateExpenseInput{
				Amount:   350,
				Category: "еда",
				SpentAt:  "вчера",
				Comment:  "кофе",
			},
			wantMsg:    `invalid date "вчера"`,
			wantStored: false,
		},
		{
			name: "negative amount",
			input: tools.CreateExpenseInput{
				Amount:   -1,
				Category: "еда",
				SpentAt:  "2026-10-15",
				Comment:  "кофе",
			},
			wantMsg:    "amount must be a positive number",
			wantStored: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeExpenseRepository()
			service := expense.NewExpenseService(repo)

			tool := tools.NewCreateExpenseTool(service, owner)
			got, err := runTool(t, tool, tt.input)
			if err != nil {
				t.Fatalf("RunRaw() error = %v", err)
			}

			if got.Success != tt.wantStored {
				t.Errorf("Success = %v, want %v (message = %q)", got.Success, tt.wantStored, got.Message)
			}

			if tt.wantMsg == "" {
				if got.Message != "" {
					t.Errorf("Message = %q, want an empty message", got.Message)
				}
			} else if !strings.Contains(got.Message, tt.wantMsg) {
				t.Errorf("Message = %q, want it to contain %q", got.Message, tt.wantMsg)
			}

			if tt.wantStored {
				expenses, err := service.List(context.Background(), expense.ListFilter{UserID: owner})
				if err != nil {
					t.Fatalf("List() error = %v", err)
				}

				if len(expenses) != 1 {
					t.Fatalf("List() = %d expenses, want 1", len(expenses))
				}

				stored := expenses[0]
				if stored.UserID != owner {
					t.Errorf("stored user_id = %v, want %v", stored.UserID, owner)
				}

				if stored.Amount != tt.input.Amount {
					t.Errorf("stored amount = %d, want %d", stored.Amount, tt.input.Amount)
				}
			} else if len(repo.expenses) != 0 {
				t.Errorf("failed create persisted %d expenses, want 0", len(repo.expenses))
			}
		})
	}
}
