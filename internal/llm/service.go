package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/Deahesi/expenses-backend/internal/config"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm/tools"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/openrouter"
	"github.com/google/uuid"
)

type LLMService struct {
	genkit         *genkit.Genkit
	expenseService *expense.ExpenseService
	systemPrompt   string
	flow           *core.Flow[*SendMessageInput, string, struct{}]
}

func NewLLMService(
	ctx context.Context,
	expenseService *expense.ExpenseService,
	cfg config.LLMConfig,
) (*LLMService, error) {
	g := genkit.Init(
		ctx,
		genkit.WithPlugins(
			&openrouter.OpenRouter{
				APIKey:  cfg.OpenRouterAPIKey,
				AppName: "Expenses Bot",
			},
		),
		genkit.WithDefaultModel(cfg.Model),
	)

	s := &LLMService{
		genkit:         g,
		systemPrompt:   cfg.SystemPrompt,
		expenseService: expenseService,
	}

	s.flow = genkit.DefineFlow(g, "generateExpenseRequest", s.generate)

	return s, nil
}

func (s *LLMService) SendMessage(
	ctx context.Context,
	input *SendMessageInput,
) (string, error) {
	response, err := s.flow.Action.Run(ctx, input, nil)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (s *LLMService) generate(
	ctx context.Context,
	input *SendMessageInput,
) (string, error) {
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return "", fmt.Errorf("generate message: %w", err)
	}

	tools := []ai.ToolRef{
		tools.NewListExpensesTool(s.expenseService, userID),
		tools.NewCreateExpenseTool(s.expenseService, userID),
		tools.NewDeleteExpenseTool(s.expenseService, userID),
		tools.NewEditExpenseTool(s.expenseService, userID),
	}

	result, err := genkit.Generate(
		ctx,
		s.genkit,
		ai.WithConfig(&openrouter.ChatConfig{
			MaxOutputTokens: 2048,
		}),
		ai.WithSystem(s.systemPrompt),
		ai.WithPrompt(input.Text),
		ai.WithTools(tools...),
	)

	if err != nil && !(result != nil && errors.Is(err, ai.ErrGenerationBlocked)) {
		return "", fmt.Errorf("generate message: %w", err)
	}

	return result.Text(), nil
}
