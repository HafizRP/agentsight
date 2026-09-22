package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"agentsight/internal/models"
	"agentsight/internal/repository"
)

var (
	nonAlphaNumericRegex = regexp.MustCompile(`[^a-z0-9]+`)
	multipleDashRegex    = regexp.MustCompile(`-+`)
)

type SkillService interface {
	GetBySlug(ctx context.Context, slug string, currentUserID int) (*models.SkillDetail, error)
	GetByID(ctx context.Context, id int) (*models.Skill, error)
	List(ctx context.Context, filter models.SkillFilter) ([]models.Skill, int, error)
	ListByCategory(ctx context.Context, category string, page, limit int) ([]models.Skill, int, error)
	ListByPlatform(ctx context.Context, platform string, page, limit int) ([]models.Skill, int, error)
	GetRelated(ctx context.Context, skill *models.Skill, limit int) ([]models.Skill, error)
	GetStats(ctx context.Context) (*models.Stats, error)
	GenerateSlug(name string) string
}

type skillService struct {
	skillRepo    repository.SkillRepository
	bookmarkRepo repository.BookmarkRepository
}

func NewSkillService(skillRepo repository.SkillRepository, bookmarkRepo repository.BookmarkRepository) SkillService {
	return &skillService{
		skillRepo:    skillRepo,
		bookmarkRepo: bookmarkRepo,
	}
}

func (s *skillService) GetBySlug(ctx context.Context, slug string, currentUserID int) (*models.SkillDetail, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, errors.New("empty slug")
	}

	skill, err := s.skillRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	isBookmarked := false
	if currentUserID > 0 && s.bookmarkRepo != nil {
		isBookmarked, _ = s.bookmarkRepo.IsBookmarked(ctx, currentUserID, skill.ID)
	}

	related, err := s.skillRepo.GetRelated(ctx, skill, 4)
	if err != nil {
		related = []models.Skill{}
	}

	return &models.SkillDetail{
		Skill:         *skill,
		IsBookmarked:  isBookmarked,
		RelatedSkills: related,
	}, nil
}

func (s *skillService) GetByID(ctx context.Context, id int) (*models.Skill, error) {
	return s.skillRepo.GetByID(ctx, id)
}

func (s *skillService) List(ctx context.Context, filter models.SkillFilter) ([]models.Skill, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.skillRepo.List(ctx, filter)
}

func (s *skillService) ListByCategory(ctx context.Context, category string, page, limit int) ([]models.Skill, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	return s.skillRepo.ListByCategory(ctx, category, limit, offset)
}

func (s *skillService) ListByPlatform(ctx context.Context, platform string, page, limit int) ([]models.Skill, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	return s.skillRepo.ListByPlatform(ctx, platform, limit, offset)
}

func (s *skillService) GetRelated(ctx context.Context, skill *models.Skill, limit int) ([]models.Skill, error) {
	if limit <= 0 {
		limit = 4
	}
	return s.skillRepo.GetRelated(ctx, skill, limit)
}

func (s *skillService) GetStats(ctx context.Context) (*models.Stats, error) {
	return s.skillRepo.GetStats(ctx)
}

func (s *skillService) GenerateSlug(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	slug := nonAlphaNumericRegex.ReplaceAllString(lower, "-")
	slug = multipleDashRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "skill-untitled"
	}
	return slug
}
