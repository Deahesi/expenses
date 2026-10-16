package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/httpapi"
	"github.com/Deahesi/expenses-backend/internal/httpapi/expenseHandlers"
	"github.com/google/uuid"
)

type expenseListRepository struct {
	filter      expense.ListFilter
	records     []domain.ExpenseWithUser
	err         error
	getAllCalls int
	listCalls   int
	getCalls    int
}

func (r *expenseListRepository) GetAllWithUser(_ context.Context, filter expense.ListFilter) ([]domain.ExpenseWithUser, error) {
	r.getAllCalls++
	r.filter = filter
	return r.records, r.err
}

func (r *expenseListRepository) ListWithUser(_ context.Context, filter expense.ListFilter) ([]domain.ExpenseWithUser, error) {
	r.listCalls++
	r.filter = filter
	return r.records, r.err
}

func (r *expenseListRepository) GetWithUser(_ context.Context, id, userID uuid.UUID) (domain.ExpenseWithUser, error) {
	r.getCalls++
	if r.err != nil {
		return domain.ExpenseWithUser{}, r.err
	}
	for _, record := range r.records {
		if record.ID == id && record.UserID == userID {
			return record, nil
		}
	}
	return domain.ExpenseWithUser{}, domain.ErrExpenseNotFound
}

func (r *expenseListRepository) GetAll(context.Context, expense.ListFilter) ([]domain.Expense, error) {
	return nil, errors.New("unexpected GetAll call")
}

func (r *expenseListRepository) List(context.Context, expense.ListFilter) ([]domain.Expense, error) {
	return nil, errors.New("unexpected List call")
}

func (r *expenseListRepository) Create(context.Context, domain.Expense) (domain.Expense, error) {
	return domain.Expense{}, errors.New("unexpected Create call")
}

func (r *expenseListRepository) Get(context.Context, uuid.UUID, uuid.UUID) (domain.Expense, error) {
	return domain.Expense{}, errors.New("unexpected Get call")
}

func (r *expenseListRepository) Update(context.Context, domain.Expense) (domain.Expense, error) {
	return domain.Expense{}, errors.New("unexpected Update call")
}

func (r *expenseListRepository) Delete(context.Context, uuid.UUID, uuid.UUID) error {
	return errors.New("unexpected Delete call")
}

