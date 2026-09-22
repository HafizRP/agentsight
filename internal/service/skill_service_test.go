package service_test

import (
	"context"
	"errors"
	"testing"

	"agentsight/internal/models"
	"agentsight/internal/service"
)

type mockBookmarkRepo struct {
	bookmarkedMap map[int]bool
}

func (m *mockBookmarkRepo) Toggle(ctx context.Context, userID, skillID int) (bool, error) {
	current := m.bookmarkedMap[skillID]
	m.bookmarkedMap[skillID] = !current
	return !current, nil
}

func (m *mockBookmarkRepo) ListByUser(ctx context.Context, userID int, limit, offset int) ([]models.Skill, int, error) {
	return nil, 0, nil
}

func (m *mockBookmarkRepo) IsBookmarked(ctx context.Context, userID, skillID int) (bool, error) {
	return m.bookmarkedMap[skillID], nil
}

func (m *mockBookmarkRepo) GetBookmarkedSkillIDs(ctx context.Context, userID int) (map[int]bool, error) {
	return m.bookmarkedMap, nil
}

type fullMockSkillRepo struct {
	mockSkillRepo
	skillsBySlug map[string]*models.Skill
	skillsByID   map[int]*models.Skill
	related      []models.Skill
}

func (m *fullMockSkillRepo) GetBySlug(ctx context.Context, slug string) (*models.Skill, error) {
	s, ok := m.skillsBySlug[slug]
	if !ok {
		return nil, errors.New("skill not found")
	}
	return s, nil
}

func (m *fullMockSkillRepo) GetByID(ctx context.Context, id int) (*models.Skill, error) {
	s, ok := m.skillsByID[id]
	if !ok {
		return nil, errors.New("skill not found")
	}
	return s, nil
}

func (m *fullMockSkillRepo) GetRelated(ctx context.Context, skill *models.Skill, limit int) ([]models.Skill, error) {
	return m.related, nil
}

func TestSkillService_GenerateSlug(t *testing.T) {
	svc := service.NewSkillService(nil, nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple title",
			input:    "React Best Practices",
			expected: "react-best-practices",
		},
		{
			name:     "symbols and versions",
			input:    "MCP Server: File System (v2.0)!",
			expected: "mcp-server-file-system-v2-0",
		},
		{
			name:     "spaces and underscores",
			input:    "   claude_code   guidelines   ",
			expected: "claude-code-guidelines",
		},
		{
			name:     "special characters only",
			input:    "@#$%^&*()+={}[]",
			expected: "skill-untitled",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "skill-untitled",
		},
		{
			name:     "multiple consecutive hyphens",
			input:    "Go----Concurrency---Patterns",
			expected: "go-concurrency-patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.GenerateSlug(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSkillService_GetBySlug(t *testing.T) {
	skill := &models.Skill{
		ID:       10,
		Name:     "Docker Compose Expert",
		Slug:     "docker-compose-expert",
		Platform: models.PlatformClaude,
		Category: models.CategorySkills,
	}

	related := []models.Skill{
		{ID: 11, Name: "Kubernetes Helper", Slug: "kubernetes-helper"},
	}

	repo := &fullMockSkillRepo{
		skillsBySlug: map[string]*models.Skill{
			"docker-compose-expert": skill,
		},
		related: related,
	}

	bookmarkRepo := &mockBookmarkRepo{
		bookmarkedMap: map[int]bool{
			10: true,
		},
	}

	svc := service.NewSkillService(repo, bookmarkRepo)
	ctx := context.Background()

	// 1. Unauthenticated user
	detailUnauth, err := svc.GetBySlug(ctx, "docker-compose-expert", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detailUnauth.IsBookmarked {
		t.Errorf("expected is_bookmarked=false for unauthenticated user")
	}
	if len(detailUnauth.RelatedSkills) != 1 {
		t.Errorf("expected 1 related skill, got %d", len(detailUnauth.RelatedSkills))
	}

	// 2. Authenticated user who bookmarked
	detailAuth, err := svc.GetBySlug(ctx, "docker-compose-expert", 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !detailAuth.IsBookmarked {
		t.Errorf("expected is_bookmarked=true for bookmarked user")
	}

	// 3. Empty slug error
	_, err = svc.GetBySlug(ctx, "   ", 0)
	if err == nil {
		t.Errorf("expected error on empty slug, got nil")
	}

	// 4. Non-existent slug error
	_, err = svc.GetBySlug(ctx, "non-existent-skill", 0)
	if err == nil {
		t.Errorf("expected error on not found skill, got nil")
	}
}
