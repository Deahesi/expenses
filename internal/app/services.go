package app

import (
	"context"

	"github.com/Deahesi/expenses-backend/internal/bot"
	"github.com/Deahesi/expenses-backend/internal/config"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm"
	"github.com/Deahesi/expenses-backend/internal/messages"
	"github.com/Deahesi/expenses-backend/internal/user"
)

type Services struct {
	Users    *user.UserService
	Expenses *expense.ExpenseService
	Bot      *bot.BotService
}

func NewServices(
	ctx context.Context,
	repositories *Repositories,
	cfg *config.Config,
) (*Services, error) {
	users := user.NewUserService(repositories.Users)
	expenses := expense.NewExpenseService(repositories.Expenses)
	messages := messages.NewMessagesService(repositories.RedisStorage.Repositories.Message)

	llm, err := llm.NewLLMService(
		ctx,
		expenses,
		cfg.LLM,
	)

	if err != nil {
		return nil, err
	}

	bot, err := bot.NewBotService(
		cfg.Telegram.TelegramBotApiKey,
		users,
		expenses,
		llm,
		messages,
	)

	if err != nil {
		return nil, err
	}

	return &Services{
		Users:    users,
		Expenses: expenses,
		Bot:      bot,
	}, nil
}
