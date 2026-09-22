package service

import (
	"context"

	"agentsight/internal/models"
	"agentsight/internal/repository"
)

type TrendingService interface {
	GetTrending(ctx context.Context, limit int) ([]models.Skill, error)
	CalculateStarsVelocity(starsCurrent, starsPrevious int, days int) int
}

type trendingService struct {
	skillRepo repository.SkillRepository
}

func NewTrendingService(skillRepo repository.SkillRepository) TrendingService {
	return &trendingService{skillRepo: skillRepo}
}

func (s *trendingService) GetTrending(ctx context.Context, limit int) ([]models.Skill, error) {
	if limit <= 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}
	return s.skillRepo.GetTrending(ctx, limit)
}

func (s *trendingService) CalculateStarsVelocity(starsCurrent, starsPrevious int, days int) int {
	if days <= 0 {
		days = 7
	}
	diff := starsCurrent - starsPrevious
	if diff < 0 {
		diff = 0
	}
	return (diff * 7) / days
}
