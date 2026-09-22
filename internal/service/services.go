package service

import (
	"agentsight/internal/config"
	"agentsight/internal/repository"
)

type Services struct {
	Search   SearchService
	Skill    SkillService
	Auth     AuthService
	Trending TrendingService
}

func NewServices(repos *repository.Repositories, cfg *config.Config) *Services {
	return &Services{
		Search:   NewSearchService(repos.Skills),
		Skill:    NewSkillService(repos.Skills, repos.Bookmarks),
		Auth:     NewAuthService(cfg, repos.Users, repos.Sessions),
		Trending: NewTrendingService(repos.Skills),
	}
}
