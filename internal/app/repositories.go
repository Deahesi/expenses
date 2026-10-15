package app

import (
	"github.com/Deahesi/expenses-backend/internal/storage/postgres"
	db "github.com/Deahesi/expenses-backend/internal/storage/postgres/gen"
)

type Repositories struct {
	Users    *postgres.UserRepository
	Expenses *postgres.ExpensesRepository
}

func NewRepositories(queries *db.Queries) *Repositories {
	return &Repositories{
		Users:    postgres.NewUserRepository(queries),
		Expenses: postgres.NewExpensesRepository(queries),
	}
}
