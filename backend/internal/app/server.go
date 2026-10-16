package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Deahesi/expenses-backend/internal/config"
	"github.com/Deahesi/expenses-backend/internal/httpapi"
)

type Server struct {
	server *http.Server
}

func NewServer(
	services *Services,
	cfg *config.Config,
) *Server {
	api := httpapi.NewServer(services.Expenses, services.Users)

	server := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: api.Handler(),
	}

	return &Server{
		server: server,
	}
}

func (server *Server) Serve(ctx context.Context, cfg *config.Config) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- server.server.ListenAndServe()
	}()

	fmt.Printf("Server serve on %s\n", server.server.Addr)

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err

	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.HTTP.WriteTimeout,
	)
	defer cancel()

	if err := server.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	return nil
}
