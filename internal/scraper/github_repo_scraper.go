package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// RepoTarget defines a target repository to scrape.
type RepoTarget struct {
	Owner         string `json:"owner"`
	Repo          string `json:"repo"`
	DefaultBranch string `json:"default_branch"`
	ParseType     string `json:"parse_type"` // "cursorrules", "mcp_servers", "agents_md", "awesome_list"
	FilePath      string `json:"file_path"`
}

// GitHubRepoMetadata holds basic repository stats.
type GitHubRepoMetadata struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	HTMLURL         string    `json:"html_url"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DefaultRepoTargets provides the initial target repositories.
var DefaultRepoTargets = []RepoTarget{
	{
		Owner:         "PatrickJS",
		Repo:          "awesome-cursorrules",
		DefaultBranch: "main",
		ParseType:     "cursorrules",
		FilePath:      "README.md",
	},
	{
		Owner:         "punkpeye",
		Repo:          "awesome-mcp-servers",
		DefaultBranch: "main",
		ParseType:     "mcp_servers",
		FilePath:      "README.md",
	},
	{
		Owner:         "modelcontextprotocol",
		Repo:          "servers",
		DefaultBranch: "main",
		ParseType:     "mcp_directory",
		FilePath:      "README.md",
	},
	{
		Owner:         "agentsmd",
		Repo:          "agents.md",
		DefaultBranch: "main",
		ParseType:     "agents_md",
		FilePath:      "README.md",
	},
	{
		Owner:         "e2b-dev",
		Repo:          "awesome-ai-agents",
		DefaultBranch: "main",
		ParseType:     "awesome_list",
		FilePath:      "README.md",
	},
}

// GitHubRepoScraper parses curated awesome-lists and specific agent repositories.
type GitHubRepoScraper struct {
	client   *http.Client
	token    string
	baseURL  string
	rawURL   string
	targets  []RepoTarget
	fetcher  *Fetcher
	parser   *AwesomeListParser
}

// NewGitHubRepoScraper creates a new GitHubRepoScraper instance.
func NewGitHubRepoScraper(client *http.Client, token string, fetcher *Fetcher) *GitHubRepoScraper {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if fetcher == nil {
		fetcher = NewFetcher(client)
	}

	return &GitHubRepoScraper{
		client:   client,
		token:    token,
		baseURL:  "https://api.github.com",
		rawURL:   "https://raw.githubusercontent.com",
		targets:  DefaultRepoTargets,
		fetcher:  fetcher,
		parser:   NewAwesomeListParser(),
	}
}

// SetBaseURL overrides the GitHub API URL for testing.
func (s *GitHubRepoScraper) SetBaseURL(baseURL string) {
	s.baseURL = strings.TrimRight(baseURL, "/")
}

// SetRawURL overrides the raw GitHub content URL for testing.
func (s *GitHubRepoScraper) SetRawURL(rawURL string) {
	s.rawURL = strings.TrimRight(rawURL, "/")
}

// SetTargets overrides target repositories.
func (s *GitHubRepoScraper) SetTargets(targets []RepoTarget) {
	s.targets = targets
}

// Name identifies this scraper adapter.
func (s *GitHubRepoScraper) Name() string {
	return "github_repo"
}

// FetchRepoMetadata queries the GitHub API for repo stars and forks.
func (s *GitHubRepoScraper) FetchRepoMetadata(ctx context.Context, owner, repo string) (*GitHubRepoMetadata, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s", s.baseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create repo metadata request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "AgentSight-Scraper/1.0 (+https://github.com/agentsight)")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("repo metadata request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		slog.Warn("GitHub API rate limit hit while fetching repo metadata", "owner", owner, "repo", repo)
		return &GitHubRepoMetadata{
			Name:     repo,
			FullName: fmt.Sprintf("%s/%s", owner, repo),
			HTMLURL:  fmt.Sprintf("https://github.com/%s/%s", owner, repo),
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for %s/%s", resp.StatusCode, owner, repo)
	}

	var meta GitHubRepoMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("failed to decode repo metadata: %w", err)
	}

	return &meta, nil
}

// Discover executes scraping across all configured repository targets.
func (s *GitHubRepoScraper) Discover(ctx context.Context) ([]RawSkill, error) {
	var allSkills []RawSkill
	seenMap := make(map[string]struct{})

	for _, target := range s.targets {
		select {
		case <-ctx.Done():
			return allSkills, ctx.Err()
		default:
		}

		slog.Info("scraping target repository", "owner", target.Owner, "repo", target.Repo, "type", target.ParseType)

		// 1. Fetch Repo Metadata (stars, forks)
		meta, err := s.FetchRepoMetadata(ctx, target.Owner, target.Repo)
		if err != nil {
			slog.Warn("failed to fetch repo metadata, using defaults", "owner", target.Owner, "repo", target.Repo, "err", err)
			meta = &GitHubRepoMetadata{
				Name:     target.Repo,
				FullName: fmt.Sprintf("%s/%s", target.Owner, target.Repo),
				HTMLURL:  fmt.Sprintf("https://github.com/%s/%s", target.Owner, target.Repo),
			}
		}

		// 2. Fetch Target File (README.md)
		branch := target.DefaultBranch
		if branch == "" {
			branch = "main"
		}
		filePath := target.FilePath
		if filePath == "" {
			filePath = "README.md"
		}

		fileURL := fmt.Sprintf("%s/%s/%s/%s/%s", s.rawURL, target.Owner, target.Repo, branch, filePath)
		content, err := s.fetcher.FetchString(ctx, fileURL)
		if err != nil {
			// Try fallback to "master" branch
			if branch == "main" {
				fallbackURL := fmt.Sprintf("%s/%s/%s/master/%s", s.rawURL, target.Owner, target.Repo, filePath)
				content, err = s.fetcher.FetchString(ctx, fallbackURL)
			}
			if err != nil {
				slog.Warn("failed to fetch target file content", "url", fileURL, "err", err)
				continue
			}
		}

		// 3. Parse content based on ParseType
		var parsedSkills []RawSkill

		switch target.ParseType {
		case "cursorrules":
			parsedSkills = s.parseCursorrulesRepo(content, target, meta)
		case "mcp_servers":
			parsedSkills = s.parseMCPServersRepo(content, target, meta)
		case "mcp_directory":
			parsedSkills = s.parseMCPDirectoryRepo(content, target, meta)
		case "agents_md":
			parsedSkills = s.parseAgentsMDRepo(content, target, meta)
		default:
			parsedSkills = s.parser.ParseToRawSkills(content, target.Owner, target.Repo, meta.StargazersCount, meta.ForksCount)
		}

		for _, skill := range parsedSkills {
			key := fmt.Sprintf("%s:%s", skill.Name, skill.SourceURL)
			if _, exists := seenMap[key]; exists {
				continue
			}
			seenMap[key] = struct{}{}
			allSkills = append(allSkills, skill)
		}
	}

	return allSkills, nil
}

func (s *GitHubRepoScraper) parseCursorrulesRepo(content string, target RepoTarget, meta *GitHubRepoMetadata) []RawSkill {
	skills := s.parser.ParseToRawSkills(content, target.Owner, target.Repo, meta.StargazersCount, meta.ForksCount)
	for i := range skills {
		skills[i].Platform = "cursor"
		skills[i].Category = "rules"
		if skills[i].FilePath == "" {
			skills[i].FilePath = fmt.Sprintf(".cursor/rules/%s.mdc", strings.ToLower(skills[i].Name))
		}
		if skills[i].InstallSnippet == "" {
			skills[i].InstallSnippet = fmt.Sprintf("curl -o %s %s", skills[i].FilePath, skills[i].SourceURL)
		}
	}
	return skills
}

func (s *GitHubRepoScraper) parseMCPServersRepo(content string, target RepoTarget, meta *GitHubRepoMetadata) []RawSkill {
	skills := s.parser.ParseToRawSkills(content, target.Owner, target.Repo, meta.StargazersCount, meta.ForksCount)
	for i := range skills {
		skills[i].Platform = "mcp"
		skills[i].Category = "mcp_servers"
		if skills[i].InstallSnippet == "" && skills[i].RepoName != "" {
			cleanRepo := strings.TrimPrefix(skills[i].RepoName, "server-")
			skills[i].InstallSnippet = fmt.Sprintf("npx -y @modelcontextprotocol/server-%s", cleanRepo)
		}
	}
	return skills
}

func (s *GitHubRepoScraper) parseMCPDirectoryRepo(content string, target RepoTarget, meta *GitHubRepoMetadata) []RawSkill {
	skills := s.parser.ParseToRawSkills(content, target.Owner, target.Repo, meta.StargazersCount, meta.ForksCount)
	for i := range skills {
		skills[i].Platform = "mcp"
		skills[i].Category = "mcp_servers"
		skills[i].RepoOwner = target.Owner
		skills[i].RepoName = target.Repo
		skills[i].RepoURL = fmt.Sprintf("https://github.com/%s/%s", target.Owner, target.Repo)
		if skills[i].InstallSnippet == "" {
			cleanName := strings.ToLower(strings.TrimSpace(skills[i].Name))
			skills[i].InstallSnippet = fmt.Sprintf("npx -y @modelcontextprotocol/server-%s", cleanName)
		}
	}
	return skills
}

func (s *GitHubRepoScraper) parseAgentsMDRepo(content string, target RepoTarget, meta *GitHubRepoMetadata) []RawSkill {
	skills := s.parser.ParseToRawSkills(content, target.Owner, target.Repo, meta.StargazersCount, meta.ForksCount)
	for i := range skills {
		skills[i].Platform = "claude"
		skills[i].Category = "skills"
		if skills[i].FilePath == "" {
			skills[i].FilePath = "AGENTS.md"
		}
		if skills[i].InstallSnippet == "" {
			skills[i].InstallSnippet = fmt.Sprintf("curl -o AGENTS.md %s", skills[i].SourceURL)
		}
	}
	return skills
}
