package scraper_test

import (
	"testing"
	"time"

	"agentsight/internal/models"
	"agentsight/internal/scraper"
)

func TestNormalizer_GenerateSlug(t *testing.T) {
	n := scraper.NewNormalizer()

	tests := []struct {
		input    string
		expected string
	}{
		{"Next.js 14 App Router & Server Actions", "next-js-14-app-router-server-actions"},
		{"FastAPI & Python 3.12 (Async)", "fastapi-python-3-12-async"},
		{"  Go Chi / REST API Clean Architecture  ", "go-chi-rest-api-clean-architecture"},
		{"---Multiple---Hyphens---", "multiple-hyphens"},
		{"!@#$%^&*()_+", "skill-untitled"},
		{"", "skill-untitled"},
		{"Vue 3 + Pinia", "vue-3-pinia"},
	}

	for _, tt := range tests {
		actual := n.GenerateSlug(tt.input)
		if actual != tt.expected {
			t.Errorf("GenerateSlug(%q) = %q, expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestNormalizer_DetectPlatform(t *testing.T) {
	n := scraper.NewNormalizer()

	tests := []struct {
		filePath  string
		sourceURL string
		repoURL   string
		name      string
		expected  models.PlatformType
	}{
		{".cursorrules", "", "", "Cursor Rules", models.PlatformCursor},
		{".cursor/rules/nextjs.mdc", "", "", "NextJS MDC", models.PlatformCursor},
		{"rules/react.mdc", "https://cursor.directory/react", "", "React Rule", models.PlatformCursor},
		{"AGENTS.md", "https://github.com/agentsmd/agents.md", "", "AGENTS Spec", models.PlatformClaude},
		{"CLAUDE.md", "", "https://github.com/anthropics/claude-code", "Claude Code", models.PlatformClaude},
		{".gemini/skills/SKILL.md", "", "", "Gemini Skill", models.PlatformGemini},
		{"mcp_config.json", "", "", "MCP Config", models.PlatformMCP},
		{"README.md", "https://mcpservers.org/postgres", "", "Postgres Server", models.PlatformMCP},
		{"README.md", "https://github.com/modelcontextprotocol/servers", "", "MCP Server", models.PlatformMCP},
		{".github/copilot-instructions.md", "", "", "Copilot TS", models.PlatformCopilot},
		{"main.go", "https://github.com/example/repo", "", "Unknown Project", models.PlatformGeneric},
	}

	for _, tt := range tests {
		actual := n.DetectPlatform(tt.filePath, tt.sourceURL, tt.repoURL, tt.name)
		if actual != tt.expected {
			t.Errorf("DetectPlatform(%q, %q, %q, %q) = %v, expected %v",
				tt.filePath, tt.sourceURL, tt.repoURL, tt.name, actual, tt.expected)
		}
	}
}

func TestNormalizer_DetectCategory(t *testing.T) {
	n := scraper.NewNormalizer()

	tests := []struct {
		platform models.PlatformType
		name     string
		desc     string
		content  string
		expected models.CategoryType
	}{
		{models.PlatformMCP, "Postgres Server", "PostgreSQL database context provider", "", models.CategoryMCPServers},
		{models.PlatformCursor, "React Guidelines", "Coding rules and standards", "", models.CategoryRules},
		{models.PlatformGeneric, "LangGraph Agent", "StateGraph workflow orchestrator", "stategraph conditional edges", models.CategoryFrameworks},
		{models.PlatformGeneric, "STRIDE Model", "Threat model and prompt template", "threat model evaluation persona", models.CategoryPrompts},
		{models.PlatformGeneric, "OpenAPI Tool Generator", "CLI generator for schemas", "tool cli generator for json schema", models.CategoryTools},
		{models.PlatformClaude, "Fullstack TDD Workflow", "Autonomous development practices", "implement tdd cycle", models.CategorySkills},
	}

	for _, tt := range tests {
		actual := n.DetectCategory(tt.platform, tt.name, tt.desc, tt.content)
		if actual != tt.expected {
			t.Errorf("DetectCategory(%v, %q, ...) = %v, expected %v", tt.platform, tt.name, actual, tt.expected)
		}
	}
}

func TestNormalizer_DetectLanguage(t *testing.T) {
	n := scraper.NewNormalizer()

	tests := []struct {
		filePath string
		content  string
		expected string
	}{
		{"handler.go", "package main", "go"},
		{"index.ts", "export const handler = () => {}", "typescript"},
		{"app.py", "def main():\n    pass", "python"},
		{"main.rs", "fn main() {}", "rust"},
		{"schema.sql", "SELECT id, name FROM users;", "sql"},
		{"Dockerfile", "FROM golang:1.22-alpine", "dockerfile"},
		{"README.md", "func NewServer() *Server", "go"},
		{"README.md", "def run_agent():", "python"},
		{"README.md", "interface AgentContext { id: string }", "typescript"},
		{"README.md", "Plain text instructions", "markdown"},
	}

	for _, tt := range tests {
		actual := n.DetectLanguage(tt.filePath, tt.content)
		if actual != tt.expected {
			t.Errorf("DetectLanguage(%q, %q) = %q, expected %q", tt.filePath, tt.content, actual, tt.expected)
		}
	}
}

func TestNormalizer_ExtractInstallSnippet(t *testing.T) {
	n := scraper.NewNormalizer()

	contentWithCode := "## Installation\n```bash\nnpx -y @modelcontextprotocol/server-postgres\n```"
	snippet := n.ExtractInstallSnippet(models.PlatformMCP, "postgres", "", "server-postgres", contentWithCode)
	if snippet != "npx -y @modelcontextprotocol/server-postgres" {
		t.Errorf("expected extracted code block command, got: %s", snippet)
	}

	// Fallback Cursor
	cursorFallback := n.ExtractInstallSnippet(models.PlatformCursor, "react", "https://example.com/react.mdc", "", "")
	if cursorFallback != "curl -o .cursor/rules/react.mdc https://example.com/react.mdc" {
		t.Errorf("expected cursor fallback curl command, got: %s", cursorFallback)
	}

	// Fallback Claude
	claudeFallback := n.ExtractInstallSnippet(models.PlatformClaude, "agents", "https://example.com/AGENTS.md", "", "")
	if claudeFallback != "curl -o AGENTS.md https://example.com/AGENTS.md" {
		t.Errorf("expected claude fallback curl command, got: %s", claudeFallback)
	}
}

func TestNormalizer_ExtractTags(t *testing.T) {
	n := scraper.NewNormalizer()

	tags := n.ExtractTags(models.PlatformCursor, "typescript", "Next.js App Router", "Fullstack react and tailwind guidelines", "Includes docker and postgres setups with tdd testing")

	tagMap := make(map[string]bool)
	for _, tag := range tags {
		tagMap[tag] = true
	}

	expectedKeywords := []string{"cursor", "typescript", "nextjs", "react", "tailwind", "docker", "postgres", "tdd", "testing"}
	for _, kw := range expectedKeywords {
		if !tagMap[kw] {
			t.Errorf("expected tag %q in extracted tags: %v", kw, tags)
		}
	}
}

func TestNormalizer_Normalize_Success(t *testing.T) {
	n := scraper.NewNormalizer()
	now := time.Now()

	raw := scraper.RawSkill{
		Name:         "Awesome Next.js Cursor Rule",
		Description:  "Production Next.js 14 conventions with TypeScript",
		SourceURL:    "https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/nextjs.mdc",
		RepoURL:      "https://github.com/PatrickJS/awesome-cursorrules",
		RepoOwner:    "PatrickJS",
		RepoName:     "awesome-cursorrules",
		StarsCount:   3200,
		ForksCount:   240,
		FilePath:     "rules/nextjs.mdc",
		ContentRaw:   "# Next.js Rules\nAlways use App Router and Server Components.",
		LastSyncedAt: &now,
	}

	skill, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("Normalize failed unexpectedly: %v", err)
	}

	if skill.Name != raw.Name {
		t.Errorf("expected name %q, got %q", raw.Name, skill.Name)
	}
	if skill.Slug != "awesome-next-js-cursor-rule" {
		t.Errorf("expected slug 'awesome-next-js-cursor-rule', got %q", skill.Slug)
	}
	if skill.Platform != models.PlatformCursor {
		t.Errorf("expected platform cursor, got %v", skill.Platform)
	}
	if skill.Category != models.CategoryRules {
		t.Errorf("expected category rules, got %v", skill.Category)
	}
	if skill.Subcategory != "backend" {
		t.Errorf("expected subcategory backend, got %q", skill.Subcategory)
	}
	if skill.Language != "typescript" {
		t.Errorf("expected language typescript, got %q", skill.Language)
	}
	if skill.StarsCount != 3200 {
		t.Errorf("expected stars 3200, got %d", skill.StarsCount)
	}
	if skill.InstallSnippet == "" {
		t.Errorf("expected non-empty install snippet")
	}
}

func TestNormalizer_Normalize_ValidationFailure(t *testing.T) {
	n := scraper.NewNormalizer()

	raw := scraper.RawSkill{
		Name:     "",
		RepoName: "",
	}

	_, err := n.Normalize(raw)
	if err == nil {
		t.Errorf("expected error when normalizing skill with empty name and repo, got nil")
	}
}
