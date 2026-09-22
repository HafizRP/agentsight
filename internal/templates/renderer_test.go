package templates_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"agentsight/internal/models"
	"agentsight/internal/templates"
)

func TestHelperFunctions(t *testing.T) {
	// 1. Slugify
	if got := templates.Slugify("Cursor Rules for Python & Go!"); got != "cursor-rules-for-python-go" {
		t.Errorf("Slugify unexpected: %s", got)
	}

	// 2. Truncate
	if got := templates.Truncate("Hello World", 5); got != "Hello..." {
		t.Errorf("Truncate unexpected: %s", got)
	}
	if got := templates.Truncate("Short", 10); got != "Short" {
		t.Errorf("Truncate unexpected: %s", got)
	}

	// 3. TimeAgo
	now := time.Now()
	if got := templates.TimeAgo(now.Add(-2 * time.Minute)); got != "2m ago" {
		t.Errorf("TimeAgo unexpected: %s", got)
	}
	if got := templates.TimeAgo(now.Add(-3 * time.Hour)); got != "3h ago" {
		t.Errorf("TimeAgo unexpected: %s", got)
	}
	if got := templates.TimeAgo(now.Add(-5 * 24 * time.Hour)); got != "5d ago" {
		t.Errorf("TimeAgo unexpected: %s", got)
	}
	if got := templates.TimeAgo(nil); got != "never" {
		t.Errorf("TimeAgo nil unexpected: %s", got)
	}

	// 4. FormatNumber
	if got := templates.FormatNumber(500); got != "500" {
		t.Errorf("FormatNumber 500 unexpected: %s", got)
	}
	if got := templates.FormatNumber(1500); got != "1.5k" {
		t.Errorf("FormatNumber 1500 unexpected: %s", got)
	}
	if got := templates.FormatNumber(2000); got != "2k" {
		t.Errorf("FormatNumber 2000 unexpected: %s", got)
	}
	if got := templates.FormatNumber(1000000); got != "1M" {
		t.Errorf("FormatNumber 1M unexpected: %s", got)
	}

	// 5. PlatformColor
	platforms := []string{"cursor", "claude", "gemini", "mcp", "copilot", "generic"}
	for _, p := range platforms {
		col := templates.PlatformColor(p)
		if col == "" {
			t.Errorf("PlatformColor empty for %s", p)
		}
	}

	// 6. CategoryIcon
	categories := []string{"rules", "skills", "mcp_servers", "prompts", "frameworks", "tools", "other"}
	for _, c := range categories {
		icon := templates.CategoryIcon(c)
		if !strings.Contains(string(icon), "<svg") {
			t.Errorf("CategoryIcon missing svg for %s", c)
		}
	}

	// 7. Dict, Add, Sub, HasTag
	dict, err := templates.Dict("key1", "val1", "key2", 42)
	if err != nil || dict["key1"] != "val1" || dict["key2"] != 42 {
		t.Errorf("Dict unexpected: %v, %v", dict, err)
	}
	if templates.Add(2, 3) != 5 {
		t.Errorf("Add unexpected")
	}
	if templates.Sub(5, 2) != 3 {
		t.Errorf("Sub unexpected")
	}
	if !templates.HasTag([]string{"go", "HTMX"}, "htmx") {
		t.Errorf("HasTag unexpected false")
	}
}

func TestRenderer_AllTemplates(t *testing.T) {
	r, err := templates.NewRenderer()
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	mockSkill := models.Skill{
		ID:             1,
		Name:           "Test Skill",
		Slug:           "test-skill",
		Description:    "A test skill description",
		Platform:       models.PlatformCursor,
		Category:       models.CategoryRules,
		StarsCount:     150,
		StarsVelocity:  25,
		InstallSnippet: "npm i -g test-skill",
		ContentRaw:     "# Test Skill\nRule content goes here",
		CreatedAt:      time.Now(),
		Tags:           []string{"test", "rule"},
	}

	mockStats := &models.Stats{
		TotalSkills:    100,
		TotalStars:     25000,
		TotalSources:   15,
		PlatformCounts: map[string]int{"cursor": 50},
		CategoryCounts: map[string]int{"rules": 50},
	}

	pages := []struct {
		name string
		data interface{}
	}{
		{
			name: "index.html",
			data: map[string]interface{}{
				"Title":  "Home",
				"Skills": []models.Skill{mockSkill},
				"Stats":  mockStats,
			},
		},
		{
			name: "search.html",
			data: map[string]interface{}{
				"Title": "Search",
				"Params": models.SearchParams{
					Query:    "test",
					Platform: "cursor",
				},
				"Result": &models.SearchResult{
					Skills:     []models.Skill{mockSkill},
					Total:      1,
					Page:       1,
					Limit:      20,
					TotalPages: 1,
				},
			},
		},
		{
			name: "skill_detail.html",
			data: map[string]interface{}{
				"Title":         "Test Skill",
				"Skill":         mockSkill,
				"IsBookmarked":  true,
				"RelatedSkills": []models.Skill{mockSkill},
			},
		},
		{
			name: "trending.html",
			data: map[string]interface{}{
				"Title":     "Trending",
				"Skills":    []models.Skill{mockSkill},
				"Timeframe": "week",
			},
		},
		{
			name: "category.html",
			data: map[string]interface{}{
				"Title":      "Rules",
				"Category":   "rules",
				"Skills":     []models.Skill{mockSkill},
				"Total":      1,
				"Page":       1,
				"TotalPages": 1,
			},
		},
		{
			name: "platform.html",
			data: map[string]interface{}{
				"Title":      "Cursor",
				"Platform":   "cursor",
				"Skills":     []models.Skill{mockSkill},
				"Total":      1,
				"Page":       1,
				"TotalPages": 1,
			},
		},
		{
			name: "auth/login.html",
			data: map[string]interface{}{
				"Title": "Sign In",
			},
		},
	}

	for _, tc := range pages {
		t.Run("Page/"+tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := r.Render(&buf, tc.name, tc.data); err != nil {
				t.Fatalf("Render %s failed: %v", tc.name, err)
			}
			out := buf.String()
			if !strings.Contains(out, "<!DOCTYPE html>") {
				t.Errorf("%s missing doctype", tc.name)
			}
			if !strings.Contains(out, "AgentSight") {
				t.Errorf("%s missing AgentSight brand", tc.name)
			}
		})
	}

	partials := []struct {
		name string
		data interface{}
	}{
		{
			name: "partials/_skill_card.html",
			data: mockSkill,
		},
		{
			name: "partials/_skill_list.html",
			data: map[string]interface{}{
				"Skills":     []models.Skill{mockSkill},
				"Total":      1,
				"Page":       1,
				"TotalPages": 1,
			},
		},
		{
			name: "partials/_search_results.html",
			data: map[string]interface{}{
				"Skills":     []models.Skill{mockSkill},
				"Total":      1,
				"Page":       1,
				"TotalPages": 1,
				"Query":      "test",
			},
		},
		{
			name: "partials/_bookmark_button.html",
			data: map[string]interface{}{
				"SkillID":    1,
				"Bookmarked": true,
			},
		},
	}

	for _, tc := range partials {
		t.Run("Partial/"+tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := r.RenderPartial(&buf, tc.name, tc.data); err != nil {
				t.Fatalf("RenderPartial %s failed: %v", tc.name, err)
			}
			if buf.Len() == 0 {
				t.Errorf("RenderPartial %s returned empty output", tc.name)
			}
		})
	}
}
