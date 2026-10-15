package tools

import (
	"log/slog"
	"time"

	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
)

type EditExpenseInput struct {
	ExpenseID string `json:"expense_id"`
	Amount    int64  `json:"amount"`
	Category  string `json:"category"`
	SpentAt   string `json:"spent_at"`
	Comment   string `json:"comment"`
}

type EditExpenseOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Expense LLMExpense
}

func NewEditExpenseTool(
	service *expense.ExpenseService,
	userID uuid.UUID,
) *ai.ToolAction[EditExpenseInput, EditExpenseOutput] {

	return ai.NewTool(
		"edit_expense",
		`Edit an existing expense belonging to the current user.

		Use this tool when the user wants to change an existing expense.

		Editable fields:
		- amount: Expense amount in RUB.
		- category: Expense category.
		- spent_at: Expense date in RFC3339 format.
		- comment: Expense description.

		Rules:
		- Always use list_expenses to identify the target expense first.
		- Use the exact expense_id returned by list_expenses.
		- Retrieve the current expense values before editing.
		- Preserve all fields that the user did not request to change.
		- Provide complete updated expense data.
		- Never guess an expense ID.
		- If the target expense is ambiguous, ask for clarification.
		- Never claim success unless the tool confirms it.

		Example: "последнее было не 350, а 450" means finding
		the most recent expense, verifying its amount is 350,
		and updating the amount to 450 while preserving other fields.`,

		func(
			ctx *ai.ToolContext,
			input EditExpenseInput,
		) (EditExpenseOutput, error) {
			slog.Log(ctx, slog.LevelInfo, "Tool update\n")

			spentAt, err := parseDate(input.SpentAt, time.Now().Location())
			if err != nil {
				return EditExpenseOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}

			id, err := uuid.Parse(input.ExpenseID)
			if err != nil {
				return EditExpenseOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}

			expense, err := service.Update(ctx, id, expense.UpdateExpenseInput{
				Amount:   input.Amount,
				Category: input.Category,
				SpentAt:  spentAt,
				Comment:  input.Comment,
			})
			if err != nil {
				return EditExpenseOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}

			return EditExpenseOutput{
				Success: true,
				Expense: domainExpenseToLLMType(expense),
			}, nil
		},
	)
}
