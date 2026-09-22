package service_test

import (
	"context"
	"testing"
	"time"

	"agentsight/internal/models"
	"agentsight/internal/service"
)

type mockSkillRepo struct {
	lastSearchParams models.SearchParams
	lastFilter       models.SkillFilter
	searchResults    []models.Skill
	searchTotal      int
	searchErr        error
}

func (m *mockSkillRepo) Create(ctx context.Context, skill *models.Skill) (*models.Skill, error) {
	return skill, nil
}
func (m *mockSkillRepo) GetByID(ctx context.Context, id int) (*models.Skill, error) {
	return nil, nil
}
func (m *mockSkillRepo) GetBySlug(ctx context.Context, slug string) (*models.Skill, error) {
	return nil, nil
}
func (m *mockSkillRepo) Update(ctx context.Context, skill *models.Skill) error {
	return nil
}
func (m *mockSkillRepo) Delete(ctx context.Context, id int) error {
	return nil
}
func (m *mockSkillRepo) List(ctx context.Context, filter models.SkillFilter) ([]models.Skill, int, error) {
	m.lastFilter = filter
	return m.searchResults, m.searchTotal, m.searchErr
}
func (m *mockSkillRepo) Search(ctx context.Context, params models.SearchParams) ([]models.Skill, int, error) {
	m.lastSearchParams = params
	return m.searchResults, m.searchTotal, m.searchErr
}
func (m *mockSkillRepo) GetTrending(ctx context.Context, limit int) ([]models.Skill, error) {
	return nil, nil
}
func (m *mockSkillRepo) ListByCategory(ctx context.Context, category string, limit, offset int) ([]models.Skill, int, error) {
	return nil, 0, nil
}
func (m *mockSkillRepo) ListByPlatform(ctx context.Context, platform string, limit, offset int) ([]models.Skill, int, error) {
	return nil, 0, nil
}
func (m *mockSkillRepo) GetRelated(ctx context.Context, skill *models.Skill, limit int) ([]models.Skill, error) {
	return nil, nil
}
func (m *mockSkillRepo) GetStats(ctx context.Context) (*models.Stats, error) {
	return nil, nil
}

