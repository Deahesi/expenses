package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/storage/postgres/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// foreignKeyViolation is the PostgreSQL error code for foreign key violations.
const foreignKeyViolation = "23503"

// ExpensesRepository implements expense.ExpensesRepository on top of PostgreSQL.
type ExpensesRepository struct {
	queries *gen.Queries
}

func NewExpensesRepository(queries *gen.Queries) *ExpensesRepository {
	return &ExpensesRepository{queries: queries}
}

func (r *ExpensesRepository) Create(ctx context.Context, record domain.Expense) (domain.Expense, error) {
	created, err := r.queries.CreateExpense(ctx, gen.CreateExpenseParams{
		ID:       record.ID,
		UserID:   record.UserID,
		Amount:   record.Amount,
		Category: string(record.Category),
		SpentAt:  record.SpentAt,
		Comment:  nullableText(record.Comment),
	})
	if err != nil {
		return domain.Expense{}, mapExpenseError(err)
	}

	return toDomainExpense(created), nil
}

func (r *ExpensesRepository) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.Expense, error) {
	row, err := r.queries.GetExpense(ctx, gen.GetExpenseParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return domain.Expense{}, mapExpenseError(err)
	}

	return toDomainExpense(row), nil
}

func (r *ExpensesRepository) GetAll(ctx context.Context, filter expense.ListFilter) ([]domain.Expense, error) {
	rows, err := r.queries.GetAllExpenses(ctx, gen.GetAllExpensesParams{
		UserID:   nullableUUID(filter.UserID),
		Category: nullableText(filter.Category),
		DateFrom: nullableDate(filter.DateFrom),
		DateTo:   nullableDate(filter.DateTo),
	})
	if err != nil {
		return nil, mapExpenseError(err)
	}

	return toDomainExpenses(rows), nil
}

func (r *ExpensesRepository) List(ctx context.Context, filter expense.ListFilter) ([]domain.Expense, error) {
	rows, err := r.queries.ListExpenses(ctx, gen.ListExpensesParams{
		UserID:   filter.UserID,
		Category: nullableText(filter.Category),
		DateFrom: nullableDate(filter.DateFrom),
		DateTo:   nullableDate(filter.DateTo),
	})
	if err != nil {
		return nil, mapExpenseError(err)
	}

	return toDomainExpenses(rows), nil
}

func (r *ExpensesRepository) GetWithUser(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) (domain.ExpenseWithUser, error) {
	row, err := r.queries.GetExpenseWithUser(ctx, gen.GetExpenseWithUserParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return domain.ExpenseWithUser{}, mapExpenseError(err)
	}

	return toDomainExpenseWithUser(row.Expense, row.User), nil
}

func (r *ExpensesRepository) GetAllWithUser(
	ctx context.Context,
	filter expense.ListFilter,
) ([]domain.ExpenseWithUser, error) {
	rows, err := r.queries.GetAllExpensesWithUser(ctx, gen.GetAllExpensesWithUserParams{
		UserID:   nullableUUID(filter.UserID),
		Category: nullableText(filter.Category),
		DateFrom: nullableDate(filter.DateFrom),
		DateTo:   nullableDate(filter.DateTo),
	})
	if err != nil {
		return nil, mapExpenseError(err)
	}

	expenses := make([]domain.ExpenseWithUser, 0, len(rows))
	for _, row := range rows {
		expenses = append(expenses, toDomainExpenseWithUser(row.Expense, row.User))
	}

	return expenses, nil
}

func (r *ExpensesRepository) ListWithUser(
	ctx context.Context,
	filter expense.ListFilter,
) ([]domain.ExpenseWithUser, error) {
	rows, err := r.queries.ListExpensesWithUser(ctx, gen.ListExpensesWithUserParams{
		UserID:   filter.UserID,
		Category: nullableText(filter.Category),
		DateFrom: nullableDate(filter.DateFrom),
		DateTo:   nullableDate(filter.DateTo),
	})
	if err != nil {
		return nil, mapExpenseError(err)
	}

	expenses := make([]domain.ExpenseWithUser, 0, len(rows))
	for _, row := range rows {
		expenses = append(expenses, toDomainExpenseWithUser(row.Expense, row.User))
	}

	return expenses, nil
}

func (r *ExpensesRepository) Update(ctx context.Context, record domain.Expense) (domain.Expense, error) {
	updated, err := r.queries.UpdateExpense(ctx, gen.UpdateExpenseParams{
		ID:       record.ID,
		UserID:   record.UserID,
		Amount:   record.Amount,
		Category: string(record.Category),
		SpentAt:  record.SpentAt,
		Comment:  nullableText(record.Comment),
	})
	if err != nil {
		return domain.Expense{}, mapExpenseError(err)
	}

	return toDomainExpense(updated), nil
}

func (r *ExpensesRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	rowsAffected, err := r.queries.DeleteExpense(ctx, gen.DeleteExpenseParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return mapExpenseError(err)
	}

	if rowsAffected == 0 {
		return domain.ErrExpenseNotFound
	}

	return nil
}

// nullableUUID maps uuid.Nil (no filter) to SQL NULL.
func nullableUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: id != uuid.Nil}
}

// nullableText maps an empty string to SQL NULL.
func nullableText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

// nullableDate maps a zero time to SQL NULL.
func nullableDate(value time.Time) pgtype.Date {
	return pgtype.Date{Time: value, Valid: !value.IsZero()}
}

func toDomainExpenses(rows []gen.Expense) []domain.Expense {
	expenses := make([]domain.Expense, 0, len(rows))
	for _, row := range rows {
		expenses = append(expenses, toDomainExpense(row))
	}

	return expenses
}

func toDomainExpense(row gen.Expense) domain.Expense {
	return domain.Expense{
		ID:        row.ID,
		UserID:    row.UserID,
		Amount:    row.Amount,
		Category:  domain.ExpenseCategory(row.Category),
		SpentAt:   row.SpentAt,
		Comment:   row.Comment.String,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toDomainExpenseWithUser(record gen.Expense, owner gen.User) domain.ExpenseWithUser {
	return domain.ExpenseWithUser{
		Expense: toDomainExpense(record),
		User:    toDomainUser(&owner),
	}
}

func mapExpenseError(err error) error {
	var pgErr *pgconn.PgError

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.ErrExpenseNotFound
	case errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation:
		return domain.ErrUserNotFound
	default:
		return err
	}
}
