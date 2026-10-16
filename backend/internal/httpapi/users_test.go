package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/httpapi"
	"github.com/Deahesi/expenses-backend/internal/user"
	"github.com/google/uuid"
)

type usersRepository struct {
	records   []domain.User
	err       error
	listCalls int
}

func (r *usersRepository) List(context.Context) ([]domain.User, error) {
	r.listCalls++
	return r.records, r.err
}

func (r *usersRepository) Create(context.Context, domain.User) (domain.User, error) {
	return domain.User{}, errors.New("unexpected Create call")
}

func (r *usersRepository) Get(context.Context, uuid.UUID) (domain.User, error) {
	return domain.User{}, errors.New("unexpected Get call")
}

func (r *usersRepository) Upsert(context.Context, domain.User) (domain.User, error) {
	return domain.User{}, errors.New("unexpected Upsert call")
}

func (r *usersRepository) GetByTelegramID(context.Context, int64) (domain.User, error) {
	return domain.User{}, errors.New("unexpected GetByTelegramID call")
}

func (r *usersRepository) Update(context.Context, domain.User) (domain.User, error) {
	return domain.User{}, errors.New("unexpected Update call")
}

func (r *usersRepository) Delete(context.Context, uuid.UUID) error {
	return errors.New("unexpected Delete call")
}

func TestGetUsers(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	records := []domain.User{
		{
			ID: uuid.New(), TelegramID: 100, FirstName: "Иван", LastName: "Петров",
			CreatedAt: now, UpdatedAt: now.Add(time.Hour),
		},
		{
			ID: uuid.New(), TelegramID: 200, FirstName: "Анна",
			CreatedAt: now, UpdatedAt: now,
		},
	}
	for _, tt := range []struct {
		name    string
		path    string
		records []domain.User
		err     error
		status  int
	}{
		{name: "users without expenses", path: "/users", records: records, status: http.StatusOK},
		{name: "trailing slash", path: "/users/", records: records, status: http.StatusOK},
		{name: "nil list", path: "/users", status: http.StatusOK},
		{name: "empty list", path: "/users/", records: []domain.User{}, status: http.StatusOK},
		{
			name: "storage error", path: "/users", err: errors.New("private database error"),
			status: http.StatusInternalServerError,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &usersRepository{records: tt.records, err: tt.err}
			server := httpapi.NewServer(nil, user.NewUserService(repo))
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if response.Code != tt.status || repo.listCalls != 1 {
				t.Fatalf("status = %d, repository calls = %d, body = %s", response.Code, repo.listCalls, response.Body.String())
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("Content-Type = %q", response.Header().Get("Content-Type"))
			}
			if tt.err != nil {
				var body struct {
					Error string `json:"error"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Error != "internal server error" || strings.Contains(response.Body.String(), tt.err.Error()) {
					t.Fatalf("unexpected error response: %s", response.Body.String())
				}
				return
			}

			var got []domain.User
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.records) {
				t.Fatalf("got %d users, want %d", len(got), len(tt.records))
			}
			if len(got) == 0 {
				if strings.TrimSpace(response.Body.String()) != "[]" {
					t.Fatalf("empty response = %s, want []", response.Body.String())
				}
			} else if !reflect.DeepEqual(got, tt.records) {
				t.Fatalf("users = %+v, want %+v", got, tt.records)
			}
		})
	}
}
