package postgres

import (
	"context"
	"errors"
	"fmt"
	"math/big"
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
		Amount:   numericFromKopecks(record.Amount),
		Category: string(record.Category),
		SpentAt:  nullableDate(record.SpentAt),
		Comment:  nullableText(record.Comment),
	})
	if err != nil {
		return domain.Expense{}, mapExpenseError(err)
	}

	return toDomainExpense(created)
}

func (r *ExpensesRepository) Get(ctx context.Context, id uuid.UUID) (domain.Expense, error) {
	row, err := r.queries.GetExpense(ctx, id)
	if err != nil {
		return domain.Expense{}, mapExpenseError(err)
	}

	return toDomainExpense(row)
}

func (r *ExpensesRepository) List(ctx context.Context, filter expense.ListFilter) ([]domain.Expense, error) {
	rows, err := r.queries.ListExpenses(ctx, gen.ListExpensesParams{
		UserID:   nullableUUID(filter.UserID),
		Category: nullableText(string(filter.Category)),
		DateFrom: nullableDate(filter.DateFrom),
		DateTo:   nullableDate(filter.DateTo),
	})
	if err != nil {
		return nil, mapExpenseError(err)
	}

	expenses := make([]domain.Expense, 0, len(rows))
	for _, row := range rows {
		record, err := toDomainExpense(row)
		if err != nil {
			return nil, err
		}

		expenses = append(expenses, record)
	}

	return expenses, nil
}

func (r *ExpensesRepository) Update(ctx context.Context, record domain.Expense) (domain.Expense, error) {
	updated, err := r.queries.UpdateExpense(ctx, gen.UpdateExpenseParams{
		ID:       record.ID,
		Amount:   numericFromKopecks(record.Amount),
		Category: string(record.Category),
		SpentAt:  nullableDate(record.SpentAt),
		Comment:  nullableText(record.Comment),
	})
	if err != nil {
		return domain.Expense{}, mapExpenseError(err)
	}

	return toDomainExpense(updated)
}

func (r *ExpensesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	rowsAffected, err := r.queries.DeleteExpense(ctx, id)
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

// numericFromKopecks builds a NUMERIC value from minor units (kopecks).
func numericFromKopecks(kopecks int64) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(kopecks), Exp: -2, Valid: true}
}

// kopecksFromNumeric converts a NUMERIC value into minor units (kopecks);
// it fails if the value has more than two decimal places.
func kopecksFromNumeric(value pgtype.Numeric) (int64, error) {
	if !value.Valid {
		return 0, errors.New("amount is null")
	}

	if value.NaN {
		return 0, errors.New("amount is NaN")
	}

	unscaled := new(big.Int)
	if value.Int != nil {
		unscaled.Set(value.Int)
	}

	// value = unscaled * 10^Exp, so kopecks = unscaled * 10^(Exp+2).
	exp := int64(value.Exp) + 2
	if exp >= 0 {
		factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(exp), nil)
		unscaled.Mul(unscaled, factor)
	} else {
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(-exp), nil)
		quotient, remainder := new(big.Int).QuoRem(unscaled, divisor, new(big.Int))
		if remainder.Sign() != 0 {
			return 0, fmt.Errorf("amount has more than two decimal places")
		}

		unscaled = quotient
	}

	if !unscaled.IsInt64() {
		return 0, errors.New("amount is out of range")
	}

	return unscaled.Int64(), nil
}

func toDomainExpense(row gen.Expense) (domain.Expense, error) {
	kopecks, err := kopecksFromNumeric(row.Amount)
	if err != nil {
		return domain.Expense{}, fmt.Errorf("scan expense amount: %w", err)
	}

	return domain.Expense{
		ID:        row.ID,
		UserID:    row.UserID,
		Amount:    kopecks,
		Category:  domain.ExpenseCategory(row.Category),
		SpentAt:   row.SpentAt.Time,
		Comment:   row.Comment.String,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func mapExpenseError(err error) error {
	var pgErr *pgconn.PgError

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.ErrExpenseNotFound
	case errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation:
		// The only foreign key on expenses is user_id.
		return domain.ErrUserNotFound
	default:
		return err
	}
}
