package expenseHandlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func newCreateExpenseInput(req CreateExpenseRequest) (expense.CreateExpenseInput, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Category = domain.ExpenseCategory(strings.TrimSpace(string(req.Category)))
	req.SpentAt = strings.TrimSpace(req.SpentAt)
	req.Comment = strings.TrimSpace(req.Comment)
	if err := req.Validate(); err != nil {
		return expense.CreateExpenseInput{}, err
	}

	spentAt, _ := time.Parse(time.DateOnly, req.SpentAt)
	return expense.CreateExpenseInput{
		UserID:   uuid.MustParse(req.UserID),
		Amount:   req.Amount,
		Category: string(req.Category),
		SpentAt:  spentAt,
		Comment:  req.Comment,
	}, nil
}

func newUpdateExpenseInput(req UpdateExpenseRequest) (expense.UpdateExpenseInput, error) {
	req.Category = domain.ExpenseCategory(strings.TrimSpace(string(req.Category)))
	req.SpentAt = strings.TrimSpace(req.SpentAt)
	req.Comment = strings.TrimSpace(req.Comment)
	if err := req.Validate(); err != nil {
		return expense.UpdateExpenseInput{}, err
	}

	spentAt, _ := time.Parse(time.DateOnly, req.SpentAt)
	return expense.UpdateExpenseInput{
		Amount:   req.Amount,
		Category: string(req.Category),
		SpentAt:  spentAt,
		Comment:  req.Comment,
	}, nil
}

func parseListFilter(r *http.Request) (expense.ListFilter, error) {
	query := r.URL.Query()
	req := GetAllExpensesRequest{
		UserID:   strings.TrimSpace(query.Get("user_id")),
		Category: domain.ExpenseCategory(strings.TrimSpace(query.Get("category"))),
		DateFrom: strings.TrimSpace(query.Get("date_from")),
		DateTo:   strings.TrimSpace(query.Get("date_to")),
	}
	if userID := chi.URLParam(r, "user_id"); userID != "" {
		req.UserID = userID
	}
	if err := req.Validate(); err != nil {
		return expense.ListFilter{}, err
	}

	filter := expense.ListFilter{Category: string(req.Category)}
	if req.UserID != "" {
		filter.UserID = uuid.MustParse(req.UserID)
	}
	filter.DateFrom, _ = time.Parse(time.DateOnly, req.DateFrom)
	filter.DateTo, _ = time.Parse(time.DateOnly, req.DateTo)
	return filter, nil
}

func newExpenseResponse(record domain.Expense) ExpenseResponse {
	return ExpenseResponse{
		ID:        record.ID,
		UserID:    record.UserID,
		Amount:    float64(record.Amount),
		Category:  string(record.Category),
		SpentAt:   record.SpentAt.Format(time.DateOnly),
		Comment:   record.Comment,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func newExpenseWithUserResponse(record domain.ExpenseWithUser) ExpenseWithUserResponse {
	return ExpenseWithUserResponse{
		ExpenseResponse: newExpenseResponse(record.Expense),
		User: ExpenseUserResponse{
			ID:         record.User.ID,
			TelegramID: record.User.TelegramID,
			FirstName:  record.User.FirstName,
			LastName:   record.User.LastName,
		},
	}
}

func newExpenseWithUserListResponse(expenses []domain.ExpenseWithUser) []ExpenseWithUserResponse {
	result := make([]ExpenseWithUserResponse, 0, len(expenses))
	for _, record := range expenses {
		result = append(result, newExpenseWithUserResponse(record))
	}
	return result
}
