package scraper

import (
	"context"
	"time"
)

// RawSkill represents an unprocessed skill discovered by a scraper adapter.
type RawSkill struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	SourceURL      string     `json:"source_url"`
	RepoURL        string     `json:"repo_url"`
	RepoOwner      string     `json:"repo_owner"`
	RepoName       string     `json:"repo_name"`
	StarsCount     int        `json:"stars_count"`
	ForksCount     int        `json:"forks_count"`
	StarsVelocity  int        `json:"stars_velocity"`
	ContentRaw     string     `json:"content_raw"`
	ContentPreview string     `json:"content_preview"`
	InstallSnippet string     `json:"install_snippet"`
	FilePath       string     `json:"file_path"`
	Language       string     `json:"language"`
	Platform       string     `json:"platform"`
	Category       string     `json:"category"`
	Subcategory    string     `json:"subcategory"`
	Tags           []string   `json:"tags"`
	LastSyncedAt   *time.Time `json:"last_synced_at,omitempty"`
}

// Scraper defines the interface for all source discovery adapters.
type Scraper interface {
	Discover(ctx context.Context) ([]RawSkill, error)
	Name() string
}
