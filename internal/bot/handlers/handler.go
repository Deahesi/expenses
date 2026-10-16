package handlers

import (
	"context"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm"
	"github.com/Deahesi/expenses-backend/internal/user"
)

type userService interface {
	UpsertTelegramUser(context.Context, user.UpsertTelegramUserInput) (domain.User, error)
}

type llmService interface {
	SendMessage(context.Context, *llm.SendMessageInput) (string, error)
}

type messagesService interface {
	GetMessages(context.Context, int64) ([]domain.Message, error)
	AddMessage(context.Context, *domain.Message) error
}

type MessageHandler struct {
	usersService    userService
	expensesService *expense.ExpenseService
	llmService      llmService
	messagesService messagesService
}

func NewMessageHandler(
	usersService userService,
	expensesService *expense.ExpenseService,
	llmService llmService,
	messagesService messagesService,
) *MessageHandler {
	return &MessageHandler{
		usersService:    usersService,
		expensesService: expensesService,
		llmService:      llmService,
		messagesService: messagesService,
	}
}
