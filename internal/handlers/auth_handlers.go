package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"agentsight/internal/middleware"
	"agentsight/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := hex.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	authURL := h.authService.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := q.Get("code")
	state := q.Get("state")

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" || stateCookie.Value != state {
		respondError(w, http.StatusBadRequest, "invalid or expired oauth state")
		return
	}

	// clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})

	if code == "" {
		respondError(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	_, sessionToken, err := h.authService.HandleCallback(r.Context(), code)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "authentication failed: "+err.Error())
		return
	}

	isSecure := strings.HasPrefix(strings.ToLower(r.Proto), "https") || r.Header.Get("X-Forwarded-Proto") == "https"
	middleware.SetSessionCookie(w, sessionToken, time.Now().Add(14*24*time.Hour), isSecure)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	token := middleware.GetSessionToken(r)
	if token != "" {
		_ = h.authService.Logout(r.Context(), token)
	}

	middleware.ClearSessionCookie(w)

	if wantsJSON(r) {
		respondJSON(w, http.StatusOK, map[string]string{
			"status": "logged out",
		})
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
