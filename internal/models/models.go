package models

import "time"

type PlatformType string

const (
	PlatformCursor  PlatformType = "cursor"
	PlatformClaude  PlatformType = "claude"
	PlatformGemini  PlatformType = "gemini"
	PlatformMCP     PlatformType = "mcp"
	PlatformCopilot PlatformType = "copilot"
	PlatformGeneric PlatformType = "generic"
)

func (p PlatformType) IsValid() bool {
	switch p {
	case PlatformCursor, PlatformClaude, PlatformGemini, PlatformMCP, PlatformCopilot, PlatformGeneric:
		return true
	default:
		return false
	}
}

type CategoryType string

const (
	CategoryRules      CategoryType = "rules"
	CategorySkills     CategoryType = "skills"
	CategoryMCPServers CategoryType = "mcp_servers"
	CategoryPrompts    CategoryType = "prompts"
	CategoryFrameworks CategoryType = "frameworks"
	CategoryTools      CategoryType = "tools"
)

func (c CategoryType) IsValid() bool {
	switch c {
	case CategoryRules, CategorySkills, CategoryMCPServers, CategoryPrompts, CategoryFrameworks, CategoryTools:
		return true
	default:
		return false
	}
}

type Skill struct {
	ID             int          `json:"id"`
	Name           string       `json:"name"`
	Slug           string       `json:"slug"`
	Description    string       `json:"description"`
	Platform       PlatformType `json:"platform"`
	Category       CategoryType `json:"category"`
	Subcategory    string       `json:"subcategory"`
	SourceURL      string       `json:"source_url"`
	RepoURL        string       `json:"repo_url"`
	RepoOwner      string       `json:"repo_owner"`
	RepoName       string       `json:"repo_name"`
	StarsCount     int          `json:"stars_count"`
	ForksCount     int          `json:"forks_count"`
	StarsVelocity  int          `json:"stars_velocity"`
	ContentRaw     string       `json:"content_raw"`
	ContentPreview string       `json:"content_preview"`
	InstallSnippet string       `json:"install_snippet"`
	FilePath       string       `json:"file_path"`
	Language       string       `json:"language"`
	Tags           []string     `json:"tags"`
	LastSyncedAt   *time.Time   `json:"last_synced_at,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type Source struct {
	ID                  int        `json:"id"`
	Name                string     `json:"name"`
	URL                 string     `json:"url"`
	SourceType          string     `json:"source_type"`
	ScrapeStrategy      string     `json:"scrape_strategy"`
	IsActive            bool       `json:"is_active"`
	LastScrapedAt       *time.Time `json:"last_scraped_at,omitempty"`
	ScrapeIntervalHours int        `json:"scrape_interval_hours"`
	CreatedAt           time.Time  `json:"created_at"`
}

type User struct {
	ID        int       `json:"id"`
	GitHubID  string    `json:"github_id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Bookmark struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	SkillID   int       `json:"skill_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ScrapeLog struct {
	ID           int       `json:"id"`
	SourceID     *int      `json:"source_id,omitempty"`
	Status       string    `json:"status"`
	SkillsFound  int       `json:"skills_found"`
	SkillsNew    int       `json:"skills_new"`
	SkillsUpdated int      `json:"skills_updated"`
	ErrorMessage string    `json:"error_message,omitempty"`
	DurationMS   int       `json:"duration_ms"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type SkillDetail struct {
	Skill         Skill   `json:"skill"`
	IsBookmarked  bool    `json:"is_bookmarked"`
	RelatedSkills []Skill `json:"related_skills"`
}

type SkillFilter struct {
	Platform    string
	Category    string
	Subcategory string
	Language    string
	Tag         string
	Page        int
	Limit       int
	Sort        string
}

type SearchParams struct {
	Query       string `json:"query"`
	Platform    string `json:"platform"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Language    string `json:"language"`
	Tag         string `json:"tag"`
	Sort        string `json:"sort"`
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
}

type SearchResult struct {
	Skills     []Skill `json:"skills"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	TotalPages int     `json:"total_pages"`
	Query      string  `json:"query"`
}

type Stats struct {
	TotalSkills    int            `json:"total_skills"`
	TotalStars     int            `json:"total_stars"`
	TotalSources   int            `json:"total_sources"`
	PlatformCounts map[string]int `json:"platform_counts"`
	CategoryCounts map[string]int `json:"category_counts"`
}
