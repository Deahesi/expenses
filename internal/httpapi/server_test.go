package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Deahesi/expenses-backend/internal/httpapi"
)

func TestInvalidRequestsAreRejectedBeforeCallingServices(t *testing.T) {
	server := httpapi.NewServer(nil, nil)
	for _, tt := range []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "invalid expense JSON",
			method: http.MethodPost,
			path:   "/expenses/9ef26bfc-1f07-4373-ab7d-8a1c5c119888",
			body:   "{",
		},
		{
			name:   "invalid expense owner",
			method: http.MethodPost,
			path:   "/expenses/invalid",
			body:   `{"user_id":"invalid","amount":350,"category":"еда","spent_at":"2026-10-16"}`,
		},
		{
			name:   "invalid expense date",
			method: http.MethodPost,
			path:   "/expenses/9ef26bfc-1f07-4373-ab7d-8a1c5c119888",
			body:   `{"user_id":"9ef26bfc-1f07-4373-ab7d-8a1c5c119888","amount":350,"category":"еда","spent_at":"вчера"}`,
		},
		{
			name:   "invalid member filter",
			method: http.MethodGet,
			path:   "/expenses/?user_id=invalid",
		},
		{
			name:   "invalid category filter",
			method: http.MethodGet,
			path:   "/expenses/?category=unknown",
		},
		{
			name:   "inverted period",
			method: http.MethodGet,
			path:   "/expenses/?date_from=2026-10-31&date_to=2026-10-01",
		},
		{
			name:   "invalid scoped owner",
			method: http.MethodGet,
			path:   "/expenses/invalid",
		},
		{
			name:   "invalid expense id",
			method: http.MethodGet,
			path:   "/expenses/9ef26bfc-1f07-4373-ab7d-8a1c5c119888/invalid",
		},
		{
			name:   "invalid date filter",
			method: http.MethodGet,
			path:   "/expenses/?date_from=invalid",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Error == "" {
				t.Fatalf("expected a JSON error, got %s", response.Body.String())
			}
		})
	}
}
