package bot

import (
	"context"

	"github.com/Deahesi/expenses-backend/internal/bot/handlers"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm"
	"github.com/Deahesi/expenses-backend/internal/messages"
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
	messagesService *messages.MessagesService,
) (*BotService, error) {
	messageHandler := handlers.NewMessageHandler(
		usersService,
		expensesService,
		llmService,
		messagesService,
	)

	client, err := telegram.New(
		token,
		telegram.WithDefaultHandler(messageHandler.HandleMessage),
		telegram.WithNotAsyncHandlers(),
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
