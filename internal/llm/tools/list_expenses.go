package tools

import (
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
		`Retrieve expenses belonging to the current user.

		Use this tool to:
		- Display expenses for a specific period or category.
		- Find expenses before editing or deleting them.
		- Identify the most recent expense.
		- Retrieve expense details for answering user questions.

		Available filters:
		- category: Filter expenses by category.
		- date_from: Start of the date range.
		- date_to: End of the date range.

		Rules:
		- Apply relevant filters based on the user's request.
		- Leave filters unspecified when they are not needed.
		- Use returned expense IDs for edit_expense and delete_expense.
		- Never invent expense IDs or transaction details.
		- If multiple expenses match an ambiguous request, ask for clarification.
		- Do not assume an incomplete list contains all matching expenses.

		Example: "покажи траты на еду за неделю" means retrieving
		expenses in the food category for the requested week.`,

		func(
			ctx *ai.ToolContext,
			input ListExpensesInput,
		) (ListExpensesOutput, error) {
			slog.Log(ctx, slog.LevelInfo, "Tool list\n")

			dateFrom, err := parseDate(input.DateFrom, time.Now().Location())
			if err != nil {
				return ListExpensesOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}
			dateTo, err := parseDate(input.DateTo, time.Now().Location())
			if err != nil {
				return ListExpensesOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}

			expenses, err := service.List(ctx, expense.ListFilter{
				UserID:   userID,
				Category: input.Category,
				DateFrom: dateFrom,
				DateTo:   dateTo,
			})
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
