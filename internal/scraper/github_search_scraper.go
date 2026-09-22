package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultSearchQueries lists standard agent rule and skill config patterns on GitHub.
var DefaultSearchQueries = []string{
	"filename:.cursorrules",
	"filename:AGENTS.md",
	"filename:CLAUDE.md",
	"filename:SKILL.md path:.gemini",
	"filename:mcp_config.json",
}

// GitHubCodeSearchResult represents the GitHub code search JSON response.
type GitHubCodeSearchResult struct {
	TotalCount        int              `json:"total_count"`
	IncompleteResults bool             `json:"incomplete_results"`
	Items             []GitHubCodeItem `json:"items"`
}

// GitHubCodeItem represents a single matching file in a repository.
type GitHubCodeItem struct {
	Name       string               `json:"name"`
	Path       string               `json:"path"`
	SHA        string               `json:"sha"`
	URL        string               `json:"url"`
	HTMLURL    string               `json:"html_url"`
	Repository GitHubRepositoryItem `json:"repository"`
}

// GitHubRepositoryItem contains repository metadata inside code search results.
type GitHubRepositoryItem struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	FullName        string `json:"full_name"`
	Owner           struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"owner"`
	HTMLURL         string `json:"html_url"`
	Description     string `json:"description"`
	Fork            bool   `json:"fork"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
}

// GitHubSearchScraper queries GitHub's Code Search API to discover agent skill files.
type GitHubSearchScraper struct {
	client   *http.Client
	token    string
	baseURL  string
	rawURL   string
	queries  []string
	maxPages int
	perPage  int
	fetcher  *Fetcher
}

// NewGitHubSearchScraper creates a configured GitHubSearchScraper.
func NewGitHubSearchScraper(client *http.Client, token string, fetcher *Fetcher) *GitHubSearchScraper {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if fetcher == nil {
		fetcher = NewFetcher(client)
	}

	return &GitHubSearchScraper{
		client:   client,
		token:    token,
		baseURL:  "https://api.github.com",
		rawURL:   "https://raw.githubusercontent.com",
		queries:  DefaultSearchQueries,
		maxPages: 10,
		perPage:  100,
		fetcher:  fetcher,
	}
}

// SetBaseURL overrides the GitHub API base URL (useful for integration tests).
func (s *GitHubSearchScraper) SetBaseURL(baseURL string) {
	s.baseURL = strings.TrimRight(baseURL, "/")
}

// SetRawURL overrides the raw GitHub content URL.
func (s *GitHubSearchScraper) SetRawURL(rawURL string) {
	s.rawURL = strings.TrimRight(rawURL, "/")
}

// SetQueries overrides default search queries.
func (s *GitHubSearchScraper) SetQueries(queries []string) {
	s.queries = queries
}

// SetMaxPages limits pagination depth.
func (s *GitHubSearchScraper) SetMaxPages(pages int) {
	s.maxPages = pages
}

// Name identifies this scraper adapter.
func (s *GitHubSearchScraper) Name() string {
	return "github_search"
}

// Discover executes all search queries and returns discovered skills.
func (s *GitHubSearchScraper) Discover(ctx context.Context) ([]RawSkill, error) {
	var discovered []RawSkill
	seenMap := make(map[string]struct{})

	for _, query := range s.queries {
		slog.Info("starting GitHub code search query", "query", query)

		for page := 1; page <= s.maxPages; page++ {
			select {
			case <-ctx.Done():
				return discovered, ctx.Err()
			default:
			}

			endpoint := fmt.Sprintf("%s/search/code?q=%s&page=%d&per_page=%d",
				s.baseURL, url.QueryEscape(query), page, s.perPage)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
			if err != nil {
				slog.Error("failed to create code search request", "err", err, "query", query)
				break
			}

			req.Header.Set("Accept", "application/vnd.github.v3+json")
			req.Header.Set("User-Agent", "AgentSight-Scraper/1.0 (+https://github.com/agentsight)")
			if s.token != "" {
				req.Header.Set("Authorization", "Bearer "+s.token)
			}

			resp, err := s.client.Do(req)
			if err != nil {
				slog.Warn("GitHub code search request failed", "query", query, "err", err)
				break
			}

			// Rate limit inspection
			remaining := resp.Header.Get("X-RateLimit-Remaining")
			if remaining != "" {
				if remVal, err := strconv.Atoi(remaining); err == nil && remVal <= 0 {
					slog.Warn("GitHub code search rate limit reached; gracefully concluding query", "query", query)
					resp.Body.Close()
					break
				}
			}

			if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
				slog.Warn("GitHub code search rate-limited by HTTP status code", "status", resp.StatusCode, "query", query)
				resp.Body.Close()
				break
			}

			if resp.StatusCode != http.StatusOK {
				bodySample, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
				resp.Body.Close()
				slog.Warn("GitHub code search returned non-200 status", "status", resp.StatusCode, "body", string(bodySample))
				break
			}

			var result GitHubCodeSearchResult
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				slog.Error("failed to decode GitHub code search response", "err", err)
				break
			}
			resp.Body.Close()

			if len(result.Items) == 0 {
				break
			}

			for _, item := range result.Items {
				owner := item.Repository.Owner.Login
				if owner == "" && strings.Contains(item.Repository.FullName, "/") {
					parts := strings.Split(item.Repository.FullName, "/")
					owner = parts[0]
				}

				dedupKey := fmt.Sprintf("%s/%s:%s", owner, item.Repository.Name, item.Path)
				if _, exists := seenMap[dedupKey]; exists {
					continue
				}
				seenMap[dedupKey] = struct{}{}

				skillName := formatSearchSkillName(item.Repository.Name, item.Name)
				desc := item.Repository.Description
				if desc == "" {
					desc = fmt.Sprintf("%s configuration file discovered in %s", item.Name, item.Repository.FullName)
				}

				// Construct direct raw URL
				rawContentURL := fmt.Sprintf("%s/%s/%s/HEAD/%s",
					s.rawURL, owner, item.Repository.Name, item.Path)

				rawSkill := RawSkill{
					Name:           skillName,
					Description:    desc,
					SourceURL:      item.HTMLURL,
					RepoURL:        item.Repository.HTMLURL,
					RepoOwner:      owner,
					RepoName:       item.Repository.Name,
					StarsCount:     item.Repository.StargazersCount,
					ForksCount:     item.Repository.ForksCount,
					ContentPreview: desc,
					FilePath:       item.Path,
				}

				// Attempt to fetch raw content with size/timeout guard
				if s.fetcher != nil {
					fetchCtx, fetchCancel := context.WithTimeout(ctx, 10*time.Second)
					content, err := s.fetcher.FetchString(fetchCtx, rawContentURL)
					fetchCancel()
					if err == nil && content != "" {
						rawSkill.ContentRaw = content
						if len(content) > 300 {
							rawSkill.ContentPreview = content[:300] + "..."
						} else {
							rawSkill.ContentPreview = content
						}
					}
				}

				discovered = append(discovered, rawSkill)
			}

			// If fewer items returned than requested per page, we reached the end of results
			if len(result.Items) < s.perPage {
				break
			}
		}
	}

	return discovered, nil
}

func formatSearchSkillName(repoName, fileName string) string {
	cleanName := strings.TrimSuffix(repoName, ".git")
	cleanFile := strings.TrimPrefix(fileName, ".")
	if strings.Contains(strings.ToLower(cleanName), strings.ToLower(cleanFile)) {
		return cleanName
	}
	return fmt.Sprintf("%s %s", cleanName, cleanFile)
}
