package scraper

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"agentsight/internal/models"
)

var (
	nonAlphaNumericRegex = regexp.MustCompile(`[^a-z0-9]+`)
	multipleDashRegex    = regexp.MustCompile(`-+`)
	codeBlockRegex       = regexp.MustCompile("(?s)```(?:bash|sh|zsh)?\\s*\\n([^`]+)```")
)

// Normalizer transforms raw scraped data into validated and standardized models.Skill entities.
type Normalizer struct{}

// NewNormalizer creates a new Normalizer instance.
func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

// GenerateSlug converts any name into a normalized, URL-safe kebab-case slug.
func (n *Normalizer) GenerateSlug(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	slug := nonAlphaNumericRegex.ReplaceAllString(lower, "-")
	slug = multipleDashRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "skill-untitled"
	}
	return slug
}

// DetectPlatform determines the target platform based on file path, source URL, or metadata.
func (n *Normalizer) DetectPlatform(filePath, sourceURL, repoURL, name string) models.PlatformType {
	target := strings.ToLower(fmt.Sprintf("%s %s %s %s", filePath, sourceURL, repoURL, name))

	switch {
	case strings.Contains(target, ".cursorrules") ||
		strings.Contains(target, ".cursor/rules") ||
		strings.Contains(target, "cursor.directory") ||
		strings.Contains(target, "awesome-cursorrules") ||
		strings.HasSuffix(filePath, ".mdc"):
		return models.PlatformCursor

	case strings.Contains(target, "agents.md") ||
		strings.Contains(target, "claude.md") ||
		strings.Contains(target, ".claude") ||
		strings.Contains(target, "claude-code"):
		return models.PlatformClaude

	case strings.Contains(target, ".gemini") ||
		strings.Contains(target, "gemini-skills") ||
		strings.Contains(target, "google-gemini"):
		return models.PlatformGemini

	case strings.Contains(target, "mcp_config.json") ||
		strings.Contains(target, "mcpservers.org") ||
		strings.Contains(target, "awesome-mcp-servers") ||
		strings.Contains(target, "modelcontextprotocol"):
		return models.PlatformMCP

	case strings.Contains(target, "copilot-instructions.md") ||
		strings.Contains(target, "copilot"):
		return models.PlatformCopilot

	default:
		return models.PlatformGeneric
	}
}

// DetectCategory identifies the skill category based on platform, name, and content.
func (n *Normalizer) DetectCategory(platform models.PlatformType, name, description, content string) models.CategoryType {
	text := strings.ToLower(fmt.Sprintf("%s %s %s", name, description, content))

	if platform == models.PlatformMCP || strings.Contains(text, "mcp server") || strings.Contains(text, "model context protocol") {
		return models.CategoryMCPServers
	}

	if platform == models.PlatformCursor || platform == models.PlatformCopilot || strings.Contains(text, "rule") || strings.Contains(text, "standards") {
		return models.CategoryRules
	}

	if strings.Contains(text, "stategraph") || strings.Contains(text, "langgraph") ||
		strings.Contains(text, "crewai") || strings.Contains(text, "autogen") ||
		strings.Contains(text, "semantic kernel") || strings.Contains(text, "llamaindex") ||
		strings.Contains(text, "framework") || strings.Contains(text, "react loop") {
		return models.CategoryFrameworks
	}

	if strings.Contains(text, "prompt") || strings.Contains(text, "persona") || strings.Contains(text, "threat model") {
		return models.CategoryPrompts
	}

	if strings.Contains(text, "tool") || strings.Contains(text, "generator") || strings.Contains(text, "cli") || strings.Contains(text, "extension") {
		return models.CategoryTools
	}

	return models.CategorySkills
}

