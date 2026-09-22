package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"agentsight/internal/config"
	"agentsight/internal/models"
	"agentsight/internal/repository"
)

// Scheduler manages scheduled scraping intervals, executes source adapters,
// normalizes skills, persists records, and records audit logs.
type Scheduler struct {
	sourcesRepo   repository.SourceRepository
	skillsRepo    repository.SkillRepository
	scrapeLogRepo repository.ScrapeLogRepository
	normalizer    *Normalizer
	fetcher       *Fetcher
	cfg           *config.Config
	scrapers      map[string]Scraper
	pollInterval  time.Duration
	stopCh        chan struct{}
	wg            sync.WaitGroup
	running       bool
	mu            sync.Mutex
}

// NewScheduler creates a fully configured Scheduler.
func NewScheduler(repos *repository.Repositories, cfg *config.Config) *Scheduler {
	if cfg == nil {
		cfg = config.Load()
	}

	fetcher := NewFetcher(&http.Client{Timeout: 30 * time.Second})
	normalizer := NewNormalizer()

	ghSearch := NewGitHubSearchScraper(nil, cfg.GitHubToken, fetcher)
	ghRepo := NewGitHubRepoScraper(nil, cfg.GitHubToken, fetcher)
	website := NewWebsiteScraper(cfg.ChromeBin, fetcher)

	scrapers := map[string]Scraper{
		"github_search":    ghSearch,
		"github_repo":      ghRepo,
		"website_chromedp": website,
		"website":          website,
	}

	return &Scheduler{
		sourcesRepo:   repos.Sources,
		skillsRepo:    repos.Skills,
		scrapeLogRepo: repos.ScrapeLogs,
		normalizer:    normalizer,
		fetcher:       fetcher,
		cfg:           cfg,
		scrapers:      scrapers,
		pollInterval:  15 * time.Minute,
		stopCh:        make(chan struct{}),
	}
}

// RegisterScraper binds a custom scraper adapter.
func (s *Scheduler) RegisterScraper(name string, scraper Scraper) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scrapers[name] = scraper
}

// SetPollInterval overrides the polling check interval.
func (s *Scheduler) SetPollInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pollInterval = interval
}

// Start launches the background scheduler loop.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	slog.Info("starting scraper background scheduler", "interval", s.pollInterval)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		// Run an initial check after startup delay
		select {
		case <-time.After(5 * time.Second):
			if err := s.CheckAndRunDueSources(ctx); err != nil {
				slog.Warn("initial scraping cycle encountered errors", "err", err)
			}
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}

		ticker := time.NewTicker(s.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.CheckAndRunDueSources(ctx); err != nil {
					slog.Warn("periodic scraping cycle encountered errors", "err", err)
				}
			case <-s.stopCh:
				slog.Info("scraper scheduler stopped")
				return
			case <-ctx.Done():
				slog.Info("scraper scheduler context cancelled")
				return
			}
		}
	}()
}

// Stop gracefully shuts down the background scheduler loop.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	s.wg.Wait()
}

// CheckAndRunDueSources queries active sources and runs any that have exceeded their interval.
func (s *Scheduler) CheckAndRunDueSources(ctx context.Context) error {
	if s.sourcesRepo == nil {
		return nil
	}

	sources, err := s.sourcesRepo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to list active sources: %w", err)
	}

	now := time.Now()
	for _, source := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		intervalHours := source.ScrapeIntervalHours
		if intervalHours <= 0 {
			intervalHours = 24
		}

		isDue := false
		if source.LastScrapedAt == nil {
			isDue = true
		} else {
			elapsed := now.Sub(*source.LastScrapedAt)
			if elapsed >= time.Duration(intervalHours)*time.Hour {
				isDue = true
			}
		}

		if isDue {
			slog.Info("source is due for scraping", "id", source.ID, "name", source.Name, "type", source.SourceType)
			if _, err := s.RunSource(ctx, source); err != nil {
				slog.Error("error running source scraper", "source", source.Name, "err", err)
			}
		}
	}

	return nil
}

