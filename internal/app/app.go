package app

import (
	"context"
	"fmt"

	"github.com/Deahesi/expenses-backend/internal/config"
	db "github.com/Deahesi/expenses-backend/internal/storage/postgres/gen"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	pool         *pgxpool.Pool
	server       *Server
	services     *Services
	repositories *Repositories
}

func NewApplication(ctx context.Context, cfg *config.Config) (*Application, error) {
	pool, err := pgxpool.New(ctx, cfg.Postgres.URL)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	queries := db.New(pool)

	repositories, err := NewRepositories(ctx, cfg, queries)
	if err != nil {
		return nil, fmt.Errorf("repositories initialization: %w", err)
	}

	services, err := NewServices(ctx, repositories, cfg)
	if err != nil {
		return nil, fmt.Errorf("services initialization: %w", err)
	}

	server := NewServer(services, cfg)

	return &Application{
		pool:         pool,
		server:       server,
		services:     services,
		repositories: repositories,
	}, nil
}

func (a *Application) Run(ctx context.Context, cfg *config.Config) error {
	a.services.Bot.Run(ctx)
	return a.server.Serve(ctx, cfg)
}

func (a *Application) Close() error {
	a.pool.Close()
	return a.repositories.Close()
}
