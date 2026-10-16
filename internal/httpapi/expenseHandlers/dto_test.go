package expenseHandlers

import (
	"strings"
	"testing"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/google/uuid"
)

func TestExpenseRequestValidation(t *testing.T) {
	for _, tt := range []struct {
		name     string
		amount   int64
		category domain.ExpenseCategory
		date     string
		comment  string
		wantErr  bool
	}{
		{
			name:     "valid",
			amount:   350,
			category: "еда",
			date:     "2026-10-16",
		},
		{
			name:     "maximum amount",
			amount:   9999999999,
			category: "еда",
			date:     "2026-10-16",
		},
		{
			name:     "zero amount",
			amount:   0,
			category: "еда",
			date:     "2026-10-16",
			wantErr:  true,
		},
		{
			name:     "negative amount",
			amount:   -1,
			category: "еда",
			date:     "2026-10-16",
			wantErr:  true,
		},
		{
			name:     "amount too large",
			amount:   1e10,
			category: "еда",
			date:     "2026-10-16",
			wantErr:  true,
		},
		{
			name:    "missing category",
			amount:  350,
			date:    "2026-10-16",
			wantErr: true,
		},
		{
			name:     "unknown category",
			amount:   350,
			category: "unknown",
			date:     "2026-10-16",
			wantErr:  true,
		},
		{
			name:     "missing date",
			amount:   350,
			category: "еда",
			wantErr:  true,
		},
		{
			name:     "invalid date",
			amount:   350,
			category: "еда",
			date:     "2026-02-30",
			wantErr:  true,
		},
		{
			name:     "comment at limit",
			amount:   350,
			category: "еда",
			date:     "2026-10-16",
			comment:  strings.Repeat("я", 500),
		},
		{
			name:     "comment too long",
			amount:   350,
			category: "еда",
			date:     "2026-10-16",
			comment:  strings.Repeat("я", 501),
			wantErr:  true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			create := CreateExpenseRequest{
				UserID:   uuid.NewString(),
				Amount:   tt.amount,
				Category: tt.category,
				SpentAt:  tt.date,
				Comment:  tt.comment,
			}
			if err := create.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("CreateExpenseRequest.Validate() = %v, want error = %t", err, tt.wantErr)
			}

			update := UpdateExpenseRequest{
				Amount:   tt.amount,
				Category: tt.category,
				SpentAt:  tt.date,
				Comment:  tt.comment,
			}
			if err := update.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("UpdateExpenseRequest.Validate() = %v, want error = %t", err, tt.wantErr)
			}
		})
	}
}

func TestCreateExpenseRequestRequiresUser(t *testing.T) {
	for _, userID := range []string{"", "invalid", uuid.Nil.String()} {
		t.Run(userID, func(t *testing.T) {
			req := CreateExpenseRequest{
				UserID: userID, Amount: 350, Category: "еда", SpentAt: "2026-10-16",
			}
			if err := req.Validate(); err == nil {
				t.Fatal("expected an invalid user_id error")
			}
		})
	}
}

func TestGetAllExpensesRequestValidation(t *testing.T) {
	for _, tt := range []struct {
		name    string
		request GetAllExpensesRequest
		wantErr bool
	}{
		{
			name: "no filters",
		},
		{
			name: "all filters",
			request: GetAllExpensesRequest{
				UserID: uuid.NewString(), Category: "еда", DateFrom: "2026-10-01", DateTo: "2026-10-31",
			},
		},
		{
			name:    "single day",
			request: GetAllExpensesRequest{DateFrom: "2026-10-16", DateTo: "2026-10-16"},
		},
		{
			name:    "end date only",
			request: GetAllExpensesRequest{DateTo: "2026-10-16"},
		},
		{
			name:    "invalid user",
			request: GetAllExpensesRequest{UserID: "invalid"},
			wantErr: true,
		},
		{
			name:    "nil user",
			request: GetAllExpensesRequest{UserID: uuid.Nil.String()},
			wantErr: true,
		},
		{
			name:    "invalid category",
			request: GetAllExpensesRequest{Category: "unknown"},
			wantErr: true,
		},
		{
			name:    "invalid start date",
			request: GetAllExpensesRequest{DateFrom: "yesterday"},
			wantErr: true,
		},
		{
			name:    "invalid end date",
			request: GetAllExpensesRequest{DateTo: "2026-02-30"},
			wantErr: true,
		},
		{
			name:    "inverted period",
			request: GetAllExpensesRequest{DateFrom: "2026-10-31", DateTo: "2026-10-01"},
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.request.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() = %v, want error = %t", err, tt.wantErr)
			}
		})
	}
}
