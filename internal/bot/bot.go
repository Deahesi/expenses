package bot

import (
	"context"

	"github.com/Deahesi/expenses-backend/internal/bot/handlers"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm"
	"github.com/Deahesi/expenses-backend/internal/user"
	telegram "github.com/go-telegram/bot"
)

type BotService struct {
	client *telegram.Bot
}

func NewBotService(
	token string,
	usersService *user.UserService,
	expensesService *expense.ExpenseService,
	llmService *llm.LLMService,
) (*BotService, error) {
	messageHandler := handlers.NewMessageHandler(
		usersService,
		expensesService,
		llmService,
	)

	client, err := telegram.New(
		token,
		telegram.WithDefaultHandler(messageHandler.HandleMessage),
	)
	if err != nil {
		return nil, err
	}

	return &BotService{
		client: client,
	}, nil
}

func (s *BotService) Run(ctx context.Context) {
	go func() {
		s.client.Start(ctx)
	}()
}
