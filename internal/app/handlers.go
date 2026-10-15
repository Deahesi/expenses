package app

import (
	expenseHandlers "github.com/Deahesi/expenses-backend/internal/httpapi/expenseHandlers"
	userHandlers "github.com/Deahesi/expenses-backend/internal/httpapi/userHandlers"
)

type Handlers struct {
	Users    *userHandlers.UserHandler
	Expenses *expenseHandlers.ExpenseHandler
}

func NewHandlers(services *Services) *Handlers {
	return &Handlers{
		Users:    userHandlers.NewUserHandler(services.Users),
		Expenses: expenseHandlers.NewExpenseHandler(services.Expenses),
	}
}
