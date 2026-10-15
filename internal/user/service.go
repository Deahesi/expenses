package user

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/google/uuid"
)

const maxNameLength = 100

// Service implements user management use cases.
type UserService struct {
	repo UsersRepository
}

func NewUserService(repo UsersRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (domain.User, error) {
	firstName, err := normalizeName("first_name", input.FirstName, true)
	if err != nil {
		return domain.User{}, err
	}

	lastName, err := normalizeName("last_name", input.LastName, false)
	if err != nil {
		return domain.User{}, err
	}

	if input.TelegramID < 0 {
		return domain.User{}, fmt.Errorf("%w: telegram_id must be positive", domain.ErrInvalidInput)
	}

	created, err := s.repo.Create(ctx, domain.User{
		ID:         uuid.New(),
		TelegramID: input.TelegramID,
		FirstName:  firstName,
		LastName:   lastName,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}

func (s *UserService) UpsertTelegramUser(
	ctx context.Context,
	input UpsertTelegramUserInput,
) (domain.User, error) {
	if input.TelegramID <= 0 {
		return domain.User{}, fmt.Errorf("invalid telegram id")
	}

	return s.repo.Upsert(ctx, domain.User{
		ID:         uuid.New(),
		TelegramID: input.TelegramID,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
	})
}

func (s *UserService) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

// GetByTelegramID resolves a Telegram account to a user; the bot uses it
// to identify the sender of a message.
func (s *UserService) GetByTelegramID(ctx context.Context, telegramID int64) (domain.User, error) {
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by telegram id: %w", err)
	}

	return user, nil
}

func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (domain.User, error) {
	firstName, err := normalizeName("first_name", input.FirstName, true)
	if err != nil {
		return domain.User{}, err
	}

	lastName, err := normalizeName("last_name", input.LastName, false)
	if err != nil {
		return domain.User{}, err
	}

	updated, err := s.repo.Update(ctx, domain.User{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}

	return updated, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

// normalizeName trims the value and validates its length; names are stored
// as empty strings when absent.
func normalizeName(field, value string, required bool) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		if required {
			return "", fmt.Errorf("%w: %s is required", domain.ErrInvalidInput, field)
		}

		return "", nil
	}

	if utf8.RuneCountInString(value) > maxNameLength {
		return "", fmt.Errorf(
			"%w: %s must be at most %d characters",
			domain.ErrInvalidInput, field, maxNameLength,
		)
	}

	return value, nil
}
