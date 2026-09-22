package scraper_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"agentsight/internal/config"
	"agentsight/internal/models"
	"agentsight/internal/repository"
	"agentsight/internal/scraper"
)

func TestGitHubSearchScraper_Discover(t *testing.T) {
	// Mock GitHub API Server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-github-token" {
			t.Errorf("expected Authorization header 'Bearer test-github-token', got %q", authHeader)
		}

		if r.URL.Path == "/search/code" {
			_ = r.URL.Query().Get("q")
			page := r.URL.Query().Get("page")

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-RateLimit-Remaining", "4999")

			if page == "1" {
				resp := scraper.GitHubCodeSearchResult{
					TotalCount: 1,
					Items: []scraper.GitHubCodeItem{
						{
							Name:    ".cursorrules",
							Path:    ".cursorrules",
							SHA:     "abc12345",
							HTMLURL: "https://github.com/test-owner/test-repo/blob/main/.cursorrules",
							Repository: scraper.GitHubRepositoryItem{
								ID:          1001,
								Name:        "test-repo",
								FullName:    "test-owner/test-repo",
								Owner: struct {
									Login     string `json:"login"`
									AvatarURL string `json:"avatar_url"`
								}{
									Login: "test-owner",
								},
								Description: "Awesome test repository for agent skills",
								HTMLURL:     "https://github.com/test-owner/test-repo",
								StargazersCount: 1250,
								ForksCount:      80,
							},
						},
					},
				}
				_ = json.NewEncoder(w).Encode(resp)
			} else {
				// Empty page to terminate pagination
				_ = json.NewEncoder(w).Encode(scraper.GitHubCodeSearchResult{
					TotalCount: 1,
					Items:      []scraper.GitHubCodeItem{},
				})
			}
			return
		}

		http.NotFound(w, r)
	}))
	defer apiServer.Close()

	// Mock Raw GitHub Content Server
	rawServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Cursor Rules\nAlways use TypeScript strict mode."))
	}))
	defer rawServer.Close()

	fetcher := scraper.NewFetcher(apiServer.Client())
	s := scraper.NewGitHubSearchScraper(apiServer.Client(), "test-github-token", fetcher)
	s.SetBaseURL(apiServer.URL)
	s.SetRawURL(rawServer.URL)
	s.SetQueries([]string{"filename:.cursorrules"})
	s.SetMaxPages(2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	skills, err := s.Discover(ctx)
	if err != nil {
		t.Fatalf("Discover returned unexpected error: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("expected 1 discovered skill, got %d", len(skills))
	}

	skill := skills[0]
	if skill.RepoOwner != "test-owner" {
		t.Errorf("expected RepoOwner 'test-owner', got %q", skill.RepoOwner)
	}
	if skill.RepoName != "test-repo" {
		t.Errorf("expected RepoName 'test-repo', got %q", skill.RepoName)
	}
	if skill.StarsCount != 1250 {
		t.Errorf("expected 1250 stars, got %d", skill.StarsCount)
	}
	if skill.ContentRaw == "" {
		t.Errorf("expected non-empty ContentRaw fetched from raw server")
	}
}

func TestGitHubSearchScraper_RateLimitGraceful(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer apiServer.Close()

	s := scraper.NewGitHubSearchScraper(apiServer.Client(), "", nil)
	s.SetBaseURL(apiServer.URL)
	s.SetQueries([]string{"filename:.cursorrules"})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	skills, err := s.Discover(ctx)
	if err != nil {
		t.Fatalf("expected no error on rate limit degradation, got: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("expected 0 skills on rate limited server, got %d", len(skills))
	}
}

func TestGitHubRepoScraper_Discover(t *testing.T) {
	// Mock GitHub API Server for repo stats
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/PatrickJS/awesome-cursorrules" {
			meta := scraper.GitHubRepoMetadata{
				ID:              2001,
				Name:            "awesome-cursorrules",
				FullName:        "PatrickJS/awesome-cursorrules",
				Description:     "Curated list of Cursor rules",
				HTMLURL:         "https://github.com/PatrickJS/awesome-cursorrules",
				StargazersCount: 4500,
				ForksCount:      320,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(meta)
			return
		}
		http.NotFound(w, r)
	}))
	defer apiServer.Close()

	// Mock Raw Content Server for README.md
	rawServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readme := `## Rules
- [Next.js App Router](https://github.com/PatrickJS/awesome-cursorrules/blob/main/rules/nextjs.mdc) - Best practices for Next.js 14.
- [FastAPI Python](https://github.com/PatrickJS/awesome-cursorrules/blob/main/rules/fastapi.mdc) - Clean async FastAPI endpoints.
`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(readme))
	}))
	defer rawServer.Close()

	fetcher := scraper.NewFetcher(apiServer.Client())
	repoScraper := scraper.NewGitHubRepoScraper(apiServer.Client(), "fake-token", fetcher)
	repoScraper.SetBaseURL(apiServer.URL)
	repoScraper.SetRawURL(rawServer.URL)
	repoScraper.SetTargets([]scraper.RepoTarget{
		{
			Owner:         "PatrickJS",
			Repo:          "awesome-cursorrules",
			DefaultBranch: "main",
			ParseType:     "cursorrules",
			FilePath:      "README.md",
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	skills, err := repoScraper.Discover(ctx)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if len(skills) != 2 {
		t.Fatalf("expected 2 discovered skills, got %d", len(skills))
	}

	if skills[0].StarsCount != 4500 {
		t.Errorf("expected 4500 stars, got %d", skills[0].StarsCount)
	}
	if skills[0].Platform != "cursor" {
		t.Errorf("expected platform 'cursor', got %q", skills[0].Platform)
	}
}

func TestWebsiteScraper_FallbackHTML(t *testing.T) {
	// Mock web directory HTML server
	mockDirectory := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head><title>Cursor Directory</title></head>
<body>
  <div class="card">
    <h3>React 19 & Next.js Guidelines</h3>
    <p>Comprehensive standards for React Server Actions and hooks.</p>
    <code>curl -o .cursor/rules/react-19.mdc https://cursor.directory/react-19</code>
  </div>
  <div class="card">
    <h3>Tailwind CSS Dark Mode</h3>
    <p>Color scale and responsiveness guidelines.</p>
    <code>curl -o .cursor/rules/tailwind.mdc https://cursor.directory/tailwind</code>
  </div>
</body>
</html>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer mockDirectory.Close()

	fetcher := scraper.NewFetcher(mockDirectory.Client())
	webScraper := scraper.NewWebsiteScraper("/non/existent/chromium-bin", fetcher)
	webScraper.SetTargets([]scraper.WebsiteTarget{
		{
			Name:     "Test Directory",
			URL:      mockDirectory.URL,
			Platform: "cursor",
			Category: "rules",
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	skills, err := webScraper.Discover(ctx)
	if err != nil {
		t.Fatalf("Discover returned unexpected error: %v", err)
	}

	if len(skills) != 2 {
		t.Fatalf("expected 2 skills extracted via HTML fallback, got %d", len(skills))
	}

	if skills[0].Name != "React 19 & Next.js Guidelines" {
		t.Errorf("expected skill title 'React 19 & Next.js Guidelines', got %q", skills[0].Name)
	}
	if skills[0].Platform != "cursor" {
		t.Errorf("expected platform 'cursor', got %q", skills[0].Platform)
	}
}

// In-Memory Mock Repositories for Scheduler Testing

type mockSourceRepo struct {
	mu      sync.Mutex
	sources []models.Source
}

func (m *mockSourceRepo) Create(ctx context.Context, source *models.Source) (*models.Source, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	source.ID = len(m.sources) + 1
	m.sources = append(m.sources, *source)
	return source, nil
}

func (m *mockSourceRepo) GetByID(ctx context.Context, id int) (*models.Source, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sources {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("source %d not found", id)
}

func (m *mockSourceRepo) List(ctx context.Context) ([]models.Source, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sources, nil
}

func (m *mockSourceRepo) ListActive(ctx context.Context) ([]models.Source, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var active []models.Source
	for _, s := range m.sources {
		if s.IsActive {
			active = append(active, s)
		}
	}
	return active, nil
}

func (m *mockSourceRepo) Update(ctx context.Context, source *models.Source) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.sources {
		if m.sources[i].ID == source.ID {
			m.sources[i] = *source
			return nil
		}
	}
	return fmt.Errorf("source %d not found", source.ID)
}

func (m *mockSourceRepo) UpdateLastScraped(ctx context.Context, id int, t time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.sources {
		if m.sources[i].ID == id {
			m.sources[i].LastScrapedAt = &t
			return nil
		}
	}
	return fmt.Errorf("source %d not found", id)
}

type mockSkillRepo struct {
	mu     sync.Mutex
	skills map[string]*models.Skill
}

func newMockSkillRepo() *mockSkillRepo {
	return &mockSkillRepo{
		skills: make(map[string]*models.Skill),
	}
}

func (m *mockSkillRepo) Create(ctx context.Context, skill *models.Skill) (*models.Skill, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	skill.ID = len(m.skills) + 1
	skill.CreatedAt = time.Now()
	skill.UpdatedAt = time.Now()
	clone := *skill
	m.skills[skill.Slug] = &clone
	return skill, nil
}

func (m *mockSkillRepo) GetByID(ctx context.Context, id int) (*models.Skill, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.skills {
		if s.ID == id {
			clone := *s
			return &clone, nil
		}
	}
	return nil, fmt.Errorf("skill %d not found", id)
}

func (m *mockSkillRepo) GetBySlug(ctx context.Context, slug string) (*models.Skill, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, exists := m.skills[slug]; exists {
		clone := *s
		return &clone, nil
	}
	return nil, fmt.Errorf("skill with slug %q not found", slug)
}

func (m *mockSkillRepo) Update(ctx context.Context, skill *models.Skill) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.skills[skill.Slug]; exists {
		skill.UpdatedAt = time.Now()
		clone := *skill
		m.skills[skill.Slug] = &clone
		return nil
	}
	return fmt.Errorf("skill %q not found", skill.Slug)
}

func (m *mockSkillRepo) Delete(ctx context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for slug, s := range m.skills {
		if s.ID == id {
			delete(m.skills, slug)
			return nil
		}
	}
	return nil
}

func (m *mockSkillRepo) List(ctx context.Context, filter models.SkillFilter) ([]models.Skill, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []models.Skill
	for _, s := range m.skills {
		list = append(list, *s)
	}
	return list, len(list), nil
}

func (m *mockSkillRepo) Search(ctx context.Context, params models.SearchParams) ([]models.Skill, int, error) {
	return m.List(ctx, models.SkillFilter{})
}

func (m *mockSkillRepo) GetTrending(ctx context.Context, limit int) ([]models.Skill, error) {
	list, _, _ := m.List(ctx, models.SkillFilter{})
	return list, nil
}

func (m *mockSkillRepo) ListByCategory(ctx context.Context, category string, limit, offset int) ([]models.Skill, int, error) {
	return m.List(ctx, models.SkillFilter{})
}

func (m *mockSkillRepo) ListByPlatform(ctx context.Context, platform string, limit, offset int) ([]models.Skill, int, error) {
	return m.List(ctx, models.SkillFilter{})
}

func (m *mockSkillRepo) GetRelated(ctx context.Context, skill *models.Skill, limit int) ([]models.Skill, error) {
	return []models.Skill{}, nil
}

func (m *mockSkillRepo) GetStats(ctx context.Context) (*models.Stats, error) {
	return &models.Stats{TotalSkills: len(m.skills)}, nil
}

type mockScrapeLogRepo struct {
	mu   sync.Mutex
	logs []models.ScrapeLog
}

func (m *mockScrapeLogRepo) Create(ctx context.Context, log *models.ScrapeLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	log.ID = len(m.logs) + 1
	m.logs = append(m.logs, *log)
	return nil
}

func (m *mockScrapeLogRepo) ListRecent(ctx context.Context, limit int) ([]models.ScrapeLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.logs, nil
}

type dummyScraper struct {
	name   string
	skills []scraper.RawSkill
}

func (d *dummyScraper) Discover(ctx context.Context) ([]scraper.RawSkill, error) {
	return d.skills, nil
}

func (d *dummyScraper) Name() string {
	return d.name
}

func TestScheduler_CheckAndRunDueSources(t *testing.T) {
	sourceRepo := &mockSourceRepo{
		sources: []models.Source{
			{
				ID:                  1,
				Name:                "Test Due Source",
				URL:                 "https://example.com/source",
				SourceType:          "test_type",
				ScrapeStrategy:      "test_strategy",
				IsActive:            true,
				LastScrapedAt:       nil, // Due immediately
				ScrapeIntervalHours: 24,
			},
		},
	}

	skillRepo := newMockSkillRepo()
	scrapeLogRepo := &mockScrapeLogRepo{}

	repos := &repository.Repositories{
		Sources:    sourceRepo,
		Skills:     skillRepo,
		ScrapeLogs: scrapeLogRepo,
	}

	cfg := &config.Config{
		Env: "testing",
	}

	sched := scraper.NewScheduler(repos, cfg)

	// Register dummy scraper that returns 2 skills
	dummy := &dummyScraper{
		name: "test_type",
		skills: []scraper.RawSkill{
			{
				Name:        "Skill Alpha",
				Description: "First test skill",
				SourceURL:   "https://example.com/alpha",
				Platform:    "cursor",
				Category:    "rules",
				StarsCount:  100,
			},
			{
				Name:        "Skill Beta",
				Description: "Second test skill",
				SourceURL:   "https://example.com/beta",
				Platform:    "claude",
				Category:    "skills",
				StarsCount:  200,
			},
		},
	}
	sched.RegisterScraper("test_type", dummy)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1st Run: Expect 2 new skills created
	err := sched.CheckAndRunDueSources(ctx)
	if err != nil {
		t.Fatalf("CheckAndRunDueSources failed: %v", err)
	}

	if len(skillRepo.skills) != 2 {
		t.Fatalf("expected 2 skills in repository, got %d", len(skillRepo.skills))
	}
	if len(scrapeLogRepo.logs) != 1 {
		t.Fatalf("expected 1 scrape log recorded, got %d", len(scrapeLogRepo.logs))
	}
	if scrapeLogRepo.logs[0].SkillsNew != 2 {
		t.Errorf("expected 2 SkillsNew in log, got %d", scrapeLogRepo.logs[0].SkillsNew)
	}

	// 2nd Run with updated stars:
	// Update dummy scraper with increased stars
	dummy.skills[0].StarsCount = 150 // +50 velocity
	// Make source due again
	sourceRepo.sources[0].LastScrapedAt = nil

	err = sched.CheckAndRunDueSources(ctx)
	if err != nil {
		t.Fatalf("second CheckAndRunDueSources failed: %v", err)
	}

	if len(scrapeLogRepo.logs) != 2 {
		t.Fatalf("expected 2 scrape logs recorded, got %d", len(scrapeLogRepo.logs))
	}

	// Verify stars velocity was updated
	updatedSkill, _ := skillRepo.GetBySlug(ctx, "skill-alpha")
	if updatedSkill.StarsCount != 150 {
		t.Errorf("expected 150 stars after update, got %d", updatedSkill.StarsCount)
	}
	if updatedSkill.StarsVelocity != 50 {
		t.Errorf("expected stars velocity 50, got %d", updatedSkill.StarsVelocity)
	}
}
