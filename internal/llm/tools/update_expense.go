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
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Expense LLMExpense `json:"expense"`
}

func NewEditExpenseTool(
	service *expense.ExpenseService,
	userID uuid.UUID,
) *ai.ToolAction[EditExpenseInput, EditExpenseOutput] {

	return ai.NewTool(
		"edit_expense",
		`Replace the field values of one existing expense belonging to the current user.
		First retrieve the target with list_expenses. Use only its returned ID.
		If the user's criteria do not identify exactly one target, ask which record.

		Required input:
		- expense_id: Exact UUID returned by list_expenses for the identified record.
		- amount: Positive integer in whole RUB.
		- category: Exactly one of еда, транспорт, жильё, развлечения, здоровье, покупки, другое.
		- spent_at: YYYY-MM-DD in the user's timezone.
		- comment: Expense description.

		All fields are required: this is a full replacement, not a partial patch.
		Change only the fields requested by the user. Copy every unchanged field
		exactly from the retrieved record; do not use today's date for an unchanged date.
		Never use this tool to create a new record.

		Output: success, message, expense. Confirm only if success=true;
		then expense contains the updated values. On failure, do not claim an update.`,

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

			expense, err := service.Update(ctx, id, userID, expense.UpdateExpenseInput{
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
