package service_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"agentsight/internal/config"
	"agentsight/internal/models"
	"agentsight/internal/service"
)

type mockUserRepo struct {
	users map[int]*models.User
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}
func (m *mockUserRepo) GetByGitHubID(ctx context.Context, githubID string) (*models.User, error) {
	for _, u := range m.users {
		if u.GitHubID == githubID {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}
func (m *mockUserRepo) Create(ctx context.Context, user *models.User) (*models.User, error) {
	user.ID = len(m.users) + 1
	m.users[user.ID] = user
	return user, nil
}
func (m *mockUserRepo) UpsertByGitHub(ctx context.Context, user *models.User) (*models.User, error) {
	for _, u := range m.users {
		if u.GitHubID == user.GitHubID {
			u.Username = user.Username
			u.AvatarURL = user.AvatarURL
			u.Email = user.Email
			return u, nil
		}
	}
	user.ID = len(m.users) + 1
	m.users[user.ID] = user
	return user, nil
}

type mockSessionRepo struct {
	sessions map[string]*models.Session
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, session *models.Session) error {
	m.sessions[session.Token] = session
	return nil
}
func (m *mockSessionRepo) GetSession(ctx context.Context, token string) (*models.Session, error) {
	s, ok := m.sessions[token]
	if !ok {
		return nil, errors.New("session not found")
	}
	return s, nil
}
func (m *mockSessionRepo) DeleteSession(ctx context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}
func (m *mockSessionRepo) DeleteExpiredSessions(ctx context.Context) error {
	return nil
}

func TestAuthService_GetAuthURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL:        "http://localhost:8080",
		GitHubClientID: "test-client-id",
	}

	svc := service.NewAuthService(cfg, nil, nil)
	authURL := svc.GetAuthURL("random-state-123")

	u, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("failed to parse auth URL: %v", err)
	}

	if u.Host != "github.com" {
		t.Errorf("expected host github.com, got %s", u.Host)
	}
	if u.Query().Get("client_id") != "test-client-id" {
		t.Errorf("expected client_id test-client-id, got %s", u.Query().Get("client_id"))
	}
	if u.Query().Get("state") != "random-state-123" {
		t.Errorf("expected state random-state-123, got %s", u.Query().Get("state"))
	}
}

func TestAuthService_ValidateSessionAndLogout(t *testing.T) {
	user := &models.User{
		ID:        1,
		GitHubID:  "12345",
		Username:  "octocat",
		AvatarURL: "https://example.com/avatar.png",
	}
	userRepo := &mockUserRepo{
		users: map[int]*models.User{1: user},
	}
	sessionRepo := &mockSessionRepo{
		sessions: map[string]*models.Session{
			"valid-token": {
				Token:     "valid-token",
				UserID:    1,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			"expired-token": {
				Token:     "expired-token",
				UserID:    1,
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
		},
	}

	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	svc := service.NewAuthService(cfg, userRepo, sessionRepo)
	ctx := context.Background()

	// 1. Valid session
	foundUser, err := svc.ValidateSession(ctx, "valid-token")
	if err != nil {
		t.Fatalf("expected valid session, got err: %v", err)
	}
	if foundUser.ID != 1 || foundUser.Username != "octocat" {
		t.Errorf("unexpected user: %+v", foundUser)
	}

	// 2. Expired session
	_, err = svc.ValidateSession(ctx, "expired-token")
	if err == nil {
		t.Errorf("expected error for expired session, got nil")
	}

	// 3. Unknown token
	_, err = svc.ValidateSession(ctx, "unknown-token")
	if err == nil {
		t.Errorf("expected error for unknown session, got nil")
	}

	// 4. Logout
	if err := svc.Logout(ctx, "valid-token"); err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if _, ok := sessionRepo.sessions["valid-token"]; ok {
		t.Errorf("expected session to be deleted after logout")
	}
}

func TestTrendingService(t *testing.T) {
	mockRepo := &mockSkillRepo{
		searchResults: []models.Skill{
			{ID: 1, Name: "Trending 1", StarsVelocity: 100},
			{ID: 2, Name: "Trending 2", StarsVelocity: 50},
		},
	}

	svc := service.NewTrendingService(mockRepo)

	// Stars velocity calculation
	v := svc.CalculateStarsVelocity(100, 30, 7) // 70 stars over 7 days -> 70
	if v != 70 {
		t.Errorf("expected velocity 70, got %d", v)
	}

	v2 := svc.CalculateStarsVelocity(50, 10, 14) // 40 stars over 14 days -> (40 * 7)/14 = 20
	if v2 != 20 {
		t.Errorf("expected velocity 20, got %d", v2)
	}

	vNegative := svc.CalculateStarsVelocity(10, 20, 7)
	if vNegative != 0 {
		t.Errorf("expected velocity 0 for negative diff, got %d", vNegative)
	}
}
