package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agentsight/internal/middleware"
	"agentsight/internal/models"
)

type mockAuthService struct {
	user *models.User
	err  error
}

func (m *mockAuthService) GetAuthURL(state string) string { return "" }
func (m *mockAuthService) HandleCallback(ctx context.Context, code string) (*models.User, string, error) {
	return m.user, "test-token", m.err
}
func (m *mockAuthService) ValidateSession(ctx context.Context, token string) (*models.User, error) {
	if token == "valid-token" {
		return m.user, nil
	}
	return nil, m.err
}
func (m *mockAuthService) Logout(ctx context.Context, token string) error { return nil }

func TestRateLimiter(t *testing.T) {
	limiter := middleware.NewRateLimiter(3, 100*time.Millisecond)
	defer limiter.Close()

	handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 3 requests should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/skills", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d expected 200, got %d", i+1, rec.Code)
		}
	}

	// 4th request should be rate limited (429)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/skills", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header to be set")
	}

	// Wait for window reset
	time.Sleep(150 * time.Millisecond)

	recReset := httptest.NewRecorder()
	handler.ServeHTTP(recReset, req)
	if recReset.Code != http.StatusOK {
		t.Errorf("after window reset, expected 200, got %d", recReset.Code)
	}
}

func TestAuthMiddleware(t *testing.T) {
	mockAuth := &mockAuthService{
		user: &models.User{
			ID:       7,
			Username: "testuser",
		},
	}

	var userInContext *models.User
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInContext = middleware.UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	stack := middleware.WithUser(mockAuth)(testHandler)

	// 1. Without cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	stack.ServeHTTP(rec, req)
	if userInContext != nil {
		t.Errorf("expected user in context to be nil")
	}

	// 2. With valid session cookie
	reqWithCookie := httptest.NewRequest(http.MethodGet, "/", nil)
	reqWithCookie.AddCookie(&http.Cookie{
		Name:  middleware.SessionCookieName,
		Value: "valid-token",
	})
	recCookie := httptest.NewRecorder()
	stack.ServeHTTP(recCookie, reqWithCookie)
	if userInContext == nil || userInContext.ID != 7 {
		t.Errorf("expected user in context to have ID 7, got %+v", userInContext)
	}

	// 3. RequireAuth blocks unauthenticated API call with 401
	protectedAPI := middleware.RequireAuth(testHandler)
	apiReq := httptest.NewRequest(http.MethodGet, "/api/v1/bookmarks", nil)
	apiRec := httptest.NewRecorder()
	protectedAPI.ServeHTTP(apiRec, apiReq)
	if apiRec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", apiRec.Code)
	}

	// 4. RequireAuth redirects unauthenticated web call to /auth/github
	webReq := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	webRec := httptest.NewRecorder()
	protectedAPI.ServeHTTP(webRec, webReq)
	if webRec.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect, got %d", webRec.Code)
	}
	if loc := webRec.Header().Get("Location"); loc != "/auth/github" {
		t.Errorf("expected redirect to /auth/github, got %s", loc)
	}
}
