package tools

import (
	"log/slog"

	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
)

type TotalExpensesOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Total   int64  `json:"total"`
}

func NewTotalExpensesTool(
	service *expense.ExpenseService,
	userID uuid.UUID,
) *ai.ToolAction[ListExpensesInput, TotalExpensesOutput] {
	return ai.NewTool(
		"total_expenses",
		`Calculate total spending belonging to the current user in whole RUB.
		Use for questions such as "сколько потратил на еду в этом месяце".

		Input: category, date_from, date_to, with an empty string for each
		unspecified filter. Nonempty dates use YYYY-MM-DD; both bounds are inclusive.
		category is empty or one of еда, транспорт, жильё, развлечения, здоровье, покупки, другое.
		Filters combine with AND. The application supplies the current user's identity.

		Output: success, message, total. Report total only on success=true.
		A successful total of zero means no matching spending. On success=false,
		explain the failure without treating total=0 as a spending result.`,
		func(ctx *ai.ToolContext, input ListExpensesInput) (TotalExpensesOutput, error) {
			slog.Log(ctx, slog.LevelInfo, "Tool total\n")

			filter, err := parseExpenseFilter(input, userID)
			if err != nil {
				return TotalExpensesOutput{Message: err.Error()}, nil
			}

			total, err := service.Total(ctx, filter)
			if err != nil {
				return TotalExpensesOutput{Message: err.Error()}, nil
			}

			return TotalExpensesOutput{Success: true, Total: total}, nil
		},
	)
}