// DetectLanguage identifies the primary programming language or format.
func (n *Normalizer) DetectLanguage(filePath, content string) string {
	lowerPath := strings.ToLower(filePath)
	switch {
	case strings.HasSuffix(lowerPath, ".ts") || strings.HasSuffix(lowerPath, ".tsx"):
		return "typescript"
	case strings.HasSuffix(lowerPath, ".js") || strings.HasSuffix(lowerPath, ".jsx"):
		return "javascript"
	case strings.HasSuffix(lowerPath, ".py"):
		return "python"
	case strings.HasSuffix(lowerPath, ".go"):
		return "go"
	case strings.HasSuffix(lowerPath, ".rs"):
		return "rust"
	case strings.HasSuffix(lowerPath, ".java"):
		return "java"
	case strings.HasSuffix(lowerPath, ".kt") || strings.HasSuffix(lowerPath, ".kts"):
		return "kotlin"
	case strings.HasSuffix(lowerPath, ".swift"):
		return "swift"
	case strings.HasSuffix(lowerPath, ".rb"):
		return "ruby"
	case strings.HasSuffix(lowerPath, ".ex") || strings.HasSuffix(lowerPath, ".exs"):
		return "elixir"
	case strings.HasSuffix(lowerPath, ".cs"):
		return "csharp"
	case strings.HasSuffix(lowerPath, ".cpp") || strings.HasSuffix(lowerPath, ".cc") || strings.HasSuffix(lowerPath, ".cxx"):
		return "cpp"
	case strings.HasSuffix(lowerPath, ".sql"):
		return "sql"
	case strings.HasSuffix(lowerPath, ".yaml") || strings.HasSuffix(lowerPath, ".yml"):
		return "yaml"
	case strings.HasSuffix(lowerPath, "dockerfile") || strings.Contains(lowerPath, "docker"):
		return "dockerfile"
	}

	lowerContent := strings.ToLower(content)
	lowerAll := strings.ToLower(fmt.Sprintf("%s %s", filePath, content))
	switch {
	case strings.Contains(lowerAll, "typescript") || strings.Contains(lowerAll, "interface ") ||
		strings.Contains(lowerAll, ": string") || strings.Contains(lowerAll, "next.js") ||
		strings.Contains(lowerAll, "nextjs") || strings.Contains(lowerAll, "from 'react'") ||
		strings.Contains(lowerAll, "export default function"):
		return "typescript"
	case strings.Contains(lowerAll, "python") || strings.Contains(lowerAll, "def ") || strings.Contains(lowerAll, "import pydantic") || strings.Contains(lowerAll, "fastapi"):
		return "python"
	case strings.Contains(lowerAll, "golang") || strings.Contains(lowerAll, "func ") || strings.Contains(lowerAll, "package ") || strings.Contains(lowerAll, "go-chi"):
		return "go"
	case strings.Contains(lowerAll, "rust") || strings.Contains(lowerAll, "fn ") || strings.Contains(lowerAll, "impl ") || strings.Contains(lowerAll, "axum"):
		return "rust"
	case strings.Contains(lowerContent, "dockerfile") || strings.Contains(lowerContent, "from alpine"):
		return "dockerfile"
	case strings.Contains(lowerContent, "select ") && strings.Contains(lowerContent, "from "):
		return "sql"
	default:
		return "markdown"
	}
}

// ExtractInstallSnippet discovers install/usage commands or generates an idiomatic fallback.
func (n *Normalizer) ExtractInstallSnippet(platform models.PlatformType, slug, sourceURL, repoName, content string) string {
	matches := codeBlockRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			lines := strings.Split(strings.TrimSpace(match[1]), "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "npx ") ||
					strings.HasPrefix(trimmed, "npm install") ||
					strings.HasPrefix(trimmed, "pip install") ||
					strings.HasPrefix(trimmed, "curl ") ||
					strings.HasPrefix(trimmed, "cargo add") ||
					strings.HasPrefix(trimmed, "go get") {
					return trimmed
				}
			}
		}
	}

	// Synthesize fallback install snippet based on platform
	switch platform {
	case models.PlatformCursor:
		if sourceURL != "" {
			return fmt.Sprintf("curl -o .cursor/rules/%s.mdc %s", slug, sourceURL)
		}
		return fmt.Sprintf("curl -o .cursor/rules/%s.mdc https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/%s.mdc", slug, slug)

	case models.PlatformClaude:
		if sourceURL != "" {
			return fmt.Sprintf("curl -o AGENTS.md %s", sourceURL)
		}
		return "curl -o AGENTS.md https://raw.githubusercontent.com/agentsmd/agents.md/main/examples/fullstack-tdd.md"

	case models.PlatformGemini:
		if sourceURL != "" {
			return fmt.Sprintf("mkdir -p .gemini/skills && curl -o .gemini/skills/%s.md %s", slug, sourceURL)
		}
		return fmt.Sprintf("mkdir -p .gemini/skills && curl -o .gemini/skills/%s.md https://example.com/%s.md", slug, slug)

	case models.PlatformCopilot:
		if sourceURL != "" {
			return fmt.Sprintf("curl -o .github/copilot-instructions.md %s", sourceURL)
		}
		return "curl -o .github/copilot-instructions.md https://raw.githubusercontent.com/github/copilot-instructions/main/typescript.md"

	case models.PlatformMCP:
		if repoName != "" {
			return fmt.Sprintf("npx -y @modelcontextprotocol/server-%s", strings.TrimPrefix(repoName, "server-"))
		}
		return fmt.Sprintf("npx -y @modelcontextprotocol/server-%s", slug)

	default:
		if sourceURL != "" {
			return fmt.Sprintf("curl -o skills/%s.md %s", slug, sourceURL)
		}
		return fmt.Sprintf("# Install %s", slug)
	}
}

