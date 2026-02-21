package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminOnlyMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := AdminOnlyMiddleware(nextHandler)

	t.Run("Customer is Blocked", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", nil)
		w := httptest.NewRecorder()

		claims := UserClaims{
			UserID: "some-uuid-1234",
			Role:   "customer",
		}
		ctx := context.WithValue(req.Context(), UserContextKey, claims)
		req = req.WithContext(ctx)

		handlerToTest.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden for customer, got %d", w.Code)
		}
	})

	t.Run("Admin is Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", nil)
		w := httptest.NewRecorder()

		claims := UserClaims{
			UserID: "some-uuid-5678",
			Role:   "admin",
		}
		ctx := context.WithValue(req.Context(), UserContextKey, claims)
		req = req.WithContext(ctx)

		handlerToTest.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 OK for admin, got %d", w.Code)
		}
	})
}
