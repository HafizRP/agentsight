package service

import (
	"context"
	"strings"

	"agentsight/internal/models"
	"agentsight/internal/repository"
)

type SearchService interface {
	Search(ctx context.Context, params models.SearchParams) (*models.SearchResult, error)
	List(ctx context.Context, filter models.SkillFilter) ([]models.Skill, int, error)
}

type searchService struct {
	repo repository.SkillRepository
}

func NewSearchService(repo repository.SkillRepository) SearchService {
	return &searchService{repo: repo}
}

func (s *searchService) Search(ctx context.Context, params models.SearchParams) (*models.SearchResult, error) {
	params.Query = strings.TrimSpace(params.Query)
	params.Platform = strings.ToLower(strings.TrimSpace(params.Platform))
	params.Category = strings.ToLower(strings.TrimSpace(params.Category))
	params.Subcategory = strings.ToLower(strings.TrimSpace(params.Subcategory))
	params.Language = strings.TrimSpace(params.Language)
	params.Tag = strings.TrimSpace(params.Tag)

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	} else if params.Limit > 100 {
		params.Limit = 100
	}

	validSorts := map[string]bool{
		"relevance": true,
		"stars":     true,
		"newest":    true,
		"trending":  true,
	}
	if !validSorts[params.Sort] {
		if params.Query != "" {
			params.Sort = "relevance"
		} else {
			params.Sort = "stars"
		}
	}

	skills, total, err := s.repo.Search(ctx, params)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + params.Limit - 1) / params.Limit
	}

	return &models.SearchResult{
		Skills:     skills,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
		Query:      params.Query,
	}, nil
}

func (s *searchService) List(ctx context.Context, filter models.SkillFilter) ([]models.Skill, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.repo.List(ctx, filter)
}
