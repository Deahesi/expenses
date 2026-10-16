package user_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/user"
	"github.com/google/uuid"
)

type fakeRepository struct {
	users map[uuid.UUID]domain.User
	err   error
}

func (r *fakeRepository) Create(_ context.Context, record domain.User) (domain.User, error) {
	if r.err != nil {
		return domain.User{}, r.err
	}
	for _, existing := range r.users {
		if existing.TelegramID == record.TelegramID {
			return domain.User{}, domain.ErrUserAlreadyExists
		}
	}

	record.CreatedAt = time.Now()
	record.UpdatedAt = record.CreatedAt
	r.users[record.ID] = record
	return record, nil
}

func (r *fakeRepository) Get(_ context.Context, id uuid.UUID) (domain.User, error) {
	if r.err != nil {
		return domain.User{}, r.err
	}
	record, ok := r.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return record, nil
}

func (r *fakeRepository) GetByTelegramID(_ context.Context, telegramID int64) (domain.User, error) {
	if r.err != nil {
		return domain.User{}, r.err
	}
	for _, record := range r.users {
		if record.TelegramID == telegramID {
			return record, nil
		}
	}
	return domain.User{}, domain.ErrUserNotFound
}

func (r *fakeRepository) Upsert(ctx context.Context, record domain.User) (domain.User, error) {
	existing, err := r.GetByTelegramID(ctx, record.TelegramID)
	if errors.Is(err, domain.ErrUserNotFound) {
		return r.Create(ctx, record)
	}
	if err != nil {
		return domain.User{}, err
	}
	return r.Update(ctx, domain.User{
		ID:        existing.ID,
		FirstName: record.FirstName,
		LastName:  record.LastName,
	})
}

func (r *fakeRepository) List(_ context.Context) ([]domain.User, error) {
	if r.err != nil {
		return nil, r.err
	}
	result := make([]domain.User, 0, len(r.users))
	for _, record := range r.users {
		result = append(result, record)
	}
	return result, nil
}

func (r *fakeRepository) Update(ctx context.Context, record domain.User) (domain.User, error) {
	existing, err := r.Get(ctx, record.ID)
	if err != nil {
		return domain.User{}, err
	}
	existing.FirstName = record.FirstName
	existing.LastName = record.LastName
	existing.UpdatedAt = time.Now()
	r.users[record.ID] = existing
	return existing, nil
}

func (r *fakeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.Get(ctx, id); err != nil {
		return err
	}
	delete(r.users, id)
	return nil
}

func TestServiceCRUD(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepository{users: make(map[uuid.UUID]domain.User)}
	service := user.NewUserService(repo)

	created, err := service.Create(ctx, user.CreateUserInput{
		TelegramID: 42,
		FirstName:  "  Иван  ",
		LastName:   "  Петров  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == uuid.Nil || created.FirstName != "Иван" || created.LastName != "Петров" {
		t.Fatalf("unexpected created user: %+v", created)
	}

	found, err := service.GetByTelegramID(ctx, 42)
	if err != nil || found.ID != created.ID {
		t.Fatalf("GetByTelegramID() = %+v, %v", found, err)
	}
	if _, err := service.Create(ctx, user.CreateUserInput{TelegramID: 42, FirstName: "Другой"}); !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("duplicate user error = %v", err)
	}

	updated, err := service.Update(ctx, created.ID, user.UpdateUserInput{FirstName: "  Пётр  "})
	if err != nil {
		t.Fatal(err)
	}
	if updated.FirstName != "Пётр" || updated.LastName != "" || updated.TelegramID != 42 {
		t.Fatalf("unexpected updated user: %+v", updated)
	}
	listed, err := service.List(ctx)
	if err != nil || len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("List() = %+v, %v", listed, err)
	}

	if err := service.Delete(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get(ctx, created.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("Get() after delete error = %v", err)
	}
	if err := service.Delete(ctx, created.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("Delete() for missing user error = %v", err)
	}
}

func TestCreateValidation(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input user.CreateUserInput
	}{
		{
			name:  "missing first name",
			input: user.CreateUserInput{TelegramID: 42, FirstName: "  "},
		},
		{
			name:  "negative telegram id",
			input: user.CreateUserInput{TelegramID: -1, FirstName: "Иван"},
		},
		{
			name:  "first name too long",
			input: user.CreateUserInput{TelegramID: 42, FirstName: strings.Repeat("я", 101)},
		},
		{
			name: "last name too long",
			input: user.CreateUserInput{
				TelegramID: 42,
				FirstName:  "Иван",
				LastName:   strings.Repeat("я", 101),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{users: make(map[uuid.UUID]domain.User)}
			_, err := user.NewUserService(repo).Create(context.Background(), tt.input)
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("Create() error = %v, want ErrInvalidInput", err)
			}
			if len(repo.users) != 0 {
				t.Error("invalid user was saved")
			}
		})
	}
}

func TestUpsertTelegramUserKeepsIdentity(t *testing.T) {
	ctx := context.Background()
	service := user.NewUserService(&fakeRepository{users: make(map[uuid.UUID]domain.User)})
	first, err := service.UpsertTelegramUser(ctx, user.UpsertTelegramUserInput{
		TelegramID: 42,
		FirstName:  "Иван",
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.UpsertTelegramUser(ctx, user.UpsertTelegramUserInput{
		TelegramID: 42,
		FirstName:  "Пётр",
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID || second.TelegramID != 42 || second.FirstName != "Пётр" {
		t.Fatalf("upsert changed identity or lost profile changes: %+v", second)
	}
	for _, id := range []int64{0, -1} {
		if _, err := service.UpsertTelegramUser(ctx, user.UpsertTelegramUserInput{TelegramID: id}); err == nil {
			t.Errorf("UpsertTelegramUser(%d) accepted an invalid ID", id)
		}
	}
}

func TestServicePreservesRepositoryErrors(t *testing.T) {
	ctx := context.Background()
	storageErr := errors.New("database unavailable")
	service := user.NewUserService(&fakeRepository{err: storageErr})

	_, err := service.Create(ctx, user.CreateUserInput{TelegramID: 42, FirstName: "Иван"})
	if !errors.Is(err, storageErr) {
		t.Errorf("Create() lost repository error: %v", err)
	}
	_, err = service.GetByTelegramID(ctx, 42)
	if !errors.Is(err, storageErr) {
		t.Errorf("GetByTelegramID() lost repository error: %v", err)
	}
	_, err = service.UpsertTelegramUser(ctx, user.UpsertTelegramUserInput{TelegramID: 42})
	if !errors.Is(err, storageErr) {
		t.Errorf("UpsertTelegramUser() lost repository error: %v", err)
	}
}
