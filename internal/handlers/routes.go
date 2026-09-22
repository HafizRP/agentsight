package handlers

import (
	"net/http"

	"agentsight/internal/middleware"
	"agentsight/internal/repository"
	"agentsight/internal/service"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	Skill    *SkillHandler
	Search   *SearchHandler
	Auth     *AuthHandler
	Bookmark *BookmarkHandler
	Health   *HealthHandler
}

func NewHandlers(services *service.Services, repos *repository.Repositories, pool *pgxpool.Pool) *Handlers {
	return &Handlers{
		Skill:    NewSkillHandler(services.Skill, services.Trending),
		Search:   NewSearchHandler(services.Search),
		Auth:     NewAuthHandler(services.Auth),
		Bookmark: NewBookmarkHandler(repos.Bookmarks),
		Health:   NewHealthHandler(pool),
	}
}

func RegisterRoutes(r chi.Router, h *Handlers, authService service.AuthService, rateLimiter *middleware.RateLimiter) {
	// Base middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.WithUser(authService))

	// Health probes
	r.Get("/healthz", h.Health.Healthz)
	r.Get("/readyz", h.Health.Readyz)

	// Web routes
	r.Get("/", h.Skill.HandleLanding)
	r.Get("/search", h.Search.HandleSearchPage)
	r.Get("/skill/{slug}", h.Skill.HandleSkillDetail)
	r.Get("/trending", h.Skill.HandleTrendingPage)
	r.Get("/category/{category}", h.Skill.HandleCategoryPage)
	r.Get("/platform/{platform}", h.Skill.HandlePlatformPage)

	// Auth routes
	r.Route("/auth", func(r chi.Router) {
		r.Get("/github", h.Auth.HandleGitHubLogin)
		r.Get("/github/callback", h.Auth.HandleGitHubCallback)
		r.Post("/logout", h.Auth.HandleLogout)
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		if rateLimiter != nil {
			r.Use(rateLimiter.Limit)
		}

		r.Get("/skills", h.Skill.APIListSkills)
		r.Get("/skills/{slug}", h.Skill.APIGetSkill)
		r.Get("/search", h.Search.APISearch)
		r.Get("/trending", h.Skill.APIGetTrending)
		r.Get("/stats", h.Skill.APIGetStats)

		// Protected bookmark routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth)
			r.Post("/bookmarks/{skill_id}", h.Bookmark.APIToggleBookmark)
			r.Get("/bookmarks", h.Bookmark.APIListBookmarks)
		})
	})
}

func SetupRouter(h *Handlers, authService service.AuthService, rateLimiter *middleware.RateLimiter) http.Handler {
	r := chi.NewRouter()
	RegisterRoutes(r, h, authService, rateLimiter)
	return r
}
