# ROLE SPECIFICATION: BACKEND ENGINEER (AGENT 1)
# TASK ID: TASK_1
# TARGET BRANCH: feat/backend-api
# DEPENDENCIES: []

## 1. OBJECTIVE
Build the backend architecture for AgentSight: database schema, data access layer, search engine, REST API, and authentication.

## 2. SCOPE OF IMPLEMENTATION

### Go Module & Entry Point
- Initialize `go mod init agentsight`.
- `cmd/server/main.go`: Bootstrap — load config from env, init pgxpool, register routes, start HTTP server on `:8080`.

### Database Migrations (`migrations/`)

**`001_init_schema.sql`**:
```sql
-- Platforms: cursor, claude, gemini, mcp, copilot, generic
CREATE TYPE platform_type AS ENUM ('cursor', 'claude', 'gemini', 'mcp', 'copilot', 'generic');

-- Categories: rules, skills, mcp_servers, prompts, frameworks, tools
CREATE TYPE category_type AS ENUM ('rules', 'skills', 'mcp_servers', 'prompts', 'frameworks', 'tools');

-- skills: the core entity — one record per discovered agent skill
skills (
  id SERIAL PK,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) UNIQUE NOT NULL,
  description TEXT,
  platform platform_type NOT NULL,
  category category_type NOT NULL,
  subcategory VARCHAR(100),            -- e.g., "frontend", "backend", "devops", "ai-ml"
  source_url VARCHAR(500) NOT NULL,    -- GitHub raw URL or website
  repo_url VARCHAR(500),               -- GitHub repo URL
  repo_owner VARCHAR(100),
  repo_name VARCHAR(200),
  stars_count INT DEFAULT 0,
  forks_count INT DEFAULT 0,
  stars_velocity INT DEFAULT 0,        -- stars gained in last 7 days (for trending)
  content_raw TEXT,                     -- raw file content (for search)
  content_preview VARCHAR(500),        -- truncated preview
  install_snippet TEXT,                 -- copy-paste install/setup command
  file_path VARCHAR(500),              -- path within repo (e.g., ".cursor/rules/react.mdc")
  language VARCHAR(50),                -- target language/framework
  tags TEXT[],                         -- array of tags
  last_synced_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- sources: scraper targets registry
sources (
  id SERIAL PK,
  name VARCHAR(255) NOT NULL,
  url VARCHAR(500) NOT NULL,
  source_type VARCHAR(50) NOT NULL,    -- "github_repo", "github_search", "website", "awesome_list"
  scrape_strategy VARCHAR(50),         -- "api", "chromedp", "raw_fetch"
  is_active BOOLEAN DEFAULT true,
  last_scraped_at TIMESTAMPTZ,
  scrape_interval_hours INT DEFAULT 24,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- users: for bookmarks & personalization
users (
  id SERIAL PK,
  github_id VARCHAR(50) UNIQUE,
  username VARCHAR(100) NOT NULL,
  avatar_url VARCHAR(500),
  email VARCHAR(255),
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- bookmarks: user-skill favorites
bookmarks (
  id SERIAL PK,
  user_id INT FK→users ON DELETE CASCADE,
  skill_id INT FK→skills ON DELETE CASCADE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, skill_id)
);

-- scrape_logs: audit trail for scraper runs
scrape_logs (
  id SERIAL PK,
  source_id INT FK→sources,
  status VARCHAR(20),                  -- "success", "failed", "partial"
  skills_found INT DEFAULT 0,
  skills_new INT DEFAULT 0,
  skills_updated INT DEFAULT 0,
  error_message TEXT,
  duration_ms INT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

Indexes: `skills.slug`, `skills.platform`, `skills.category`, `skills.subcategory`, `skills.stars_count DESC`, GIN index on `skills.tags`, full-text search index on `skills.name || skills.description || skills.content_raw`.

### Data Access Layer (`internal/repository/`)
- `skill_repo.go`: CRUD, search (full-text + filters), trending query (ORDER BY stars_velocity DESC), list by category/platform, get by slug.
- `source_repo.go`: CRUD for scraper sources, update last_scraped_at.
- `user_repo.go`: Find/create by GitHub ID, profile lookup.
- `bookmark_repo.go`: Toggle bookmark, list user bookmarks, check if bookmarked.
- `scrape_log_repo.go`: Create log entry, list recent logs.
- Connection pool via `jackc/pgx/v5/pgxpool`.

### Service Layer (`internal/service/`)
- `search_service.go`: Full-text search with platform/category/language filters, pagination, sort options (relevance, stars, newest, trending).
- `skill_service.go`: Get skill detail with bookmark status, related skills by tags/category.
- `auth_service.go`: GitHub OAuth flow (redirect → callback → create/find user → set session cookie).
- `trending_service.go`: Calculate stars velocity, generate trending board.

### HTTP Handlers (`internal/handlers/`)
Using `go-chi/chi/v5` router:
```
GET    /                                → Landing page (trending skills)
GET    /search?q=&platform=&category=   → Search results page
GET    /skill/{slug}                     → Skill detail page
GET    /trending                         → Trending board
GET    /category/{category}              → Browse by category
GET    /platform/{platform}              → Browse by platform

GET    /api/v1/skills                    → JSON: paginated skills list
GET    /api/v1/skills/{slug}             → JSON: skill detail
GET    /api/v1/search                    → JSON: search results
GET    /api/v1/trending                  → JSON: trending skills
POST   /api/v1/bookmarks/{skill_id}      → JSON: toggle bookmark (auth required)
GET    /api/v1/bookmarks                 → JSON: user's bookmarks (auth required)
GET    /api/v1/stats                     → JSON: platform/category counts

GET    /auth/github                      → OAuth redirect to GitHub
GET    /auth/github/callback             → OAuth callback
POST   /auth/logout                      → Clear session

GET    /healthz                          → 200 OK
GET    /readyz                           → 200 OK (DB healthy)
```

### Middleware (`internal/middleware/`)
- `session_middleware.go`: Cookie-based session management.
- `auth_middleware.go`: Extract user from session, inject into context (optional — unauthenticated users can still browse).
- `ratelimit_middleware.go`: Basic in-memory rate limiting for API endpoints.

### Tests
- Unit tests for search service (filter combinations, pagination edge cases).
- Unit tests for skill service (slug generation, related skills).
- Integration tests for auth flow.
- Run `go test -v ./...` — all GREEN.

## 3. ARTIFACT CONTRACT REQUIREMENT
- [ ] Export all endpoints, HTTP methods, and request/response structures to `contracts/api_contract.json`.
- [ ] Document all environment variables in `.env.example` (`DATABASE_URL`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `SESSION_SECRET`, `APP_PORT`).

## 4. VERIFICATION CRITERIA
1. `go build ./...` compiles cleanly.
2. `go test -v ./...` passes with exit code 0.
3. `contracts/api_contract.json` exists and is valid JSON.

## 5. EXIT PROTOCOL
1. Commit all changes to `feat/backend-api`.
2. Push branch and open a PR via `gh pr create --title "feat(backend): API, DB Schema & Search Engine" --body "..."`.
3. Record task completion artifacts in local state.
