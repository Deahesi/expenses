package expense_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/google/uuid"
)

// fakeRepository is an in-memory expense.Repository implementation.
type fakeRepository struct {
	expenses   map[uuid.UUID]domain.Expense
	lastFilter expense.ListFilter
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{expenses: make(map[uuid.UUID]domain.Expense)}
}

func (r *fakeRepository) Create(_ context.Context, expense domain.Expense) (domain.Expense, error) {
	expense.CreatedAt = time.Now()
	expense.UpdatedAt = expense.CreatedAt
	r.expenses[expense.ID] = expense

	return expense, nil
}

func (r *fakeRepository) Get(_ context.Context, id uuid.UUID) (domain.Expense, error) {
	expense, ok := r.expenses[id]
	if !ok {
		return domain.Expense{}, domain.ErrExpenseNotFound
	}

	return expense, nil
}

func (r *fakeRepository) List(_ context.Context, filter expense.ListFilter) ([]domain.Expense, error) {
	r.lastFilter = filter

	expenses := make([]domain.Expense, 0, len(r.expenses))
	for _, expense := range r.expenses {
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func (r *fakeRepository) Update(_ context.Context, expense domain.Expense) (domain.Expense, error) {
	existing, ok := r.expenses[expense.ID]
	if !ok {
		return domain.Expense{}, domain.ErrExpenseNotFound
	}

	existing.Amount = expense.Amount
	existing.Category = expense.Category
	existing.SpentAt = expense.SpentAt
	existing.Comment = expense.Comment
	existing.UpdatedAt = time.Now()
	r.expenses[expense.ID] = existing

	return existing, nil
}

func (r *fakeRepository) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.expenses[id]; !ok {
		return domain.ErrExpenseNotFound
	}

	delete(r.expenses, id)

	return nil
}

func testUser(t *testing.T) domain.User {
	t.Helper()

	return domain.User{
		ID:        uuid.New(),
		FirstName: "Ivan",
		LastName:  "Petrov",
	}
}

func TestService_Create(t *testing.T) {
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)
	ctx := context.Background()

	created, err := service.Create(ctx, expense.CreateExpenseInput{
		UserID:   testUser(t).ID,
		Amount:   350,
		Category: " еда ",
		SpentAt:  time.Date(2026, 10, 15, 23, 30, 0, 0, time.FixedZone("MSK", 3*60*60)),
		Comment:  "  кофе с круассаном  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID == uuid.Nil {
		t.Fatal("Create() did not set id")
	}

	if created.Amount != 35050 {
		t.Errorf("Create() amount = %d kopecks, want 35050", created.Amount)
	}

	if created.Category != domain.CategoryFood {
		t.Errorf("Create() category = %q, want %q", created.Category, domain.CategoryFood)
	}

	wantDate := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	if !created.SpentAt.Equal(wantDate) {
		t.Errorf("Create() spent_at = %v, want %v", created.SpentAt, wantDate)
	}

	if created.Comment != "кофе с круассаном" {
		t.Errorf("Create() comment = %q", created.Comment)
	}

	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatal("Create() did not set timestamps")
	}

	stored, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("repo.Get() error = %v", err)
	}

	if !reflect.DeepEqual(stored, created) {
		t.Errorf("stored expense = %+v, want %+v", stored, created)
	}
}

func TestService_Create_Validation(t *testing.T) {
	userID := testUser(t).ID
	validDate := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		input  expense.CreateExpenseInput
		errMsg string
	}{
		{
			name:   "missing user",
			input:  expense.CreateExpenseInput{Amount: 100, Category: "еда", SpentAt: validDate},
			errMsg: "user_id is required",
		},
		{
			name:   "zero amount",
			input:  expense.CreateExpenseInput{UserID: userID, Amount: 0, Category: "еда", SpentAt: validDate},
			errMsg: "amount must be",
		},
		{
			name:   "negative amount",
			input:  expense.CreateExpenseInput{UserID: userID, Amount: -5, Category: "еда", SpentAt: validDate},
			errMsg: "amount must be",
		},
		{
			name:   "too large amount",
			input:  expense.CreateExpenseInput{UserID: userID, Amount: 1e10, Category: "еда", SpentAt: validDate},
			errMsg: "amount must be",
		},
		{
			name:   "unknown category",
			input:  expense.CreateExpenseInput{UserID: userID, Amount: 100, Category: "путешествия", SpentAt: validDate},
			errMsg: `unknown category "путешествия"`,
		},
		{
			name:   "missing spent_at",
			input:  expense.CreateExpenseInput{UserID: userID, Amount: 100, Category: "еда"},
			errMsg: "spent_at is required",
		},
		{
			name: "comment too long",
			input: expense.CreateExpenseInput{
				UserID:   userID,
				Amount:   100,
				Category: "еда",
				SpentAt:  validDate,
				Comment:  strings.Repeat("а", 501),
			},
			errMsg: "comment must be at most 500 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			service := expense.NewExpenseService(repo)

			_, err := service.Create(context.Background(), tt.input)
			if err == nil {
				t.Fatal("Create() expected error, got nil")
			}

			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("Create() error = %v, want ErrInvalidInput", err)
			}

			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Create() error = %q, want it to contain %q", err.Error(), tt.errMsg)
			}

			if len(repo.expenses) != 0 {
				t.Errorf("Create() persisted %d expenses, want 0", len(repo.expenses))
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)
	ctx := context.Background()

	created, err := service.Create(ctx, expense.CreateExpenseInput{
		UserID:   testUser(t).ID,
		Amount:   1200,
		Category: "транспорт",
		SpentAt:  time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := service.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !reflect.DeepEqual(got, created) {
		t.Errorf("Get() = %+v, want %+v", got, created)
	}

	if _, err := service.Get(ctx, uuid.New()); !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("Get(unknown id) error = %v, want ErrExpenseNotFound", err)
	}
}

