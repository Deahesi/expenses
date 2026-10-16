package userHandlers

import (
	"log"
	"net/http"

	"github.com/Deahesi/expenses-backend/internal/httpapi/responses"
	"github.com/Deahesi/expenses-backend/internal/user"
)

type UserHandler struct {
	service *user.UserService
}

func NewUserHandler(service *user.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.List(r.Context())
	if err != nil {
		log.Printf("user handler: %v", err)
		responses.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	responses.JSON(w, http.StatusOK, newUserListResponse(users))
}
