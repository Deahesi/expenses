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
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewCreateExpenseTool(
	service *expense.ExpenseService,
	userID uuid.UUID,
) *ai.ToolAction[CreateExpenseInput, CreateExpenseOutput] {

	return ai.NewTool(
		"create_expense",
		`Create a new expense for the current user.

		Use this tool when the user explicitly wants to record a purchase, payment, or expense.

		Required information:
		- amount: Positive expense amount in RUB.
		- category: Expense category inferred from the description.
		- spent_at: Expense date in RFC3339 format.
		- comment: Short description of the expense.

		Rules:
		- Use today's date if the user does not specify one.
		- Infer the most appropriate category automatically.
		- Ask for clarification if the amount is missing or ambiguous.
		- Do not create expenses when the user only asks a question.
		- Never claim success unless the tool confirms it.

		Example: "кофе 350 вчера" means creating an expense of 350 RUB
		for coffee, dated yesterday.`,

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

			_, err = service.Create(ctx, expense.CreateExpenseInput{
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
			}, nil
		},
	)
}