func TestSearchService_PaginationDefaults(t *testing.T) {
	mockRepo := &mockSkillRepo{
		searchResults: []models.Skill{
			{ID: 1, Name: "Cursor Rule 1", Slug: "cursor-rule-1"},
		},
		searchTotal: 55,
	}

	svc := service.NewSearchService(mockRepo)
	ctx := context.Background()

	result, err := svc.Search(ctx, models.SearchParams{
		Query: "cursor",
		Page:  0,
		Limit: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockRepo.lastSearchParams.Page != 1 {
		t.Errorf("expected repo page=1, got %d", mockRepo.lastSearchParams.Page)
	}
	if mockRepo.lastSearchParams.Limit != 20 {
		t.Errorf("expected repo limit=20, got %d", mockRepo.lastSearchParams.Limit)
	}
	if result.Total != 55 {
		t.Errorf("expected total=55, got %d", result.Total)
	}
	if result.TotalPages != 3 {
		t.Errorf("expected total_pages=3 (55 / 20 ceiling), got %d", result.TotalPages)
	}
	if result.Page != 1 {
		t.Errorf("expected page=1, got %d", result.Page)
	}
	if result.Limit != 20 {
		t.Errorf("expected limit=20, got %d", result.Limit)
	}
}

func TestSearchService_PaginationEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		total         int
		inputLimit    int
		inputPage     int
		expectedPages int
		expectedLimit int
		expectedPage  int
	}{
		{
			name:          "zero total results",
			total:         0,
			inputLimit:    20,
			inputPage:     1,
			expectedPages: 0,
			expectedLimit: 20,
			expectedPage:  1,
		},
		{
			name:          "exact page boundary",
			total:         40,
			inputLimit:    20,
			inputPage:     1,
			expectedPages: 2,
			expectedLimit: 20,
			expectedPage:  1,
		},
		{
			name:          "limit exceeds 100 capped to 100",
			total:         150,
			inputLimit:    200,
			inputPage:     1,
			expectedPages: 2,
			expectedLimit: 100,
			expectedPage:  1,
		},
		{
			name:          "negative page normalized to 1",
			total:         10,
			inputLimit:    10,
			inputPage:     -3,
			expectedPages: 1,
			expectedLimit: 10,
			expectedPage:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockSkillRepo{
				searchTotal: tt.total,
			}
			svc := service.NewSearchService(mockRepo)
			res, err := svc.Search(context.Background(), models.SearchParams{
				Page:  tt.inputPage,
				Limit: tt.inputLimit,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.TotalPages != tt.expectedPages {
				t.Errorf("expected total pages %d, got %d", tt.expectedPages, res.TotalPages)
			}
			if mockRepo.lastSearchParams.Limit != tt.expectedLimit {
				t.Errorf("expected repo limit %d, got %d", tt.expectedLimit, mockRepo.lastSearchParams.Limit)
			}
			if mockRepo.lastSearchParams.Page != tt.expectedPage {
				t.Errorf("expected repo page %d, got %d", tt.expectedPage, mockRepo.lastSearchParams.Page)
			}
		})
	}
}

func TestSearchService_FilterCombinationsAndSort(t *testing.T) {
	mockRepo := &mockSkillRepo{
		searchTotal: 5,
		searchResults: []models.Skill{
			{ID: 1, Name: "React Best Practices", Platform: models.PlatformCursor, CreatedAt: time.Now()},
		},
	}
	svc := service.NewSearchService(mockRepo)

	// Test sanitization and sort defaults
	res, err := svc.Search(context.Background(), models.SearchParams{
		Query:       "  react agent  ",
		Platform:    "  CURSOR ",
		Category:    " RULES ",
		Subcategory: "Frontend",
		Language:    "TypeScript",
		Tag:         "Web",
		Sort:        "unknown_sort",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockRepo.lastSearchParams.Query != "react agent" {
		t.Errorf("expected query trimmed 'react agent', got '%s'", mockRepo.lastSearchParams.Query)
	}
	if mockRepo.lastSearchParams.Platform != "cursor" {
		t.Errorf("expected platform 'cursor', got '%s'", mockRepo.lastSearchParams.Platform)
	}
	if mockRepo.lastSearchParams.Category != "rules" {
		t.Errorf("expected category 'rules', got '%s'", mockRepo.lastSearchParams.Category)
	}
	if mockRepo.lastSearchParams.Sort != "relevance" {
		t.Errorf("expected default sort 'relevance' when query is present, got '%s'", mockRepo.lastSearchParams.Sort)
	}
	if len(res.Skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(res.Skills))
	}

	// Test sort default when query is empty
	_, _ = svc.Search(context.Background(), models.SearchParams{
		Query: "",
		Sort:  "invalid",
	})
	if mockRepo.lastSearchParams.Sort != "stars" {
		t.Errorf("expected default sort 'stars' when query empty, got '%s'", mockRepo.lastSearchParams.Sort)
	}

	// Test valid sort option preserved
	_, _ = svc.Search(context.Background(), models.SearchParams{
		Sort: "trending",
	})
	if mockRepo.lastSearchParams.Sort != "trending" {
		t.Errorf("expected sort 'trending' to be preserved, got '%s'", mockRepo.lastSearchParams.Sort)
	}
}

func TestSearchService_List(t *testing.T) {
	mockRepo := &mockSkillRepo{
		searchTotal: 2,
		searchResults: []models.Skill{
			{ID: 1, Name: "A"},
			{ID: 2, Name: "B"},
		},
	}
	svc := service.NewSearchService(mockRepo)

	skills, total, err := svc.List(context.Background(), models.SkillFilter{
		Page:  -1,
		Limit: 200,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(skills))
	}
	if mockRepo.lastFilter.Page != 1 {
		t.Errorf("expected page 1, got %d", mockRepo.lastFilter.Page)
	}
	if mockRepo.lastFilter.Limit != 100 {
		t.Errorf("expected limit capped to 100, got %d", mockRepo.lastFilter.Limit)
	}
}