// RunSource executes the scraper adapter for a specific source record.
func (s *Scheduler) RunSource(ctx context.Context, source models.Source) (*models.ScrapeLog, error) {
	start := time.Now()
	scraper := s.resolveScraper(source)
	if scraper == nil {
		err := fmt.Errorf("no scraper adapter found for source type: %s / strategy: %s", source.SourceType, source.ScrapeStrategy)
		s.recordLog(ctx, &source.ID, "failed", 0, 0, 0, err.Error(), int(time.Since(start).Milliseconds()))
		return nil, err
	}

	rawSkills, err := scraper.Discover(ctx)
	duration := int(time.Since(start).Milliseconds())

	if err != nil {
		slog.Warn("scraper execution failed", "source", source.Name, "err", err)
		log := s.recordLog(ctx, &source.ID, "failed", 0, 0, 0, err.Error(), duration)
		return log, err
	}

	skillsFound := len(rawSkills)
	skillsNew := 0
	skillsUpdated := 0

	for _, raw := range rawSkills {
		skill, normErr := s.normalizer.Normalize(raw)
		if normErr != nil {
			slog.Debug("failed to normalize skill", "name", raw.Name, "err", normErr)
			continue
		}

		if s.skillsRepo == nil {
			continue
		}

		existing, getErr := s.skillsRepo.GetBySlug(ctx, skill.Slug)
		if getErr == nil && existing != nil {
			// Update existing skill
			skill.ID = existing.ID
			skill.CreatedAt = existing.CreatedAt
			// Velocity: stars diff
			skill.StarsVelocity = skill.StarsCount - existing.StarsCount
			if err := s.skillsRepo.Update(ctx, skill); err == nil {
				skillsUpdated++
			}
		} else {
			// Create new skill
			if _, err := s.skillsRepo.Create(ctx, skill); err == nil {
				skillsNew++
			}
		}
	}

	now := time.Now()
	if s.sourcesRepo != nil {
		_ = s.sourcesRepo.UpdateLastScraped(ctx, source.ID, now)
	}

	log := s.recordLog(ctx, &source.ID, "success", skillsFound, skillsNew, skillsUpdated, "", int(time.Since(start).Milliseconds()))
	slog.Info("source scrape complete", "source", source.Name, "found", skillsFound, "new", skillsNew, "updated", skillsUpdated)

	return log, nil
}

func (s *Scheduler) resolveScraper(source models.Source) Scraper {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Direct type match
	if sc, exists := s.scrapers[source.SourceType]; exists {
		return sc
	}

	// 2. Strategy match
	if sc, exists := s.scrapers[source.ScrapeStrategy]; exists {
		return sc
	}

	// 3. Fallbacks based on URL patterns
	if strings.Contains(source.URL, "api.github.com/search") {
		return s.scrapers["github_search"]
	}
	if strings.Contains(source.URL, "github.com") {
		return s.scrapers["github_repo"]
	}
	if strings.HasPrefix(source.URL, "http") {
		return s.scrapers["website_chromedp"]
	}

	return nil
}

func (s *Scheduler) recordLog(ctx context.Context, sourceID *int, status string, found, newCount, updatedCount int, errMsg string, durationMS int) *models.ScrapeLog {
	log := &models.ScrapeLog{
		SourceID:      sourceID,
		Status:        status,
		SkillsFound:   found,
		SkillsNew:     newCount,
		SkillsUpdated: updatedCount,
		ErrorMessage:  errMsg,
		DurationMS:    durationMS,
		CreatedAt:     time.Now(),
	}

	if s.scrapeLogRepo != nil {
		if err := s.scrapeLogRepo.Create(ctx, log); err != nil {
			slog.Warn("failed to persist scrape log", "err", err)
		}
	}

	return log
}

// ParseRepoURL parses an HTTP or GitHub URL into owner and repository.
func ParseRepoURL(rawURL string) (string, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 2 {
		return parts[0], strings.TrimSuffix(parts[1], ".git"), nil
	}
	return "", "", fmt.Errorf("invalid repository URL: %s", rawURL)
}
