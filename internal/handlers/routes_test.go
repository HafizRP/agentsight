package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agentsight/internal/handlers"
	"agentsight/internal/middleware"
	"agentsight/internal/models"
	"agentsight/internal/service"
)

type mockSkillRepo struct {
	skills []models.Skill
}

func (m *mockSkillRepo) Create(ctx context.Context, skill *models.Skill) (*models.Skill, error) {
	return skill, nil
}
func (m *mockSkillRepo) GetByID(ctx context.Context, id int) (*models.Skill, error) {
	if len(m.skills) > 0 {
		return &m.skills[0], nil
	}
	return nil, nil
}
func (m *mockSkillRepo) GetBySlug(ctx context.Context, slug string) (*models.Skill, error) {
	for _, s := range m.skills {
		if s.Slug == slug {
			return &s, nil
		}
	}
	return nil, nil
}
func (m *mockSkillRepo) Update(ctx context.Context, skill *models.Skill) error { return nil }
func (m *mockSkillRepo) Delete(ctx context.Context, id int) error              { return nil }
func (m *mockSkillRepo) List(ctx context.Context, filter models.SkillFilter) ([]models.Skill, int, error) {
	return m.skills, len(m.skills), nil
}
func (m *mockSkillRepo) Search(ctx context.Context, params models.SearchParams) ([]models.Skill, int, error) {
	return m.skills, len(m.skills), nil
}
func (m *mockSkillRepo) GetTrending(ctx context.Context, limit int) ([]models.Skill, error) {
	return m.skills, nil
}
func (m *mockSkillRepo) ListByCategory(ctx context.Context, category string, limit, offset int) ([]models.Skill, int, error) {
	return m.skills, len(m.skills), nil
}
func (m *mockSkillRepo) ListByPlatform(ctx context.Context, platform string, limit, offset int) ([]models.Skill, int, error) {
	return m.skills, len(m.skills), nil
}
func (m *mockSkillRepo) GetRelated(ctx context.Context, skill *models.Skill, limit int) ([]models.Skill, error) {
	return m.skills, nil
}
func (m *mockSkillRepo) GetStats(ctx context.Context) (*models.Stats, error) {
	return &models.Stats{
		TotalSkills:    len(m.skills),
		TotalStars:     500,
		TotalSources:   10,
		PlatformCounts: map[string]int{"cursor": 2},
		CategoryCounts: map[string]int{"rules": 2},
	}, nil
}

type mockBookmarkRepo struct{}

func (m *mockBookmarkRepo) Toggle(ctx context.Context, userID, skillID int) (bool, error) {
	return true, nil
}
func (m *mockBookmarkRepo) ListByUser(ctx context.Context, userID int, limit, offset int) ([]models.Skill, int, error) {
	return []models.Skill{}, 0, nil
}
func (m *mockBookmarkRepo) IsBookmarked(ctx context.Context, userID, skillID int) (bool, error) {
	return false, nil
}
func (m *mockBookmarkRepo) GetBookmarkedSkillIDs(ctx context.Context, userID int) (map[int]bool, error) {
	return nil, nil
}

type mockAuthService struct{}

func (m *mockAuthService) GetAuthURL(state string) string { return "https://github.com/login/oauth" }
func (m *mockAuthService) HandleCallback(ctx context.Context, code string) (*models.User, string, error) {
	return &models.User{ID: 1, Username: "test"}, "token", nil
}
func (m *mockAuthService) ValidateSession(ctx context.Context, token string) (*models.User, error) {
	if token == "valid" {
		return &models.User{ID: 1, Username: "test"}, nil
	}
	return nil, nil
}
func (m *mockAuthService) Logout(ctx context.Context, token string) error { return nil }

func setupTestRouter() http.Handler {
	skills := []models.Skill{
		{
			ID:          1,
			Name:        "Cursor Rule 1",
			Slug:        "cursor-rule-1",
			Description: "Best cursor rules",
			Platform:    models.PlatformCursor,
			Category:    models.CategoryRules,
			StarsCount:  250,
			CreatedAt:   time.Now(),
		},
	}

	skillRepo := &mockSkillRepo{skills: skills}
	bookmarkRepo := &mockBookmarkRepo{}

	searchSvc := service.NewSearchService(skillRepo)
	skillSvc := service.NewSkillService(skillRepo, bookmarkRepo)
	trendingSvc := service.NewTrendingService(skillRepo)
	authSvc := &mockAuthService{}

	h := &handlers.Handlers{
		Skill:    handlers.NewSkillHandler(skillSvc, trendingSvc),
		Search:   handlers.NewSearchHandler(searchSvc),
		Auth:     handlers.NewAuthHandler(authSvc),
		Bookmark: handlers.NewBookmarkHandler(bookmarkRepo),
		Health:   handlers.NewHealthHandler(nil),
	}

	limiter := middleware.NewRateLimiter(100, time.Minute)
	return handlers.SetupRouter(h, authSvc, limiter)
}

func TestHealthEndpoints(t *testing.T) {
	router := setupTestRouter()

	// GET /healthz -> 200
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil || res["status"] != "ok" {
		t.Errorf("expected status: ok, got %v", res)
	}
}

func TestAPIEndpoints(t *testing.T) {
	router := setupTestRouter()

	endpoints := []struct {
		name       string
		method     string
		url        string
		expectCode int
	}{
		{"landing page html", http.MethodGet, "/", http.StatusOK},
		{"trending page html", http.MethodGet, "/trending", http.StatusOK},
		{"category page html", http.MethodGet, "/category/rules", http.StatusOK},
		{"platform page html", http.MethodGet, "/platform/cursor", http.StatusOK},
		{"api skills", http.MethodGet, "/api/v1/skills", http.StatusOK},
		{"api search", http.MethodGet, "/api/v1/search?q=cursor", http.StatusOK},
		{"api trending", http.MethodGet, "/api/v1/trending", http.StatusOK},
		{"api stats", http.MethodGet, "/api/v1/stats", http.StatusOK},
		{"api bookmarks unauthorized", http.MethodGet, "/api/v1/bookmarks", http.StatusUnauthorized},
		{"api toggle bookmark unauthorized", http.MethodPost, "/api/v1/bookmarks/1", http.StatusUnauthorized},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.url, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != ep.expectCode {
				t.Errorf("%s %s expected status %d, got %d", ep.method, ep.url, ep.expectCode, rec.Code)
			}
		})
	}
}
