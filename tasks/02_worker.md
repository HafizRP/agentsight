# ROLE SPECIFICATION: SCRAPER & DATA ENGINEER (AGENT 2)
# TASK ID: TASK_2
# TARGET BRANCH: feat/scraper-engine
# DEPENDENCIES: [TASK_1]

## 1. OBJECTIVE
Build the multi-source scraping engine that discovers, extracts, and indexes AI agent skills from GitHub repos, awesome lists, and community directories.

## 2. SCOPE OF IMPLEMENTATION

### Scraper Architecture (`internal/scraper/`)

The scraper is a pipeline: **Discover → Fetch → Parse → Normalize → Persist**.

#### Source Adapters
Each source type has a dedicated adapter implementing a common `Scraper` interface:

```go
type Scraper interface {
    Discover(ctx context.Context) ([]RawSkill, error)
    Name() string
}
```

**`github_search_scraper.go`** — GitHub Code Search API:
- Search for agent config files across all of GitHub:
  - `filename:.cursorrules` → Cursor legacy rules
  - `filename:AGENTS.md` → Claude agent configs
  - `filename:CLAUDE.md` → Claude project configs
  - `filename:SKILL.md path:.gemini` → Gemini/AGY skills
  - `filename:mcp_config.json` → MCP configurations
- Use GitHub REST API (`api.github.com/search/code`).
- Pagination: process up to 10 pages per query (1000 results).
- Respect rate limits: check `X-RateLimit-Remaining` headers.
- If `GITHUB_TOKEN` env var is set: authenticated (5000 req/hr). If absent: unauthenticated (60 req/hr) with aggressive backoff.

**`github_repo_scraper.go`** — Target specific known repos:
- `PatrickJS/awesome-cursorrules` → Parse README links, fetch each linked `.mdc`/`.cursorrules` file.
- `punkpeye/awesome-mcp-servers` → Parse README table rows, extract name/description/URL.
- `modelcontextprotocol/servers` → Official MCP server directory.
- `agentsmd/agents.md` → AGENTS.md specification examples.
- `awesome-ai-agents-2026` → Framework/tool listings.
- Fetch repo metadata (stars, forks, last_updated) via GitHub API.

**`website_scraper.go`** — Web directories (chromedp):
- `cursor.directory` → Scrape skill cards, names, descriptions, install commands.
- `mcpservers.org` → Scrape MCP server listings.
- Use chromedp with container-safe flags.
- 30-second timeout per page with retry logic.

**`awesome_list_parser.go`** — Generic awesome-list markdown parser:
- Parse GitHub README.md files containing curated link lists.
- Extract: link text (skill name), URL, description text after the link.
- Detect categories from markdown headings (`## Category Name`).

#### Content Fetcher (`internal/scraper/fetcher.go`)
- Fetch raw file content from GitHub via raw.githubusercontent.com URLs.
- Cache fetched content with ETag/If-None-Match for incremental syncing.
- Respect `content_length` limits (skip files > 500KB).

#### Normalizer (`internal/scraper/normalizer.go`)
- Detect `platform` from file path / source:
  - `.cursor/rules/*.mdc` or `.cursorrules` → `cursor`
  - `AGENTS.md` or `CLAUDE.md` → `claude`
  - `.gemini/*/SKILL.md` → `gemini`
  - `mcp_config.json` or MCP server repos → `mcp`
  - `.github/copilot-instructions.md` → `copilot`
- Auto-detect `category` from content patterns.
- Generate `slug` from name (lowercase, kebab-case, deduplication).
- Extract `install_snippet` from README install sections.
- Detect `language` / framework from content keywords.
- Auto-tag from content analysis.

### Cron Scheduler (`internal/scraper/scheduler.go`)
- On application startup, launch a background goroutine.
- Check `sources` table: any source where `NOW() - last_scraped_at > scrape_interval_hours`.
- Execute matching scrapers, update `last_scraped_at`, write `scrape_logs`.
- Calculate `stars_velocity` weekly: `current_stars - stars_from_7_days_ago`.

### Seed Data (`migrations/002_seed_skills.sql`)
Pre-populate with at least 100 curated skills:
- 30 Cursor Rules (React, Next.js, Vue, Tailwind, Python, Go, Rust, TypeScript, etc.)
- 20 MCP Servers (Filesystem, GitHub, Postgres, Playwright, Notion, Slack, etc.)
- 15 Claude AGENTS.md examples (various project types)
- 15 Gemini Skills (coding workflows, debugging, deployment)
- 10 Copilot instructions
- 10 Agent Framework tools

### Source Registry (`migrations/003_seed_sources.sql`)
Pre-populate the `sources` table with all known scraping targets:
```sql
INSERT INTO sources (name, url, source_type, scrape_strategy, scrape_interval_hours) VALUES
('Awesome Cursor Rules', 'https://github.com/PatrickJS/awesome-cursorrules', 'github_repo', 'api', 24),
('Awesome MCP Servers', 'https://github.com/punkpeye/awesome-mcp-servers', 'github_repo', 'api', 24),
('Official MCP Servers', 'https://github.com/modelcontextprotocol/servers', 'github_repo', 'api', 48),
('AGENTS.md Spec', 'https://github.com/agentsmd/agents.md', 'github_repo', 'api', 48),
('Cursor Directory', 'https://cursor.directory', 'website', 'chromedp', 24),
('MCP Servers Org', 'https://mcpservers.org', 'website', 'chromedp', 24),
('GitHub Code Search - cursorrules', 'https://api.github.com/search/code?q=filename:.cursorrules', 'github_search', 'api', 72),
('GitHub Code Search - AGENTS.md', 'https://api.github.com/search/code?q=filename:AGENTS.md', 'github_search', 'api', 72),
('GitHub Code Search - SKILL.md gemini', 'https://api.github.com/search/code?q=filename:SKILL.md+path:.gemini', 'github_search', 'api', 72);
```

### Artifact Contract Requirement
- [ ] Export scraper source definitions to `contracts/scraper_sources.json`.
- [ ] Document `GITHUB_TOKEN`, `CHROME_BIN` variables in `.env.example`.

## 3. VERIFICATION CRITERIA
1. Unit tests for normalizer (platform detection, slug generation, tag extraction).
2. Unit tests for awesome_list_parser with sample markdown input.
3. Integration test for GitHub API scraper with mock HTTP responses.
4. `go test -v ./...` passes with exit code 0.

## 4. EXIT PROTOCOL
1. Commit to `feat/scraper-engine`.
2. Push branch and open PR via `gh pr create --title "feat(scraper): Multi-source Agent Skill Scraper Engine" --body "..."`.
