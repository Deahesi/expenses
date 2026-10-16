package tools

import (
	"log/slog"
	"time"

	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
)

type CreateExpenseInput struct {
	Amount   int64  `json:"amount"`
	Category string `json:"category"`
	SpentAt  string `json:"spent_at"`
	Comment  string `json:"comment"`
}

type CreateExpenseOutput struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Expense LLMExpense `json:"expense"`
}

func NewCreateExpenseTool(
	service *expense.ExpenseService,
	userID uuid.UUID,
) *ai.ToolAction[CreateExpenseInput, CreateExpenseOutput] {

	return ai.NewTool(
		"create_expense",
		`Create one expense explicitly reported by the current user.
		Call when its amount is known; ask if it is missing or ambiguous.
		Do not use for editing, budgets, or hypothetical spending.

		Required input:
		- amount: Positive integer in whole RUB; 350 means 350 rubles, not kopecks.
		- category: Exactly one of еда, транспорт, жильё, развлечения, здоровье, покупки, другое.
		- spent_at: YYYY-MM-DD in the user's timezone. Use the current local date if omitted by the user.
		- comment: Short Russian description based on the message; use "Расход" if no purpose is known.

		Output: success, message, expense. On success=true, expense contains
		the actual saved values. Confirm creation using only those values.
		On success=false, explain the failure without claiming the expense was saved.`,

		func(
			ctx *ai.ToolContext,
			input CreateExpenseInput,
		) (CreateExpenseOutput, error) {
			slog.Log(ctx, slog.LevelInfo, "Tool create\n")

			spentAt, err := parseDate(input.SpentAt, time.Now().Location())
			if err != nil {
				return CreateExpenseOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}

			created, err := service.Create(ctx, expense.CreateExpenseInput{
				UserID:   userID,
				Category: input.Category,
				Amount:   input.Amount,
				SpentAt:  spentAt,
				Comment:  input.Comment,
			})

			if err != nil {
				return CreateExpenseOutput{
					Success: false,
					Message: err.Error(),
				}, nil
			}

			return CreateExpenseOutput{
				Success: true,
				Expense: domainExpenseToLLMType(created),
			}, nil
		},
	)
}
