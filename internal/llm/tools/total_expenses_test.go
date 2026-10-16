package tools_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm/tools"
	"github.com/google/uuid"
)

func TestTotalExpensesTool(t *testing.T) {
	owner := uuid.New()
	for _, tt := range []struct {
		name    string
		input   tools.ListExpensesInput
		repoErr error
		want    int64
		wantMsg string
	}{
		{
			name: "filtered own total",
			input: tools.ListExpensesInput{
				Category: "еда",
				DateFrom: "2026-10-15",
				DateTo:   "2026-10-15",
			},
			want: 800,
		},
		{
			name: "unfiltered own total",
			want: 800,
		},
		{
			name:  "no matches",
			input: tools.ListExpensesInput{Category: "транспорт"},
		},
		{
			name:  "outside period",
			input: tools.ListExpensesInput{DateFrom: "2026-10-16"},
		},
		{
			name:    "invalid date",
			input:   tools.ListExpensesInput{DateFrom: "вчера"},
			wantMsg: "invalid date",
		},
		{
			name: "inverted period",
			input: tools.ListExpensesInput{
				DateFrom: "2026-10-16",
				DateTo:   "2026-10-15",
			},
			wantMsg: "date_from must not be after date_to",
		},
		{
			name:    "invalid category",
			input:   tools.ListExpensesInput{Category: "unknown"},
			wantMsg: "unknown category",
		},
		{
			name:    "repository error",
			repoErr: errors.New("storage unavailable"),
			wantMsg: "storage unavailable",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeExpenseRepository()
			service := expense.NewExpenseService(repo)
			seedExpense(t, service, owner, 350)
			seedExpense(t, service, owner, 450)
			seedExpense(t, service, uuid.New(), 9999)
			repo.listErr = tt.repoErr

			got, err := runTool(t, tools.NewTotalExpensesTool(service, owner), tt.input)
			if err != nil {
				t.Fatalf("RunRaw() error = %v", err)
			}
			if got.Success != (tt.wantMsg == "") {
				t.Errorf("Success = %v, message = %q", got.Success, got.Message)
			}
			if got.Total != tt.want {
				t.Errorf("Total = %d, want %d", got.Total, tt.want)
			}
			if !strings.Contains(got.Message, tt.wantMsg) {
				t.Errorf("Message = %q, want %q", got.Message, tt.wantMsg)
			}
		})
	}
}
