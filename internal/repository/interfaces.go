package repository

import (
	"context"
	"time"

	"agentsight/internal/models"
)

type SkillRepository interface {
	Create(ctx context.Context, skill *models.Skill) (*models.Skill, error)
	GetByID(ctx context.Context, id int) (*models.Skill, error)
	GetBySlug(ctx context.Context, slug string) (*models.Skill, error)
	Update(ctx context.Context, skill *models.Skill) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter models.SkillFilter) ([]models.Skill, int, error)
	Search(ctx context.Context, params models.SearchParams) ([]models.Skill, int, error)
	GetTrending(ctx context.Context, limit int) ([]models.Skill, error)
	ListByCategory(ctx context.Context, category string, limit, offset int) ([]models.Skill, int, error)
	ListByPlatform(ctx context.Context, platform string, limit, offset int) ([]models.Skill, int, error)
	GetRelated(ctx context.Context, skill *models.Skill, limit int) ([]models.Skill, error)
	GetStats(ctx context.Context) (*models.Stats, error)
}

type SourceRepository interface {
	Create(ctx context.Context, source *models.Source) (*models.Source, error)
	GetByID(ctx context.Context, id int) (*models.Source, error)
	List(ctx context.Context) ([]models.Source, error)
	ListActive(ctx context.Context) ([]models.Source, error)
	Update(ctx context.Context, source *models.Source) error
	UpdateLastScraped(ctx context.Context, id int, t time.Time) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByGitHubID(ctx context.Context, githubID string) (*models.User, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
	UpsertByGitHub(ctx context.Context, user *models.User) (*models.User, error)
}

type BookmarkRepository interface {
	Toggle(ctx context.Context, userID, skillID int) (bool, error)
	ListByUser(ctx context.Context, userID int, limit, offset int) ([]models.Skill, int, error)
	IsBookmarked(ctx context.Context, userID, skillID int) (bool, error)
	GetBookmarkedSkillIDs(ctx context.Context, userID int) (map[int]bool, error)
}

type ScrapeLogRepository interface {
	Create(ctx context.Context, log *models.ScrapeLog) error
	ListRecent(ctx context.Context, limit int) ([]models.ScrapeLog, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSession(ctx context.Context, token string) (*models.Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteExpiredSessions(ctx context.Context) error
}

type Repositories struct {
	Skills    SkillRepository
	Sources   SourceRepository
	Users     UserRepository
	Bookmarks BookmarkRepository
	ScrapeLogs ScrapeLogRepository
	Sessions  SessionRepository
}