func TestExpenseLists(t *testing.T) {
	firstUser, secondUser := uuid.New(), uuid.New()
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	records := []domain.ExpenseWithUser{
		{
			Expense: domain.Expense{ID: uuid.New(), UserID: firstUser, Amount: 350, Category: domain.CategoryFood, SpentAt: from},
			User:    domain.User{ID: firstUser, TelegramID: 100, FirstName: "Иван", LastName: "Петров"},
		},
		{
			Expense: domain.Expense{ID: uuid.New(), UserID: secondUser, Amount: 450, Category: domain.CategoryFood, SpentAt: to},
			User:    domain.User{ID: secondUser, TelegramID: 200, FirstName: "Анна", LastName: "Иванова"},
		},
	}
	query := url.Values{
		"user_id": {secondUser.String()}, "category": {" еда "},
		"date_from": {"2026-10-01"}, "date_to": {"2026-10-31"},
	}

	for _, tt := range []struct {
		name       string
		path       string
		filter     expense.ListFilter
		records    []domain.ExpenseWithUser
		repoErr    error
		userScoped bool
	}{
		{
			name:    "all users without filters",
			path:    "/expenses/",
			records: records,
		},
		{
			name: "combined filters",
			path: "/expenses/?" + query.Encode(),
			filter: expense.ListFilter{
				UserID: secondUser, Category: "еда", DateFrom: from, DateTo: to,
			},
			records: records[1:],
		},
		{
			name:    "end date only",
			path:    "/expenses/?date_to=2026-10-31",
			filter:  expense.ListFilter{DateTo: to},
			records: records,
		},
		{
			name:       "owner from path overrides query",
			path:       "/expenses/" + firstUser.String() + "?user_id=" + secondUser.String(),
			filter:     expense.ListFilter{UserID: firstUser},
			records:    records[:1],
			userScoped: true,
		},
		{
			name: "empty list",
			path: "/expenses/",
		},
		{
			name:    "repository error",
			path:    "/expenses/",
			repoErr: errors.New("database unavailable"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &expenseListRepository{records: tt.records, err: tt.repoErr}
			server := httpapi.NewServer(expense.NewExpenseService(repo), nil)
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))

			wantAllCalls, wantListCalls := 1, 0
			if tt.userScoped {
				wantAllCalls, wantListCalls = 0, 1
			}
			if repo.getAllCalls != wantAllCalls || repo.listCalls != wantListCalls {
				t.Fatalf("GetAll calls = %d, List calls = %d", repo.getAllCalls, repo.listCalls)
			}
			if !reflect.DeepEqual(repo.filter, tt.filter) {
				t.Fatalf("filter = %+v, want %+v", repo.filter, tt.filter)
			}
			if tt.repoErr != nil {
				if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), tt.repoErr.Error()) {
					t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
				}
				return
			}

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			var got []expenseHandlers.ExpenseWithUserResponse
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.records) {
				t.Fatalf("got %d expenses, want %d", len(got), len(tt.records))
			}
			for i, record := range tt.records {
				if got[i].ID != record.ID || got[i].UserID != record.UserID || got[i].Amount != float64(record.Amount) {
					t.Errorf("expense = %+v, want %+v", got[i], record)
				}
				if got[i].SpentAt != record.SpentAt.Format(time.DateOnly) {
					t.Errorf("spent_at = %q", got[i].SpentAt)
				}
				if got[i].User.ID != record.User.ID || got[i].User.TelegramID != record.User.TelegramID ||
					got[i].User.FirstName != record.User.FirstName || got[i].User.LastName != record.User.LastName {
					t.Errorf("user = %+v, want %+v", got[i].User, record.User)
				}
			}
			if len(got) == 0 && strings.TrimSpace(response.Body.String()) != "[]" {
				t.Fatalf("empty response = %s, want []", response.Body.String())
			}
		})
	}
}

func TestGetExpenseWithUser(t *testing.T) {
	owner := domain.User{ID: uuid.New(), TelegramID: 100, FirstName: "Иван", LastName: "Петров"}
	record := domain.ExpenseWithUser{
		Expense: domain.Expense{
			ID: uuid.New(), UserID: owner.ID, Amount: 350, Category: domain.CategoryFood,
			SpentAt: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), Comment: "Кофе",
		},
		User: owner,
	}
	for _, tt := range []struct {
		name    string
		id      uuid.UUID
		userID  uuid.UUID
		repoErr error
		status  int
	}{
		{name: "owner", id: record.ID, userID: owner.ID, status: http.StatusOK},
		{name: "foreign owner", id: record.ID, userID: uuid.New(), status: http.StatusNotFound},
		{name: "missing expense", id: uuid.New(), userID: owner.ID, status: http.StatusNotFound},
		{
			name: "storage error", id: record.ID, userID: owner.ID,
			repoErr: errors.New("database unavailable"), status: http.StatusInternalServerError,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &expenseListRepository{records: []domain.ExpenseWithUser{record}, err: tt.repoErr}
			server := httpapi.NewServer(expense.NewExpenseService(repo), nil)
			response := httptest.NewRecorder()
			path := "/expenses/" + tt.userID.String() + "/" + tt.id.String()
			server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))

			if response.Code != tt.status || repo.getCalls != 1 {
				t.Fatalf("status = %d, repository calls = %d, body = %s", response.Code, repo.getCalls, response.Body.String())
			}
			if tt.status != http.StatusOK {
				if tt.repoErr != nil && strings.Contains(response.Body.String(), tt.repoErr.Error()) {
					t.Fatal("storage error was exposed to the client")
				}
				return
			}
			var got expenseHandlers.ExpenseWithUserResponse
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.ID != record.ID || got.Amount != 350 || got.SpentAt != "2026-10-16" ||
				got.UserID != owner.ID || got.User.ID != owner.ID || got.User.TelegramID != owner.TelegramID ||
				got.User.FirstName != owner.FirstName || got.User.LastName != owner.LastName {
				t.Fatalf("unexpected response: %+v", got)
			}
		})
	}
}
