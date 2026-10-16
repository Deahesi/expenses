package expenseHandlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/httpapi/requests"
	"github.com/Deahesi/expenses-backend/internal/httpapi/responses"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ExpenseHandler struct {
	service *expense.ExpenseService
}

func NewExpenseHandler(service *expense.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	req, err := requests.DecodeJSON[CreateExpenseRequest](r)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.UserID = chi.URLParam(r, "user_id")
	input, err := newCreateExpenseInput(req)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.service.Create(r.Context(), input)
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusCreated, newExpenseResponse(created))
}

func (h *ExpenseHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	expenses, err := h.service.GetAllWithUser(r.Context(), filter)
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, newExpenseWithUserListResponse(expenses))
}

func (h *ExpenseHandler) GetUserAll(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	expenses, err := h.service.ListWithUser(r.Context(), filter)
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, newExpenseWithUserListResponse(expenses))
}

func (h *ExpenseHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, userID, err := parseExpenseIDs(r)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	expense, err := h.service.GetWithUser(r.Context(), id, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, newExpenseWithUserResponse(expense))
}

func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, userID, err := parseExpenseIDs(r)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	req, err := requests.DecodeJSON[UpdateExpenseRequest](r)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input, err := newUpdateExpenseInput(req)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.service.Update(r.Context(), id, userID, input)
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, newExpenseResponse(updated))
}

func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, userID, err := parseExpenseIDs(r)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseExpenseIDs(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, uuid.Nil, errors.New("invalid expense id")
	}

	userID, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil || userID == uuid.Nil {
		return uuid.Nil, uuid.Nil, errors.New("invalid user id")
	}

	return id, userID, nil
}

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		responses.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrExpenseNotFound), errors.Is(err, domain.ErrUserNotFound):
		responses.Error(w, http.StatusNotFound, err.Error())
	default:
		log.Printf("expense handler: %v", err)
		responses.Error(w, http.StatusInternalServerError, "internal server error")
	}
}
