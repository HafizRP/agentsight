package scraper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

var (
	// HTML parsing regexes for website card extraction
	cardTitleRegex   = regexp.MustCompile(`(?i)<h[234][^>]*>(.*?)</h[234]>`)
	cardDescRegex    = regexp.MustCompile(`(?i)<p[^>]*>(.*?)</p>`)
	cardLinkRegex    = regexp.MustCompile(`(?i)<a[^>]+href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	cardCodeRegex    = regexp.MustCompile(`(?i)<code[^>]*>(.*?)</code>`)
	stripTagsRegex   = regexp.MustCompile(`<[^>]*>`)
	whitespaceFilter = regexp.MustCompile(`\s+`)
)

// WebsiteTarget defines a web directory to scrape.
type WebsiteTarget struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Platform string `json:"platform"`
	Category string `json:"category"`
}

// DefaultWebsiteTargets defines standard web directories.
var DefaultWebsiteTargets = []WebsiteTarget{
	{
		Name:     "Cursor Directory",
		URL:      "https://cursor.directory",
		Platform: "cursor",
		Category: "rules",
	},
	{
		Name:     "MCP Servers Org",
		URL:      "https://mcpservers.org",
		Platform: "mcp",
		Category: "mcp_servers",
	},
}

// WebsiteScraper uses headless Chromium via chromedp to scrape dynamic JavaScript web directories.
type WebsiteScraper struct {
	chromeBin string
	targets   []WebsiteTarget
	fetcher   *Fetcher
	client    *http.Client
}

// NewWebsiteScraper creates a new WebsiteScraper instance.
func NewWebsiteScraper(chromeBin string, fetcher *Fetcher) *WebsiteScraper {
	if chromeBin == "" {
		chromeBin = os.Getenv("CHROME_BIN")
		if chromeBin == "" {
			chromeBin = "/usr/bin/chromium-browser"
		}
	}
	if fetcher == nil {
		fetcher = NewFetcher(nil)
	}

	return &WebsiteScraper{
		chromeBin: chromeBin,
		targets:   DefaultWebsiteTargets,
		fetcher:   fetcher,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetTargets overrides directory scraping targets.
func (s *WebsiteScraper) SetTargets(targets []WebsiteTarget) {
	s.targets = targets
}

// SetChromeBin overrides the chromium binary path.
func (s *WebsiteScraper) SetChromeBin(bin string) {
	s.chromeBin = bin
}

// Name identifies this scraper adapter.
func (s *WebsiteScraper) Name() string {
	return "website_chromedp"
}

// Discover scrapes all configured website targets.
func (s *WebsiteScraper) Discover(ctx context.Context) ([]RawSkill, error) {
	var allSkills []RawSkill

	for _, target := range s.targets {
		select {
		case <-ctx.Done():
			return allSkills, ctx.Err()
		default:
		}

		slog.Info("scraping website target", "target", target.Name, "url", target.URL)

		var htmlContent string
		var scrapeErr error

		// Attempt chromedp headless scraping with container-safe flags
		if s.isChromiumAvailable() {
			htmlContent, scrapeErr = s.scrapeWithChromedp(ctx, target.URL)
			if scrapeErr != nil {
				slog.Warn("chromedp scraping failed, falling back to HTTP fetch", "target", target.Name, "err", scrapeErr)
			}
		} else {
			slog.Info("chromium executable not found, using HTTP fetch fallback", "bin", s.chromeBin, "target", target.Name)
		}

		// Fallback to HTTP fetcher if chromedp was unavailable or failed
		if htmlContent == "" {
			var fetchErr error
			htmlContent, fetchErr = s.fetchWithRetry(ctx, target.URL)
			if fetchErr != nil {
				slog.Warn("HTTP fetch fallback also failed for website target", "target", target.Name, "err", fetchErr)
				continue
			}
		}

		// Parse extracted HTML into RawSkills
		skills := s.parseHTMLContent(htmlContent, target)
		slog.Info("scraped website target completed", "target", target.Name, "skills_found", len(skills))
		allSkills = append(allSkills, skills...)
	}

	return allSkills, nil
}

func (s *WebsiteScraper) isChromiumAvailable() bool {
	if _, err := os.Stat(s.chromeBin); err == nil {
		return true
	}
	if _, err := exec.LookPath(s.chromeBin); err == nil {
		return true
	}
	return false
}

func (s *WebsiteScraper) scrapeWithChromedp(ctx context.Context, targetURL string) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
	)

	if s.chromeBin != "" {
		opts = append(opts, chromedp.ExecPath(s.chromeBin))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(taskCtx, 30*time.Second)
	defer timeoutCancel()

	var htmlContent string
	var lastErr error
	backoff := 100 * time.Millisecond

	for attempt := 1; attempt <= 3; attempt++ {
		err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(targetURL),
			chromedp.Sleep(2*time.Second),
			chromedp.OuterHTML("html", &htmlContent),
		)
		if err == nil && len(htmlContent) > 0 {
			return htmlContent, nil
		}
		lastErr = err
		slog.Warn("chromedp run attempt failed", "url", targetURL, "attempt", attempt, "err", err)
		select {
		case <-timeoutCtx.Done():
			return "", timeoutCtx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}

	return "", fmt.Errorf("chromedp failed after 3 attempts: %w", lastErr)
}

func (s *WebsiteScraper) fetchWithRetry(ctx context.Context, targetURL string) (string, error) {
	var lastErr error
	backoff := 100 * time.Millisecond

	for attempt := 1; attempt <= 3; attempt++ {
		reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		content, err := s.fetcher.FetchString(reqCtx, targetURL)
		cancel()
		if err == nil && len(content) > 0 {
			return content, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", errors.New("empty response body received")
}

func (s *WebsiteScraper) parseHTMLContent(html string, target WebsiteTarget) []RawSkill {
	var skills []RawSkill
	seenNames := make(map[string]struct{})

	// Extract titles from headings
	titleMatches := cardTitleRegex.FindAllStringSubmatch(html, -1)
	descMatches := cardDescRegex.FindAllStringSubmatch(html, -1)
	codeMatches := cardCodeRegex.FindAllStringSubmatch(html, -1)

	for i, tm := range titleMatches {
		if len(tm) < 2 {
			continue
		}
		rawTitle := cleanHTMLText(tm[1])
		if len(rawTitle) < 3 || len(rawTitle) > 100 {
			continue
		}
		if _, seen := seenNames[rawTitle]; seen {
			continue
		}
		seenNames[rawTitle] = struct{}{}

		// Pair description if available
		desc := ""
		if i < len(descMatches) && len(descMatches[i]) > 1 {
			desc = cleanHTMLText(descMatches[i][1])
		}
		if desc == "" {
			desc = fmt.Sprintf("%s skill for %s", rawTitle, target.Platform)
		}

		// Pair snippet if available
		snippet := ""
		if i < len(codeMatches) && len(codeMatches[i]) > 1 {
			snippet = cleanHTMLText(codeMatches[i][1])
		}

		slug := strings.ToLower(strings.ReplaceAll(rawTitle, " ", "-"))

		skill := RawSkill{
			Name:           rawTitle,
			Description:    desc,
			SourceURL:      target.URL,
			RepoURL:        target.URL,
			Platform:       target.Platform,
			Category:       target.Category,
			ContentPreview: desc,
			InstallSnippet: snippet,
		}

		if skill.InstallSnippet == "" {
			if target.Platform == "cursor" {
				skill.InstallSnippet = fmt.Sprintf("curl -o .cursor/rules/%s.mdc %s/rules/%s", slug, target.URL, slug)
			} else if target.Platform == "mcp" {
				skill.InstallSnippet = fmt.Sprintf("npx -y @modelcontextprotocol/server-%s", slug)
			}
		}

		skills = append(skills, skill)
	}

	return skills
}

func cleanHTMLText(s string) string {
	stripped := stripTagsRegex.ReplaceAllString(s, "")
	clean := strings.ReplaceAll(stripped, "&amp;", "&")
	clean = strings.ReplaceAll(clean, "&lt;", "<")
	clean = strings.ReplaceAll(clean, "&gt;", ">")
	clean = strings.ReplaceAll(clean, "&quot;", "\"")
	clean = strings.ReplaceAll(clean, "&#39;", "'")
	clean = whitespaceFilter.ReplaceAllString(clean, " ")
	return strings.TrimSpace(clean)
}
