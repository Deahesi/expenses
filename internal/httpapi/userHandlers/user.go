package userHandlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/httpapi/requests"
	"github.com/Deahesi/expenses-backend/internal/httpapi/responses"
	"github.com/Deahesi/expenses-backend/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	service *user.UserService
}

func NewUserHandler(service *user.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	req, err := requests.DecodeJSON[CreateUserRequest](r)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.service.Create(r.Context(), req.toInput())
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusCreated, newUserResponse(created))
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.List(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, newUserListResponse(users))
}

func (h *UserHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		responses.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.Get(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, newUserResponse(user))
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		responses.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	req, err := requests.DecodeJSON[UpdateUserRequest](r)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.service.Update(r.Context(), id, req.toInput())
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, newUserResponse(updated))
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		responses.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseID(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, false
	}

	return id, true
}

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		responses.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		responses.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrUserAlreadyExists):
		responses.Error(w, http.StatusConflict, err.Error())
	default:
		log.Printf("user handler: %v", err)
		responses.Error(w, http.StatusInternalServerError, "internal server error")
	}
}
