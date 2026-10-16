package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Deahesi/expenses-backend/internal/config"
	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm/tools"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/openrouter"
	"github.com/google/uuid"
)

type LLMService struct {
	genkit *genkit.Genkit

	expenseService *expense.ExpenseService

	systemPrompt string
	timezone     *time.Location
	flow         *core.Flow[*SendMessageInput, string, struct{}]
}

func NewLLMService(
	ctx context.Context,
	expenseService *expense.ExpenseService,
	cfg config.LLMConfig,
) (*LLMService, error) {
	if strings.TrimSpace(cfg.OpenRouterAPIKey) == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, fmt.Errorf("LLM_MODEL is required")
	}
	if cfg.Timezone == nil {
		cfg.Timezone = time.UTC
	}

	g := genkit.Init(
		ctx,
		genkit.WithPlugins(
			&openrouter.OpenRouter{
				APIKey:  cfg.OpenRouterAPIKey,
				AppName: "Expenses Bot",
			},
		),
		genkit.WithDefaultModel(openrouter.ModelRef(cfg.Model, nil).Name()),
	)

	s := &LLMService{
		genkit:         g,
		systemPrompt:   cfg.SystemPrompt,
		timezone:       cfg.Timezone,
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
	if input == nil || strings.TrimSpace(input.Text) == "" {
		return "", fmt.Errorf("%w: message text is required", domain.ErrInvalidInput)
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil || userID == uuid.Nil {
		return "", fmt.Errorf("%w: valid user_id is required", domain.ErrInvalidInput)
	}

	history := domainToGenkitMessages(input.Messages)
	systemPrompt := buildSystemPrompt(s.systemPrompt, time.Now().In(s.timezone))

	expenseTools := []ai.ToolRef{
		tools.NewListExpensesTool(s.expenseService, userID),
		tools.NewTotalExpensesTool(s.expenseService, userID),
		tools.NewCreateExpenseTool(s.expenseService, userID),
		tools.NewEditExpenseTool(s.expenseService, userID),
		tools.NewDeleteExpenseTool(s.expenseService, userID),
	}

	result, err := genkit.Generate(
		ctx,
		s.genkit,
		ai.WithConfig(&openrouter.ChatConfig{
			MaxOutputTokens: 2048,
		}),
		ai.WithSystem(systemPrompt),
		ai.WithPrompt(input.Text),
		ai.WithMessages(history...),
		ai.WithTools(expenseTools...),
		ai.WithToolChoice(ai.ToolChoiceAuto),
	)

	if err != nil {
		return "", fmt.Errorf("generate message: %w", err)
	}

	if result == nil {
		return "", fmt.Errorf("generate message: empty response")
	}

	text := result.Text()
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("generate message: empty response")
	}

	return text, nil
}

func buildSystemPrompt(template string, now time.Time) string {
	return strings.NewReplacer(
		"{{CURRENT_DATE}}", now.Format("2006-01-02"),
		"{{TIMEZONE}}", now.Location().String(),
	).Replace(template)
}

func domainToGenkitMessages(messages []domain.Message) []*ai.Message {
	result := make([]*ai.Message, 0, len(messages))

	for _, msg := range messages {
		if strings.TrimSpace(msg.Message) == "" {
			continue
		}

		switch msg.Role {
		case domain.MessageApp:
			result = append(result, ai.NewModelTextMessage(msg.Message))

		case domain.MessageUser:
			result = append(result, ai.NewUserTextMessage(msg.Message))
		}
	}

	return result
}
