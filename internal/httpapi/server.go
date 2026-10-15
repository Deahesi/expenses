package httpapi

import (
	"net/http"

	// "github.com/Deahesi/ai-control/internal/httpapi/middlewares"
	"github.com/Deahesi/expenses-backend/internal/httpapi/expenseHandlers"
	"github.com/Deahesi/expenses-backend/internal/httpapi/userHandlers"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	router *chi.Mux
}

func NewServer(
	userHandler *userHandlers.UserHandler,
	expenseHandler *expenseHandlers.ExpenseHandler,
) *Server {
	router := chi.NewRouter()

	// router.Use(middlewares.RecoverJSON())

	router.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.Create)
		r.Get("/", userHandler.GetAll)
		r.Get("/{id}", userHandler.GetOne)
		r.Put("/{id}", userHandler.Update)
		r.Delete("/{id}", userHandler.Delete)
	})

	router.Route("/expenses", func(r chi.Router) {
		r.Post("/", expenseHandler.Create)
		r.Get("/", expenseHandler.GetAll)
		r.Get("/{id}", expenseHandler.GetOne)
		r.Put("/{id}", expenseHandler.Update)
		r.Delete("/{id}", expenseHandler.Delete)
	})

	return &Server{
		router: router,
	}
}

func (s *Server) Handler() http.Handler {
	return s.router
}
