package expense

import (
	"context"
	"fmt"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/google/uuid"
)

func (s *ExpenseService) GetWithUser(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) (domain.ExpenseWithUser, error) {
	if id == uuid.Nil || userID == uuid.Nil {
		return domain.ExpenseWithUser{}, fmt.Errorf("%w: expense id and user_id are required", domain.ErrInvalidInput)
	}

	record, err := s.repo.GetWithUser(ctx, id, userID)
	if err != nil {
		return domain.ExpenseWithUser{}, fmt.Errorf("get expense with user: %w", err)
	}

	return record, nil
}

func (s *ExpenseService) GetAllWithUser(
	ctx context.Context,
	filter ListFilter,
) ([]domain.ExpenseWithUser, error) {
	filter, err := normalizeListFilter(filter)
	if err != nil {
		return nil, err
	}

	records, err := s.repo.GetAllWithUser(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get all expenses with users: %w", err)
	}

	return records, nil
}

func (s *ExpenseService) ListWithUser(
	ctx context.Context,
	filter ListFilter,
) ([]domain.ExpenseWithUser, error) {
	if filter.UserID == uuid.Nil {
		return nil, fmt.Errorf("%w: user_id is required", domain.ErrInvalidInput)
	}

	filter, err := normalizeListFilter(filter)
	if err != nil {
		return nil, err
	}

	records, err := s.repo.ListWithUser(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list expenses with user: %w", err)
	}

	return records, nil
}
