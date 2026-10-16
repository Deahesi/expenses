package tools

import (
	"errors"
	"log/slog"
	"time"

	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
)

type ListExpensesInput struct {
	Category string `json:"category"`
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
}

type ListExpensesOutput struct {
	Success  bool         `json:"success"`
	Message  string       `json:"message"`
	Expenses []LLMExpense `json:"expenses"`
}

func NewListExpensesTool(
	service *expense.ExpenseService,
	userID uuid.UUID,
) *ai.ToolAction[ListExpensesInput, ListExpensesOutput] {

	return ai.NewTool(
		"list_expenses",
		`Read expenses belonging to the current user. Use for viewing records
		and for identifying a record before edit_expense or delete_expense.

		Input: category, date_from, date_to. Send an empty string for each
		unspecified filter. Nonempty dates use YYYY-MM-DD; both bounds are
		inclusive calendar dates in the user's timezone. Filters combine with AND.
		category is empty or one of еда, транспорт, жильё, развлечения, здоровье, покупки, другое.
		There are no amount, text-search, sorting, or limit input parameters.

		Output: success, message, expenses. Each expense includes its existing
		ID, amount, category, spent_at, and comment. On success=true with an empty
		array, no records matched. On success=false, retrieval failed: do not
		interpret it as an empty result. Array order is not guaranteed.
		Use returned IDs for mutations. Do not calculate spending totals here.`,

		func(
			ctx *ai.ToolContext,
			input ListExpensesInput,
		) (ListExpensesOutput, error) {
			slog.Log(ctx, slog.LevelInfo, "Tool list\n")

			filter, err := parseExpenseFilter(input, userID)
			if err != nil {
				return ListExpensesOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}

			expenses, err := service.List(ctx, filter)
			if err != nil {
				return ListExpensesOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}

			return ListExpensesOutput{
				Success:  true,
				Expenses: mapExpenses(expenses),
			}, nil
		},
	)
}

func parseExpenseFilter(input ListExpensesInput, userID uuid.UUID) (expense.ListFilter, error) {
	dateFrom, err := parseDate(input.DateFrom, time.UTC)
	if err != nil && !errors.Is(err, ErrEmptyDate) {
		return expense.ListFilter{}, err
	}
	dateTo, err := parseDate(input.DateTo, time.UTC)
	if err != nil && !errors.Is(err, ErrEmptyDate) {
		return expense.ListFilter{}, err
	}

	return expense.ListFilter{
		UserID:   userID,
		Category: input.Category,
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}, nil
}
