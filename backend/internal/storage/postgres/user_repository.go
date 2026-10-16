package postgres

import (
	"context"
	"errors"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/storage/postgres/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// uniqueViolation is the PostgreSQL error code for unique constraint violations.
const uniqueViolation = "23505"

// UserRepository implements user.Repository on top of PostgreSQL.
type UserRepository struct {
	queries *gen.Queries
}

func NewUserRepository(queries *gen.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	row, err := r.queries.CreateUser(ctx, gen.CreateUserParams{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
	})
	if err != nil {
		return domain.User{}, mapUserError(err)
	}

	return toDomainUser(&row), nil
}

func (r *UserRepository) Upsert(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	row, err := r.queries.UpsertTelegramUser(
		ctx,
		gen.UpsertTelegramUserParams{
			ID:         user.ID,
			TelegramID: user.TelegramID,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
		},
	)
	if err != nil {
		return domain.User{}, mapUserError(err)
	}

	return toDomainUser(&row), nil
}

func (r *UserRepository) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.queries.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, mapUserError(err)
	}

	return toDomainUser(&row), nil
}

func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (domain.User, error) {
	row, err := r.queries.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return domain.User{}, mapUserError(err)
	}

	return toDomainUser(&row), nil
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, mapUserError(err)
	}

	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, toDomainUser(&row))
	}

	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) (domain.User, error) {
	row, err := r.queries.UpdateUser(ctx, gen.UpdateUserParams{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
	if err != nil {
		return domain.User{}, mapUserError(err)
	}

	return toDomainUser(&row), nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	rowsAffected, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return mapUserError(err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func toDomainUser(
	user *gen.User,
) domain.User {
	return domain.User{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

func mapUserError(err error) error {
	var pgErr *pgconn.PgError

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.ErrUserNotFound
	case errors.As(err, &pgErr) && pgErr.Code == uniqueViolation:
		return domain.ErrUserAlreadyExists
	default:
		return err
	}
}
