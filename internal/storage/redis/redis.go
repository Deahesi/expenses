package redis

import (
	"context"
	"fmt"

	"github.com/Deahesi/expenses-backend/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisRepositories struct {
	Message *MessagesRepository
}

type RedisStorage struct {
	client       *redis.Client
	Repositories RedisRepositories
}

func NewRedisStorage(ctx context.Context, cfg *config.RedisConfig) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	err := rdb.Ping(ctx)
	if err.Err() != nil {
		return nil, fmt.Errorf("redis initialization: %w", err.Err())
	}

	return &RedisStorage{
		client:       rdb,
		Repositories: InitRepos(rdb),
	}, nil
}

func (s *RedisStorage) Close() error {
	return s.client.Close()
}

func InitRepos(rdb *redis.Client) RedisRepositories {

	return RedisRepositories{
		Message: NewMessagesRepository(rdb),
	}
}
