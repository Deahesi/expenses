package expenseHandlers

import (
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/google/uuid"
)

func TestExpenseDTOUsesWholeRubles(t *testing.T) {
	request := CreateExpenseRequest{
		UserID:   uuid.NewString(),
		Amount:   350,
		Category: "еда",
		SpentAt:  "2026-10-16",
		Comment:  "Кофе",
	}
	input, err := newCreateExpenseInput(request)
	if err != nil {
		t.Fatal(err)
	}
	response := newExpenseResponse(domain.Expense{
		ID:       uuid.New(),
		UserID:   input.UserID,
		Amount:   input.Amount,
		Category: domain.CategoryFood,
		SpentAt:  input.SpentAt,
		Comment:  input.Comment,
	})
	if response.Amount != 350 || response.UserID.String() != request.UserID {
		t.Fatalf("expense DTO changed the amount or owner: %+v", response)
	}
	if response.SpentAt != request.SpentAt || !input.SpentAt.Equal(time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expense DTO changed the calendar date: %+v", response)
	}
}