// ExtractTags analyzes content and metadata to generate relevant searchable tags.
func (n *Normalizer) ExtractTags(platform models.PlatformType, language string, name, description, content string) []string {
	tagSet := make(map[string]struct{})

	// Add platform and language
	if platform != "" && platform != models.PlatformGeneric {
		tagSet[string(platform)] = struct{}{}
	}
	if language != "" && language != "markdown" {
		tagSet[language] = struct{}{}
	}

	searchSpace := strings.ToLower(fmt.Sprintf("%s %s %s", name, description, content))
	normalizedSpace := strings.ReplaceAll(searchSpace, ".", "")

	knownKeywords := []string{
		"nextjs", "react", "vue", "svelte", "astro", "angular", "tailwind",
		"fastapi", "django", "flask", "express", "nestjs", "spring-boot",
		"rails", "phoenix", "axum", "actix", "go-chi", "gin",
		"docker", "kubernetes", "k8s", "helm", "terraform", "ansible",
		"postgres", "postgresql", "sqlite", "redis", "elasticsearch", "clickhouse",
		"graphql", "grpc", "protobuf", "rest", "openapi",
		"tdd", "testing", "playwright", "cypress", "security", "owasp",
		"langchain", "langgraph", "autogen", "crewai", "llamaindex", "dspy",
		"mcp", "agent", "reasoning", "automation", "rag",
	}

	for _, kw := range knownKeywords {
		if strings.Contains(searchSpace, kw) || strings.Contains(normalizedSpace, kw) {
			tagSet[kw] = struct{}{}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}
	return tags
}

// Normalize validates and maps RawSkill to a production-ready models.Skill.
func (n *Normalizer) Normalize(raw RawSkill) (*models.Skill, error) {
	name := strings.TrimSpace(raw.Name)
	if name == "" {
		if raw.RepoName != "" {
			name = raw.RepoName
		} else {
			return nil, errors.New("skill name or repository name is required")
		}
	}

	slug := n.GenerateSlug(name)
	platform := models.PlatformType(raw.Platform)
	if !platform.IsValid() {
		platform = n.DetectPlatform(raw.FilePath, raw.SourceURL, raw.RepoURL, name)
	}

	category := models.CategoryType(raw.Category)
	if !category.IsValid() {
		category = n.DetectCategory(platform, name, raw.Description, raw.ContentRaw)
	}

	subcategory := strings.TrimSpace(raw.Subcategory)
	if subcategory == "" {
		switch category {
		case models.CategoryMCPServers:
			subcategory = "tools"
		case models.CategoryRules:
			subcategory = "backend"
		case models.CategoryFrameworks:
			subcategory = "ai-ml"
		default:
			subcategory = "general"
		}
	}

	language := strings.TrimSpace(raw.Language)
	if language == "" {
		language = n.DetectLanguage(raw.FilePath, raw.ContentRaw)
	}

	sourceURL := strings.TrimSpace(raw.SourceURL)
	if sourceURL == "" {
		sourceURL = raw.RepoURL
	}
	if sourceURL == "" {
		sourceURL = fmt.Sprintf("https://github.com/%s/%s", raw.RepoOwner, raw.RepoName)
	}

	installSnippet := strings.TrimSpace(raw.InstallSnippet)
	if installSnippet == "" {
		installSnippet = n.ExtractInstallSnippet(platform, slug, sourceURL, raw.RepoName, raw.ContentRaw)
	}

	tags := raw.Tags
	if len(tags) == 0 {
		tags = n.ExtractTags(platform, language, name, raw.Description, raw.ContentRaw)
	}

	// Content preview: first 500 chars clean
	contentPreview := strings.TrimSpace(raw.ContentPreview)
	if contentPreview == "" {
		contentPreview = strings.TrimSpace(raw.Description)
	}
	if contentPreview == "" && raw.ContentRaw != "" {
		contentPreview = raw.ContentRaw
	}
	if len(contentPreview) > 500 {
		contentPreview = contentPreview[:497] + "..."
	}

	now := time.Now()
	lastSynced := raw.LastSyncedAt
	if lastSynced == nil {
		lastSynced = &now
	}

	skill := &models.Skill{
		Name:           name,
		Slug:           slug,
		Description:    strings.TrimSpace(raw.Description),
		Platform:       platform,
		Category:       category,
		Subcategory:    subcategory,
		SourceURL:      sourceURL,
		RepoURL:        strings.TrimSpace(raw.RepoURL),
		RepoOwner:      strings.TrimSpace(raw.RepoOwner),
		RepoName:       strings.TrimSpace(raw.RepoName),
		StarsCount:     raw.StarsCount,
		ForksCount:     raw.ForksCount,
		StarsVelocity:  raw.StarsVelocity,
		ContentRaw:     raw.ContentRaw,
		ContentPreview: contentPreview,
		InstallSnippet: installSnippet,
		FilePath:       strings.TrimSpace(raw.FilePath),
		Language:       language,
		Tags:           tags,
		LastSyncedAt:   lastSynced,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return skill, nil
}