func TestService_List(t *testing.T) {
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)
	ctx := context.Background()

	userID := testUser(t).ID
	for _, input := range []expense.CreateExpenseInput{
		{UserID: userID, Amount: 350, Category: "еда", SpentAt: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)},
		{UserID: userID, Amount: 1200, Category: "транспорт", SpentAt: time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)},
	} {
		if _, err := service.Create(ctx, input); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	dateFrom := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	filter := expense.ListFilter{UserID: userID, Category: "еда", DateFrom: dateFrom, DateTo: dateTo}

	expenses, err := service.List(ctx, filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(expenses) != 2 {
		t.Errorf("List() = %d expenses, want 2", len(expenses))
	}

	if !reflect.DeepEqual(repo.lastFilter, filter) {
		t.Errorf("repo received filter %+v, want %+v", repo.lastFilter, filter)
	}
}

func TestService_List_Validation(t *testing.T) {
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)
	ctx := context.Background()

	tests := []struct {
		name   string
		filter expense.ListFilter
		errMsg string
	}{
		{
			name:   "unknown category",
			filter: expense.ListFilter{Category: "непонятно"},
			errMsg: `unknown category "непонятно"`,
		},
		{
			name: "inverted period",
			filter: expense.ListFilter{
				DateFrom: time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),
				DateTo:   time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			errMsg: "date_from must not be after date_to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.List(ctx, tt.filter)
			if err == nil {
				t.Fatal("List() expected error, got nil")
			}

			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("List() error = %v, want ErrInvalidInput", err)
			}

			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("List() error = %q, want it to contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)
	ctx := context.Background()

	created, err := service.Create(ctx, expense.CreateExpenseInput{
		UserID:   testUser(t).ID,
		Amount:   350,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
		Comment:  "кофе",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := service.Update(ctx, created.ID, expense.UpdateExpenseInput{
		Amount:   450,
		Category: " транспорт ",
		SpentAt:  time.Date(2026, 10, 16, 15, 45, 0, 0, time.UTC),
		Comment:  " такси ",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Amount != 450 {
		t.Errorf("Update() amount = %d kopecks, want 450", updated.Amount)
	}

	if updated.Category != domain.CategoryTransport {
		t.Errorf("Update() category = %q, want %q", updated.Category, domain.CategoryTransport)
	}

	if !updated.SpentAt.Equal(time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Update() spent_at = %v", updated.SpentAt)
	}

	if updated.Comment != "такси" {
		t.Errorf("Update() comment = %q", updated.Comment)
	}

	if updated.UserID != created.UserID {
		t.Errorf("Update() changed user_id: %v, want %v", updated.UserID, created.UserID)
	}

	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Error("Update() did not bump updated_at")
	}

	if _, err := service.Update(ctx, uuid.New(), expense.UpdateExpenseInput{
		Amount:   1,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC),
	}); !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("Update(unknown id) error = %v, want ErrExpenseNotFound", err)
	}

	if _, err := service.Update(ctx, created.ID, expense.UpdateExpenseInput{
		Amount:   -1,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC),
	}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("Update(invalid amount) error = %v, want ErrInvalidInput", err)
	}
}

func TestService_Delete(t *testing.T) {
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)
	ctx := context.Background()

	created, err := service.Create(ctx, expense.CreateExpenseInput{
		UserID:   testUser(t).ID,
		Amount:   350,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := service.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := service.Get(ctx, created.ID); !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("Get(deleted id) error = %v, want ErrExpenseNotFound", err)
	}

	if err := service.Delete(ctx, created.ID); !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("Delete(deleted id) error = %v, want ErrExpenseNotFound", err)
	}
}
