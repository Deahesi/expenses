package httpapi

import (
	"net/http"

	// "github.com/Deahesi/ai-control/internal/httpapi/middlewares"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/httpapi/expenseHandlers"
	"github.com/Deahesi/expenses-backend/internal/httpapi/userHandlers"
	"github.com/Deahesi/expenses-backend/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type Server struct {
	router *chi.Mux
}

func NewServer(
	expenseService *expense.ExpenseService,
	userService *user.UserService,
) *Server {
	router := chi.NewRouter()

	expenseHandler := expenseHandlers.NewExpenseHandler(expenseService)
	userHandler := userHandlers.NewUserHandler(userService)

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept", "Authorization", "Content-Type",
		},
		MaxAge: 300,
	}))

	router.Get("/users", userHandler.GetAll)
	router.Get("/users/", userHandler.GetAll)

	router.Route("/expenses", func(r chi.Router) {
		r.Post("/{user_id}", expenseHandler.Create)
		r.Get("/", expenseHandler.GetAll)
		r.Get("/{user_id}", expenseHandler.GetUserAll)
		r.Get("/{user_id}/{id}", expenseHandler.GetOne)
		r.Put("/{user_id}/{id}", expenseHandler.Update)
		r.Delete("/{user_id}/{id}", expenseHandler.Delete)
	})

	return &Server{
		router: router,
	}
}

func (s *Server) Handler() http.Handler {
	return s.router
}
