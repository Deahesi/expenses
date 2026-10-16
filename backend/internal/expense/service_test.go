package expense_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/google/uuid"
)

// fakeRepository is an in-memory expense.Repository implementation.
// Get, Update and Delete are scoped by (id, user_id) the same way the
// generated SQL queries are, so ownership violations surface as
// domain.ErrExpenseNotFound.
type fakeRepository struct {
	expenses      map[uuid.UUID]domain.Expense
	users         map[uuid.UUID]domain.User
	lastFilter    expense.ListFilter
	listErr       error
	getAllCalls   int
	withUserCalls int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		expenses: make(map[uuid.UUID]domain.Expense),
		users:    make(map[uuid.UUID]domain.User),
	}
}

func (r *fakeRepository) Create(_ context.Context, expense domain.Expense) (domain.Expense, error) {
	expense.CreatedAt = time.Now()
	expense.UpdatedAt = expense.CreatedAt
	r.expenses[expense.ID] = expense

	return expense, nil
}

func (r *fakeRepository) Get(_ context.Context, id uuid.UUID, userID uuid.UUID) (domain.Expense, error) {
	expense, ok := r.expenses[id]
	if !ok || expense.UserID != userID {
		return domain.Expense{}, domain.ErrExpenseNotFound
	}

	return expense, nil
}

func (r *fakeRepository) GetAll(ctx context.Context, filter expense.ListFilter) ([]domain.Expense, error) {
	r.getAllCalls++
	return r.List(ctx, filter)
}

func (r *fakeRepository) GetWithUser(ctx context.Context, id, userID uuid.UUID) (domain.ExpenseWithUser, error) {
	r.withUserCalls++
	record, err := r.Get(ctx, id, userID)
	if err != nil {
		return domain.ExpenseWithUser{}, err
	}

	return domain.ExpenseWithUser{Expense: record, User: r.users[userID]}, nil
}

func (r *fakeRepository) GetAllWithUser(ctx context.Context, filter expense.ListFilter) ([]domain.ExpenseWithUser, error) {
	r.withUserCalls++
	records, err := r.GetAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := make([]domain.ExpenseWithUser, 0, len(records))
	for _, record := range records {
		result = append(result, domain.ExpenseWithUser{Expense: record, User: r.users[record.UserID]})
	}
	return result, nil
}

func (r *fakeRepository) ListWithUser(ctx context.Context, filter expense.ListFilter) ([]domain.ExpenseWithUser, error) {
	r.withUserCalls++
	records, err := r.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := make([]domain.ExpenseWithUser, 0, len(records))
	for _, record := range records {
		result = append(result, domain.ExpenseWithUser{Expense: record, User: r.users[record.UserID]})
	}
	return result, nil
}

