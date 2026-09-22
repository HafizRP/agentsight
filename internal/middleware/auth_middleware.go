package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"agentsight/internal/models"
	"agentsight/internal/service"
)

type contextKey string

const (
	userContextKey  contextKey = "agentsight_user"
	tokenContextKey contextKey = "agentsight_session_token"
)

func UserFromContext(ctx context.Context) *models.User {
	if ctx == nil {
		return nil
	}
	val := ctx.Value(userContextKey)
	if val == nil {
		return nil
	}
	if user, ok := val.(*models.User); ok {
		return user
	}
	return nil
}

func TokenFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val := ctx.Value(tokenContextKey)
	if val == nil {
		return ""
	}
	if token, ok := val.(string); ok {
		return token
	}
	return ""
}

func WithUser(authService service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := GetSessionToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, err := authService.ValidateSession(r.Context(), token)
			if err != nil || user == nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			ctx = context.WithValue(ctx, tokenContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil {
			if strings.HasPrefix(r.URL.Path, "/api/") || strings.Contains(r.Header.Get("Accept"), "application/json") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "authentication required",
				})
				return
			}
			http.Redirect(w, r, "/auth/github", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
