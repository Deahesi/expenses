package app

import (
	"context"
	"fmt"

	"github.com/Deahesi/expenses-backend/internal/config"
	db "github.com/Deahesi/expenses-backend/internal/storage/postgres/gen"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	pool     *pgxpool.Pool
	server   *Server
	services *Services
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

	repositories := NewRepositories(queries)

	services, err := NewServices(ctx, repositories, cfg)
	if err != nil {
		return nil, fmt.Errorf("services initialization: %w", err)
	}

	handlers := NewHandlers(services)

	server := NewServer(handlers, cfg)

	return &Application{
		pool:     pool,
		server:   server,
		services: services,
	}, nil
}

func (a *Application) Run(ctx context.Context, cfg *config.Config) error {
	a.services.Bot.Run(ctx)
	return a.server.Serve(ctx, cfg)
}

func (a *Application) Close() {
	a.pool.Close()
}
