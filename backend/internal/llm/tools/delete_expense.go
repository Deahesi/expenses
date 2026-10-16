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
		`Delete one existing expense belonging to the current user.
		First retrieve the target with list_expenses and match the user's criteria.
		If no record matches, report that; if several match, ask which one.

		Input: expense_id, the exact UUID returned by list_expenses for the
		identified record. Never guess, generate, or copy an ID from user text.
		Call only when the user explicitly requests deletion.

		Output: success and message. Confirm deletion only if success=true.
		On success=false, explain the failure without claiming the expense was deleted.`,

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

			err = service.Delete(ctx, id, userID)
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