func (r *fakeRepository) List(_ context.Context, filter expense.ListFilter) ([]domain.Expense, error) {
	r.lastFilter = filter
	if r.listErr != nil {
		return nil, r.listErr
	}

	expenses := make([]domain.Expense, 0, len(r.expenses))
	for _, expense := range r.expenses {
		if filter.UserID != uuid.Nil && expense.UserID != filter.UserID {
			continue
		}
		if filter.Category != "" && string(expense.Category) != filter.Category {
			continue
		}
		if !filter.DateFrom.IsZero() && expense.SpentAt.Before(filter.DateFrom) {
			continue
		}
		if !filter.DateTo.IsZero() && expense.SpentAt.After(filter.DateTo) {
			continue
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func (r *fakeRepository) Update(_ context.Context, expense domain.Expense) (domain.Expense, error) {
	existing, ok := r.expenses[expense.ID]
	if !ok || existing.UserID != expense.UserID {
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

func (r *fakeRepository) Delete(_ context.Context, id uuid.UUID, userID uuid.UUID) error {
	expense, ok := r.expenses[id]
	if !ok || expense.UserID != userID {
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

	// Amounts are stored as whole rubles, no kopecks conversion.
	if created.Amount != 350 {
		t.Errorf("Create() amount = %d, want 350", created.Amount)
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

	stored, err := repo.Get(ctx, created.ID, created.UserID)
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

	owner := testUser(t).ID
	created, err := service.Create(ctx, expense.CreateExpenseInput{
		UserID:   owner,
		Amount:   1200,
		Category: "транспорт",
		SpentAt:  time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := service.Get(ctx, created.ID, owner)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !reflect.DeepEqual(got, created) {
		t.Errorf("Get() = %+v, want %+v", got, created)
	}

	if _, err := service.Get(ctx, uuid.New(), owner); !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("Get(unknown id) error = %v, want ErrExpenseNotFound", err)
	}

	// A user must not read another user's expense.
	if _, err := service.Get(ctx, created.ID, uuid.New()); !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("Get(foreign user) error = %v, want ErrExpenseNotFound", err)
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

	if len(expenses) != 1 {
		t.Errorf("List() = %d expenses, want 1", len(expenses))
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
			filter: expense.ListFilter{UserID: uuid.New(), Category: "непонятно"},
			errMsg: `unknown category "непонятно"`,
		},
		{
			name: "inverted period",
			filter: expense.ListFilter{
				UserID:   uuid.New(),
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

	owner := testUser(t).ID
	created, err := service.Create(ctx, expense.CreateExpenseInput{
		UserID:   owner,
		Amount:   350,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
		Comment:  "кофе",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := service.Update(ctx, created.ID, owner, expense.UpdateExpenseInput{
		Amount:   450,
		Category: " транспорт ",
		SpentAt:  time.Date(2026, 10, 16, 15, 45, 0, 0, time.UTC),
		Comment:  " такси ",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Amount != 450 {
		t.Errorf("Update() amount = %d, want 450", updated.Amount)
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

	if _, err := service.Update(ctx, uuid.New(), owner, expense.UpdateExpenseInput{
		Amount:   1,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC),
	}); !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("Update(unknown id) error = %v, want ErrExpenseNotFound", err)
	}

	if _, err := service.Update(ctx, created.ID, owner, expense.UpdateExpenseInput{
		Amount:   -1,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC),
	}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("Update(invalid amount) error = %v, want ErrInvalidInput", err)
	}
}

// A user must not update another user's expense: the update must be
// rejected and the stored record must stay untouched.
func TestService_Update_ForeignUser(t *testing.T) {
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)
	ctx := context.Background()

	owner := testUser(t).ID
	created, err := service.Create(ctx, expense.CreateExpenseInput{
		UserID:   owner,
		Amount:   350,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
		Comment:  "кофе",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = service.Update(ctx, created.ID, uuid.New(), expense.UpdateExpenseInput{
		Amount:   1,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC),
		Comment:  "взлом",
	})
	if !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Fatalf("Update(foreign user) error = %v, want ErrExpenseNotFound", err)
	}

	stored, err := service.Get(ctx, created.ID, owner)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if stored.Amount != 350 || stored.Comment != "кофе" {
		t.Errorf("foreign update modified the record: %+v", stored)
	}
}

func TestService_Delete(t *testing.T) {
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)
	ctx := context.Background()

	owner := testUser(t).ID
	created, err := service.Create(ctx, expense.CreateExpenseInput{
		UserID:   owner,
		Amount:   350,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := service.Delete(ctx, created.ID, owner); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := service.Get(ctx, created.ID, owner); !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("Get(deleted id) error = %v, want ErrExpenseNotFound", err)
	}

	if err := service.Delete(ctx, created.ID, owner); !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("Delete(deleted id) error = %v, want ErrExpenseNotFound", err)
	}
}

// A user must not delete another user's expense: the delete must be
// rejected and the record must remain available to its owner.
func TestService_Delete_ForeignUser(t *testing.T) {
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)
	ctx := context.Background()

	owner := testUser(t).ID
	created, err := service.Create(ctx, expense.CreateExpenseInput{
		UserID:   owner,
		Amount:   350,
		Category: "еда",
		SpentAt:  time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := service.Delete(ctx, created.ID, uuid.New()); !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Fatalf("Delete(foreign user) error = %v, want ErrExpenseNotFound", err)
	}

	if _, err := service.Get(ctx, created.ID, owner); err != nil {
		t.Errorf("Get(owner) after foreign delete attempt error = %v, want nil", err)
	}
}

func TestService_GetAll(t *testing.T) {
	ctx := context.Background()
	firstUser, secondUser := uuid.New(), uuid.New()
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)

	for _, input := range []expense.CreateExpenseInput{
		{UserID: firstUser, Amount: 100, Category: "еда", SpentAt: from},
		{UserID: secondUser, Amount: 200, Category: "еда", SpentAt: to},
		{UserID: firstUser, Amount: 300, Category: "транспорт", SpentAt: from.AddDate(0, 0, 14)},
		{UserID: secondUser, Amount: 400, Category: "еда", SpentAt: to.AddDate(0, 0, 1)},
	} {
		if _, err := service.Create(ctx, input); err != nil {
			t.Fatal(err)
		}
	}

	for _, tt := range []struct {
		name   string
		filter expense.ListFilter
		want   []int64
	}{
		{
			name: "all users",
			want: []int64{100, 200, 300, 400},
		},
		{
			name:   "user filter",
			filter: expense.ListFilter{UserID: firstUser},
			want:   []int64{100, 300},
		},
		{
			name:   "category across users",
			filter: expense.ListFilter{Category: " еда "},
			want:   []int64{100, 200, 400},
		},
		{
			name:   "inclusive period",
			filter: expense.ListFilter{DateFrom: from, DateTo: to},
			want:   []int64{100, 200, 300},
		},
		{
			name:   "start date only",
			filter: expense.ListFilter{DateFrom: to},
			want:   []int64{200, 400},
		},
		{
			name:   "end date only",
			filter: expense.ListFilter{DateTo: from},
			want:   []int64{100},
		},
		{
			name: "combined filters",
			filter: expense.ListFilter{
				UserID: secondUser, Category: "еда", DateFrom: from, DateTo: to,
			},
			want: []int64{200},
		},
		{
			name:   "no matches",
			filter: expense.ListFilter{Category: "здоровье"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo.getAllCalls = 0
			records, err := service.GetAll(ctx, tt.filter)
			if err != nil {
				t.Fatal(err)
			}

			amounts := make([]int64, 0, len(records))
			for _, record := range records {
				amounts = append(amounts, record.Amount)
			}
			slices.Sort(amounts)
			if !slices.Equal(amounts, tt.want) {
				t.Fatalf("GetAll() amounts = %v, want %v", amounts, tt.want)
			}

			filter := tt.filter
			filter.Category = strings.TrimSpace(filter.Category)
			if repo.getAllCalls != 1 || !reflect.DeepEqual(repo.lastFilter, filter) {
				t.Fatalf("repository calls = %d, filter = %+v", repo.getAllCalls, repo.lastFilter)
			}
		})
	}
}

func TestService_GetAllErrors(t *testing.T) {
	storageErr := errors.New("storage unavailable")
	for _, tt := range []struct {
		name    string
		filter  expense.ListFilter
		repoErr error
		wantErr error
	}{
		{
			name:    "unknown category",
			filter:  expense.ListFilter{Category: "unknown"},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "inverted period",
			filter: expense.ListFilter{
				DateFrom: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
				DateTo:   time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:    "repository error",
			repoErr: storageErr,
			wantErr: storageErr,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.listErr = tt.repoErr
			_, err := expense.NewExpenseService(repo).GetAll(context.Background(), tt.filter)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetAll() error = %v, want %v", err, tt.wantErr)
			}
			if tt.repoErr == nil && repo.getAllCalls != 0 {
				t.Fatal("invalid filters reached the repository")
			}
		})
	}
}

func TestService_ListRequiresUser(t *testing.T) {
	service := expense.NewExpenseService(newFakeRepository())
	_, err := service.List(context.Background(), expense.ListFilter{})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("List() error = %v, want ErrInvalidInput", err)
	}
}

func TestService_Total(t *testing.T) {
	owner := uuid.New()
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	service := expense.NewExpenseService(repo)
	for _, input := range []expense.CreateExpenseInput{
		{UserID: owner, Amount: 350, Category: "еда", SpentAt: from},
		{UserID: owner, Amount: 450, Category: "еда", SpentAt: to},
		{UserID: owner, Amount: 1200, Category: "транспорт", SpentAt: from},
		{UserID: owner, Amount: 100, Category: "еда", SpentAt: from.AddDate(0, 0, -1)},
		{UserID: owner, Amount: 200, Category: "еда", SpentAt: to.AddDate(0, 0, 1)},
		{UserID: uuid.New(), Amount: 9999, Category: "еда", SpentAt: from},
	} {
		if _, err := service.Create(context.Background(), input); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	for _, tt := range []struct {
		name   string
		filter expense.ListFilter
		want   int64
	}{
		{name: "category and inclusive period", filter: expense.ListFilter{UserID: owner, Category: " еда ", DateFrom: from, DateTo: to}, want: 800},
		{name: "all own expenses", filter: expense.ListFilter{UserID: owner}, want: 2300},
		{name: "no matches", filter: expense.ListFilter{UserID: owner, Category: "здоровье"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.Total(context.Background(), tt.filter)
			if err != nil {
				t.Fatalf("Total() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Total() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestService_TotalErrors(t *testing.T) {
	owner := uuid.New()
	storageErr := errors.New("storage unavailable")
	for _, tt := range []struct {
		name    string
		filter  expense.ListFilter
		repoErr error
		wantErr error
	}{
		{name: "missing user", wantErr: domain.ErrInvalidInput},
		{name: "invalid category", filter: expense.ListFilter{UserID: owner, Category: "unknown"}, wantErr: domain.ErrInvalidInput},
		{name: "inverted period", filter: expense.ListFilter{UserID: owner, DateFrom: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), DateTo: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)}, wantErr: domain.ErrInvalidInput},
		{name: "repository error", filter: expense.ListFilter{UserID: owner}, repoErr: storageErr, wantErr: storageErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.listErr = tt.repoErr
			_, err := expense.NewExpenseService(repo).Total(context.Background(), tt.filter)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Total() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_UpdateRejectsZeroAmount(t *testing.T) {
	ctx := context.Background()
	service := expense.NewExpenseService(newFakeRepository())
	created, err := service.Create(ctx, expense.CreateExpenseInput{
		UserID: uuid.New(), Amount: 350, Category: "еда", SpentAt: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = service.Update(ctx, created.ID, created.UserID, expense.UpdateExpenseInput{
		Amount: 0, Category: "еда", SpentAt: created.SpentAt,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("Update() error = %v, want ErrInvalidInput", err)
	}
	stored, err := service.Get(ctx, created.ID, created.UserID)
	if err != nil || stored.Amount != 350 {
		t.Errorf("expense changed after invalid update: %+v, error = %v", stored, err)
	}
}
