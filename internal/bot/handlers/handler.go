package handlers

import (
	"context"

	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm"
	"github.com/Deahesi/expenses-backend/internal/user"
)

type MessageProcessor interface {
	Process(
		ctx context.Context,
		userID int64,
		text string,
	) (string, error)
}

type MessageHandler struct {
	usersService    *user.UserService
	expensesService *expense.ExpenseService
	llmService      *llm.LLMService
}

func NewMessageHandler(
	usersService *user.UserService,
	expensesService *expense.ExpenseService,
	llmService *llm.LLMService,
) *MessageHandler {
	return &MessageHandler{
		usersService:    usersService,
		expensesService: expensesService,
		llmService:      llmService,
	}
}
