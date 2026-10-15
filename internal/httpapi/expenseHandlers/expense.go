package expenseHandlers

import (
	"errors"
	"fmt"
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

	input, err := req.toInput()
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

	expenses, err := h.service.List(r.Context(), filter)
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, newExpenseListResponse(expenses))
}

func (h *ExpenseHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		responses.Error(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	expense, err := h.service.Get(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, newExpenseResponse(expense))
}

func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		responses.Error(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	req, err := requests.DecodeJSON[UpdateExpenseRequest](r)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input, err := req.toInput()
	if err != nil {
		responses.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		respondError(w, err)
		return
	}

	responses.JSON(w, http.StatusOK, newExpenseResponse(updated))
}

func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		responses.Error(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseListFilter reads the web panel filters: family member, category
// and period (both bounds inclusive).
func parseListFilter(r *http.Request) (expense.ListFilter, error) {
	query := r.URL.Query()

	filter := expense.ListFilter{
		Category: query.Get("category"),
	}

	if raw := query.Get("user_id"); raw != "" {
		userID, err := parseUUID(raw)
		if err != nil {
			return expense.ListFilter{}, fmt.Errorf("user_id: %w", err)
		}

		filter.UserID = userID
	}

	if raw := query.Get("date_from"); raw != "" {
		dateFrom, err := parseDate(raw)
		if err != nil {
			return expense.ListFilter{}, fmt.Errorf("date_from: %w", err)
		}

		filter.DateFrom = dateFrom
	}

	if raw := query.Get("date_to"); raw != "" {
		dateTo, err := parseDate(raw)
		if err != nil {
			return expense.ListFilter{}, fmt.Errorf("date_to: %w", err)
		}

		filter.DateTo = dateTo
	}

	return filter, nil
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
	case errors.Is(err, domain.ErrExpenseNotFound), errors.Is(err, domain.ErrUserNotFound):
		responses.Error(w, http.StatusNotFound, err.Error())
	default:
		log.Printf("expense handler: %v", err)
		responses.Error(w, http.StatusInternalServerError, "internal server error")
	}
}
