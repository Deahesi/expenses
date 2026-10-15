package expense

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/google/uuid"
)

const (
	maxCommentLength = 500

	// maxAmountKopecks is the upper bound of the NUMERIC(12,2) amount column.
	maxAmountKopecks int64 = 999_999_999_999
	// maxAmountRubles guards the float-to-int conversion before rounding.
	maxAmountRubles = 1e10
)

// Service implements expense management use cases.
type ExpenseService struct {
	repo ExpensesRepository
}

func NewExpenseService(repo ExpensesRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

func (s *ExpenseService) Create(ctx context.Context, input CreateExpenseInput) (domain.Expense, error) {
	if input.UserID == uuid.Nil {
		return domain.Expense{}, fmt.Errorf("%w: user_id is required", domain.ErrInvalidInput)
	}

	err := validateAmount(input.Amount)
	if err != nil {
		return domain.Expense{}, err
	}

	category, err := domain.ParseExpenseCategory(input.Category)
	if err != nil {
		return domain.Expense{}, err
	}

	if input.SpentAt.IsZero() {
		return domain.Expense{}, fmt.Errorf("%w: spent_at is required", domain.ErrInvalidInput)
	}

	comment, err := normalizeComment(input.Comment)
	if err != nil {
		return domain.Expense{}, err
	}

	created, err := s.repo.Create(ctx, domain.Expense{
		ID:       uuid.New(),
		UserID:   input.UserID,
		Amount:   input.Amount,
		Category: category,
		SpentAt:  normalizeDate(input.SpentAt),
		Comment:  comment,
	})
	if err != nil {
		return domain.Expense{}, fmt.Errorf("create expense: %w", err)
	}

	return created, nil
}

func (s *ExpenseService) Get(ctx context.Context, id uuid.UUID) (domain.Expense, error) {
	expense, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Expense{}, fmt.Errorf("get expense: %w", err)
	}

	return expense, nil
}

func (s *ExpenseService) List(ctx context.Context, filter ListFilter) ([]domain.Expense, error) {
	if filter.Category != "" {
		if _, err := domain.ParseExpenseCategory(filter.Category); err != nil {
			return nil, err
		}
	}

	if !filter.DateFrom.IsZero() && !filter.DateTo.IsZero() && filter.DateFrom.After(filter.DateTo) {
		return nil, fmt.Errorf("%w: date_from must not be after date_to", domain.ErrInvalidInput)
	}

	expenses, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}

	return expenses, nil
}

func (s *ExpenseService) Update(ctx context.Context, id uuid.UUID, input UpdateExpenseInput) (domain.Expense, error) {
	err := validateAmount(input.Amount)
	if err != nil {
		return domain.Expense{}, err
	}

	category, err := domain.ParseExpenseCategory(input.Category)
	if err != nil {
		return domain.Expense{}, err
	}

	if input.SpentAt.IsZero() {
		return domain.Expense{}, fmt.Errorf("%w: spent_at is required", domain.ErrInvalidInput)
	}

	comment, err := normalizeComment(input.Comment)
	if err != nil {
		return domain.Expense{}, err
	}

	updated, err := s.repo.Update(ctx, domain.Expense{
		ID:       id,
		Amount:   input.Amount,
		Category: category,
		SpentAt:  normalizeDate(input.SpentAt),
		Comment:  comment,
	})
	if err != nil {
		return domain.Expense{}, fmt.Errorf("update expense: %w", err)
	}

	return updated, nil
}

func (s *ExpenseService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}

	return nil
}

// normalizeAmount converts a ruble amount into minor units (kopecks),
// rounding to two decimal places the same way the NUMERIC(12,2) column does.
func validateAmount(value int64) error {
	if value < 0 || value >= maxAmountRubles {
		return fmt.Errorf("%w: amount must be a positive number", domain.ErrInvalidInput)
	}

	return nil
}

// normalizeComment trims the value and validates its length; an empty
// comment is stored as NULL.
func normalizeComment(value string) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > maxCommentLength {
		return "", fmt.Errorf(
			"%w: comment must be at most %d characters",
			domain.ErrInvalidInput, maxCommentLength,
		)
	}

	return value, nil
}

// normalizeDate anchors the calendar date at UTC midnight so that the DATE
// column value is independent of the source time zone.
func normalizeDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
