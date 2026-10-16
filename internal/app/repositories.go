package app

import (
	"context"

	"github.com/Deahesi/expenses-backend/internal/config"
	"github.com/Deahesi/expenses-backend/internal/storage/postgres"
	db "github.com/Deahesi/expenses-backend/internal/storage/postgres/gen"
	"github.com/Deahesi/expenses-backend/internal/storage/redis"
)

type Repositories struct {
	Users    *postgres.UserRepository
	Expenses *postgres.ExpensesRepository

	RedisStorage *redis.RedisStorage
}

func NewRepositories(
	ctx context.Context,
	cfg *config.Config,
	queries *db.Queries,
) (*Repositories, error) {
	redisStorage, err := redis.NewRedisStorage(ctx, &cfg.Redis)

	if err != nil {
		return nil, err
	}

	return &Repositories{
		Users:        postgres.NewUserRepository(queries),
		Expenses:     postgres.NewExpensesRepository(queries),
		RedisStorage: redisStorage,
	}, nil
}

func (r *Repositories) Close() error {
	return r.RedisStorage.Close()
}
