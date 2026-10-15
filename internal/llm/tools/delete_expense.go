package tools

import (
	"log/slog"

	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
)

type DeleteExpenseInput struct {
	ExpenseID string `json:"expense_id"`
}

type DeleteExpenseOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewDeleteExpenseTool(
	service *expense.ExpenseService,
	userID uuid.UUID,
) *ai.ToolAction[DeleteExpenseInput, DeleteExpenseOutput] {

	return ai.NewTool(
		"delete_expense",
		`Permanently delete an existing expense belonging to the current user.

		Use this tool only when the user explicitly requests
		to delete or remove a recorded expense.

		Required information:
		- expense_id: Exact UUID of the expense to delete.

		Rules:
		- Always use list_expenses to identify the target expense first.
		- Use only an expense_id returned by list_expenses.
		- Never guess or generate expense IDs.
		- Ensure the expense matches the user's request.
		- If multiple expenses match, ask for clarification.
		- Do not delete expenses without a clear user request.
		- Never claim success unless the tool confirms it.

		Example: "удали такси" means finding the matching taxi expense
		and deleting it only if the target is unambiguous.`,

		func(
			ctx *ai.ToolContext,
			input DeleteExpenseInput,
		) (DeleteExpenseOutput, error) {
			slog.Log(ctx, slog.LevelInfo, "Tool delete\n")

			id, err := uuid.Parse(input.ExpenseID)
			if err != nil {
				return DeleteExpenseOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}

			err = service.Delete(ctx, id)
			if err != nil {
				return DeleteExpenseOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}

			return DeleteExpenseOutput{
				Success: true,
			}, nil
		},
	)
}
