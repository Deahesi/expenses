package expense_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/google/uuid"
)

func TestService_GetWithUser(t *testing.T) {
	owner := domain.User{ID: uuid.New(), TelegramID: 100, FirstName: "Иван", LastName: "Петров"}
	record := domain.Expense{
		ID: uuid.New(), UserID: owner.ID, Amount: 350, Category: domain.CategoryFood,
		SpentAt: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), Comment: "Кофе",
	}
	repo := newFakeRepository()
	repo.expenses[record.ID] = record
	repo.users[owner.ID] = owner
	service := expense.NewExpenseService(repo)

	for _, tt := range []struct {
		name      string
		id        uuid.UUID
		userID    uuid.UUID
		wantErr   error
		wantCalls int
	}{
		{name: "owner", id: record.ID, userID: owner.ID, wantCalls: 1},
		{name: "foreign owner", id: record.ID, userID: uuid.New(), wantErr: domain.ErrExpenseNotFound, wantCalls: 1},
		{name: "missing expense", id: uuid.New(), userID: owner.ID, wantErr: domain.ErrExpenseNotFound, wantCalls: 1},
		{name: "missing id", userID: owner.ID, wantErr: domain.ErrInvalidInput},
		{name: "missing owner", id: record.ID, wantErr: domain.ErrInvalidInput},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo.withUserCalls = 0
			got, err := service.GetWithUser(context.Background(), tt.id, tt.userID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetWithUser() error = %v, want %v", err, tt.wantErr)
			}
			if repo.withUserCalls != tt.wantCalls {
				t.Fatalf("repository calls = %d, want %d", repo.withUserCalls, tt.wantCalls)
			}
			if tt.wantErr == nil && (got.Expense != record || got.User != owner) {
				t.Fatalf("expense with user = %+v", got)
			}
		})
	}
}

func TestService_ListsWithUser(t *testing.T) {
	first := domain.User{ID: uuid.New(), TelegramID: 100, FirstName: "Иван"}
	second := domain.User{ID: uuid.New(), TelegramID: 200, FirstName: "Анна"}
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.users[first.ID], repo.users[second.ID] = first, second
	for _, record := range []domain.Expense{
		{ID: uuid.New(), UserID: first.ID, Amount: 350, Category: domain.CategoryFood, SpentAt: from},
		{ID: uuid.New(), UserID: second.ID, Amount: 450, Category: domain.CategoryFood, SpentAt: to},
	} {
		repo.expenses[record.ID] = record
	}
	service := expense.NewExpenseService(repo)

	for _, tt := range []struct {
		name       string
		filter     expense.ListFilter
		userScoped bool
		owners     []uuid.UUID
	}{
		{name: "all users", owners: []uuid.UUID{first.ID, second.ID}},
		{
			name: "combined filters",
			filter: expense.ListFilter{
				UserID: second.ID, Category: " еда ", DateFrom: from, DateTo: to,
			},
			owners: []uuid.UUID{second.ID},
		},
		{
			name:       "own expenses",
			filter:     expense.ListFilter{UserID: first.ID, DateFrom: from, DateTo: from},
			userScoped: true,
			owners:     []uuid.UUID{first.ID},
		},
		{name: "empty list", filter: expense.ListFilter{Category: "здоровье"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo.withUserCalls = 0
			var records []domain.ExpenseWithUser
			var err error
			if tt.userScoped {
				records, err = service.ListWithUser(context.Background(), tt.filter)
			} else {
				records, err = service.GetAllWithUser(context.Background(), tt.filter)
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(records) != len(tt.owners) {
				t.Fatalf("got %d expenses, want %d", len(records), len(tt.owners))
			}
			for _, record := range records {
				if !slices.Contains(tt.owners, record.UserID) || record.User != repo.users[record.UserID] {
					t.Errorf("unexpected expense owner: %+v", record)
				}
			}
			wantFilter := tt.filter
			wantFilter.Category = strings.TrimSpace(wantFilter.Category)
			if repo.withUserCalls != 1 || repo.lastFilter != wantFilter {
				t.Fatalf("repository calls = %d, filter = %+v", repo.withUserCalls, repo.lastFilter)
			}
		})
	}
}

func TestService_ListsWithUserErrors(t *testing.T) {
	storageErr := errors.New("storage unavailable")
	for _, tt := range []struct {
		name       string
		filter     expense.ListFilter
		userScoped bool
		repoErr    error
		wantErr    error
	}{
		{name: "missing owner", userScoped: true, wantErr: domain.ErrInvalidInput},
		{name: "unknown category", filter: expense.ListFilter{Category: "unknown"}, wantErr: domain.ErrInvalidInput},
		{
			name:       "invalid own category",
			filter:     expense.ListFilter{UserID: uuid.New(), Category: "unknown"},
			userScoped: true,
			wantErr:    domain.ErrInvalidInput,
		},
		{
			name: "inverted period",
			filter: expense.ListFilter{
				DateFrom: time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),
				DateTo:   time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: domain.ErrInvalidInput,
		},
		{name: "all users storage error", repoErr: storageErr, wantErr: storageErr},
		{
			name: "own storage error", filter: expense.ListFilter{UserID: uuid.New()},
			userScoped: true, repoErr: storageErr, wantErr: storageErr,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.listErr = tt.repoErr
			service := expense.NewExpenseService(repo)
			var err error
			if tt.userScoped {
				_, err = service.ListWithUser(context.Background(), tt.filter)
			} else {
				_, err = service.GetAllWithUser(context.Background(), tt.filter)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.repoErr == nil && repo.withUserCalls != 0 {
				t.Fatal("invalid filters reached the repository")
			}
		})
	}
}
