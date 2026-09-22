# AgentSight — AI Agent Skill Discovery Engine

AgentSight is a high-performance, full-stack discovery engine and registry for AI agent configurations, skills, and tools. It aggregates, indexes, and ranks Cursor rules (`.cursorrules`), Claude configurations (`AGENTS.md`, `CLAUDE.md`), Gemini skills (`SKILL.md`), and Model Context Protocol (MCP) servers across GitHub repositories and curated registries.

---

## 🏛️ System Architecture

```
                                 +--------------------------------------------------+
                                 |                   Client Web                     |
                                 |         (HTMX 1.9 + Tailwind CSS + SSR)          |
                                 +-------------------------+------------------------+
                                                           |
                                                  HTTP / SSE / REST
                                                           |
                                                           v
+---------------------------------------------------------------------------------------------------------+
|                                        AgentSight Core Application                                      |
|                                                                                                         |
|  +---------------------------+  +---------------------------+  +-------------------------------------+  |
|  |     chi/v5 HTTP Router    |  |     Template Renderer     |  |          Auth & Middleware          |  |
|  |  - /healthz, /readyz      |  |  - Embedded templates     |  |  - Session cookie management         |  |
|  |  - Web UI routes          |  |  - HTMX partial swaps     |  |  - Rate limiter (token bucket)      |  |
|  |  - REST API v1 endpoints  |  |  - Dynamic icon styling   |  |  - GitHub OAuth 2.0 flow            |  |
|  +-------------+-------------+  +---------------------------+  +-------------------------------------+  |
|                |                                                                                        |
|                v                                                                                        |
|  +---------------------------+  +---------------------------+  +-------------------------------------+  |
|  |     Service Layer         |  |     Discovery Scrapers    |  |      Scraper Scheduler Workers      |  |
|  |  - SkillService           |  |  - GitHub Repo Scraper    |  |  - Interval check (every 15 min)    |  |
|  |  - SearchService (FTS)    |  |  - GitHub Code Search     |  |  - Circuit breaker & rate limiting  |  |
|  |  - TrendingService        |  |  - Awesome List Parser    |  |  - Safe container Chromium fallback |  |
|  |  - AuthService            |  |  - Website (chromedp)     |  |  - Exponential backoff retry        |  |
|  +-------------+-------------+  +-------------+-------------+  +------------------+------------------+  |
|                |                              |                                   |                     |
|                +------------------------------+-----------------------------------+                     |
|                                               |                                                         |
|                                               v                                                         |
|                               +-------------------------------+                                         |
|                               |      jackc/pgx/v5 Pool        |                                         |
|                               +---------------+---------------+                                         |
+-----------------------------------------------|---------------------------------------------------------+
                                                |
                                                v
                               +----------------------------------+
                               |     PostgreSQL 16 Database       |
                               |  - Schema migrations             |
                               |  - Full-Text Search (tsvector)   |
                               |  - Trigram indexes (pg_trgm)     |
                               |  - Isolated port binding 5437    |
                               +----------------------------------+
```

---

## 🌟 Key Features

- **Multi-Platform Skill Indexing**: Unified registry covering Cursor (`.cursorrules`), Claude (`AGENTS.md`, `CLAUDE.md`), Gemini (`SKILL.md`), and Model Context Protocol (`mcp_servers`).
- **Hybrid Full-Text Search**: Powered by PostgreSQL `tsvector` weighted full-text search combined with `pg_trgm` fuzzy similarity matching.
- **Trending & Velocity Metrics**: Ranks skills by star velocity (`stars_velocity`) and absolute GitHub engagement to surface breakout developer tools.
- **HTMX Reactive UI**: SPA-like interactivity with zero JavaScript bundle fatigue, dynamic filtering by platform and category, instant live search, and inline bookmark toggling.
- **Automated Ingestion Pipeline**: Background scheduler scrapes GitHub repositories, Markdown awesome-lists, code search APIs, and web directories using container-resilient headless Chromium (`chromedp`).
- **Resilient Fallback Engine**: Gracefully falls back to curated offline seed datasets (`migrations/002_seed_skills.sql`) whenever external network or GitHub rate limits are encountered.
- **Strict Network Isolation**: PostgreSQL database binds exclusively to `127.0.0.1:${DB_PORT:-5437}`, preventing external port exposure.

---

## 🛠️ Tech Stack

| Component | Technology | Rationale |
|---|---|---|
| **Backend** | Go 1.22+ (`go-chi/chi/v5`) | Minimalist routing, sub-millisecond latency, zero-allocation handlers |
| **Database Pool** | `jackc/pgx/v5/pgxpool` | High-concurrency native PostgreSQL driver with connection pooling |
| **Database** | PostgreSQL 16 Alpine | Native full-text search (`tsvector`), GIN indexes, trigram fuzzy matching |
| **Frontend** | HTMX 1.9.12 + Tailwind CSS | Reactive hypermedia UI rendered via embedded Go `html/template` |
| **Scraper** | `chromedp` + Go HTTP Client | Headless browser execution with container-safe sandbox flags |
| **Containerization** | Docker & Docker Compose | Multi-stage build with pinned Alpine runner and isolated bridge network |

---

## 🚀 Quick Start & Docker Deployment

### 1. Prerequisites
- Docker Engine 24.0+ and Docker Compose v2.20+
- `curl` and `jq` (for verification)

### 2. Environment Setup
Generate `.env` from `.env.example`:
```bash
cp .env.example .env
```

### 3. Run Preflight Verification
Verify that required ports (`8084` for web and `5437` for DB) are free:
```bash
./scripts/preflight.sh
```

### 4. Build and Start Containers
```bash
docker compose -p agentsight up -d --build
```

### 5. Verify Health
Wait until both services report healthy:
```bash
curl -i http://localhost:8084/healthz
curl -i http://localhost:8084/readyz
```

### 6. Run Automated Smoke Test Suite
```bash
chmod +x verify.sh
./verify.sh
```

---

## 📡 API Reference

### Health Probes
- `GET /healthz`: Liveness check. Returns `{"status":"ok"}` (HTTP 200).
- `GET /readyz`: Readiness check. Verifies PostgreSQL connectivity. Returns `{"status":"ready","database":"connected"}` (HTTP 200).

### Web Views
- `GET /`: Landing page with hero search, platform cards, category grid, and trending skills.
- `GET /search?q={query}&platform={platform}&category={category}&page={page}`: Full-text search view with HTMX partial swap support (`#search-results`).
- `GET /trending`: Dedicated trending board ranked by 7-day velocity and total stars.
- `GET /platform/{platform}`: Skills filtered by platform (`cursor`, `claude`, `gemini`, `mcp`, `copilot`, `generic`).
- `GET /category/{category}`: Skills filtered by category (`rules`, `skills`, `mcp_servers`, `prompts`, `frameworks`, `tools`).
- `GET /skill/{slug}`: Comprehensive detail view containing raw rule configuration and one-click install snippets.
- `GET /login`: GitHub OAuth login interface.

### REST API v1
All endpoints return JSON and are rate-limited to 60 req/min per IP.

#### 1. List Skills
```http
GET /api/v1/skills?page=1&limit=20&platform=cursor&category=rules
```
Response:
```json
{
  "skills": [
    {
      "id": 1,
      "name": "FastAPI Async SQLAlchemy 2.0 Cursor Rule",
      "slug": "fastapi-async-sqlalchemy-cursor-rule",
      "description": "Production conventions for FastAPI with SQLAlchemy 2.0 async sessions...",
      "platform": "cursor",
      "category": "rules",
      "stars_count": 2840,
      "stars_velocity": 120,
      "install_snippet": "curl -o .cursorrules https://agentsight.dev/r/fastapi-async-sqlalchemy",
      "tags": ["fastapi", "python", "async", "sqlalchemy"]
    }
  ],
  "total": 501,
  "page": 1,
  "limit": 20,
  "total_pages": 26
}
```

#### 2. Get Skill Detail
```http
GET /api/v1/skills/{slug}
```
Response:
```json
{
  "id": 1,
  "name": "FastAPI Async SQLAlchemy 2.0 Cursor Rule",
  "slug": "fastapi-async-sqlalchemy-cursor-rule",
  "content_raw": "# FastAPI Production Guidelines...",
  "install_snippet": "curl -o .cursorrules ...",
  "is_bookmarked": false
}
```

#### 3. Search Skills
```http
GET /api/v1/search?q=mcp&limit=10
```

#### 4. Get Trending Skills
```http
GET /api/v1/trending?limit=10
```

#### 5. Get Ecosystem Statistics
```http
GET /api/v1/stats
```
Response:
```json
{
  "total_skills": 501,
  "total_stars": 16506988,
  "total_sources": 9,
  "platform_counts": {
    "claude": 15,
    "copilot": 11,
    "cursor": 254,
    "gemini": 15,
    "generic": 155,
    "mcp": 51
  },
  "category_counts": {
    "frameworks": 8,
    "mcp_servers": 51,
    "prompts": 3,
    "rules": 265,
    "skills": 168,
    "tools": 6
  }
}
```

#### 6. Bookmark Management (Requires Session Auth)
- `POST /api/v1/bookmarks/{skill_id}`: Toggle bookmark status.
- `GET /api/v1/bookmarks`: List authenticated user's bookmarks.

---

## 🕷️ Scraper Architecture & Sources

AgentSight runs an autonomous background scheduler (`internal/scraper/scheduler.go`) configured to synchronize sources on a 15-minute cadence:

1. **GitHub Repository Scraper (`github_repo_scraper.go`)**:
   - Ingests repositories like `PatrickJS/awesome-cursorrules`, `punkpeye/awesome-mcp-servers`, `modelcontextprotocol/servers`, and `agentsmd/agents.md`.
   - Uses file size guards (max 500KB) and markdown AST parsing.
2. **GitHub Search Scraper (`github_search_scraper.go`)**:
   - Queries code search for `.cursorrules`, `AGENTS.md`, `CLAUDE.md`, and `SKILL.md`.
   - Enforces rate-limiting resilience: logs warnings and gracefully defers execution when unauthenticated quota is reached.
3. **Headless Website Scraper (`website_scraper.go`)**:
   - Uses `chromedp` with container-safe flags: `--no-sandbox`, `--disable-dev-shm-usage`, `--headless=new`, `--disable-gpu`.
   - Automatically degrades to fast HTTP DOM extraction if Chromium is unavailable.
4. **Awesome-List Markdown Parser (`awesome_list_parser.go`)**:
   - Extracts structured entries from bullet lists and GitHub-flavored markdown tables with regex-based link normalization.

---

## ⚙️ Configuration & Environment Variables

| Variable | Default Value | Description |
|---|---|---|
| `APP_PORT` | `8084` | Host port exposed by the AgentSight web container |
| `DB_PORT` | `5437` | Host port mapped to PostgreSQL (strictly bound to `127.0.0.1`) |
| `DB_USER` | `agentsight` | PostgreSQL database user |
| `DB_PASSWORD` | `agentsight` | PostgreSQL database password |
| `DB_NAME` | `agentsight` | PostgreSQL database name |
| `DATABASE_URL` | `postgres://...` | Full PostgreSQL connection string |
| `SESSION_SECRET` | *random 32 bytes* | Secret key used for secure cookie encryption |
| `BASE_URL` | `http://localhost:8084` | Public base URL for OAuth callbacks |
| `APP_ENV` | `development` | Application mode (`development`, `production`, `test`) |
| `GITHUB_TOKEN` | *(optional)* | Personal Access Token to raise GitHub API rate limits from 60 to 5,000 req/hr |
| `GITHUB_CLIENT_ID` | *(optional)* | GitHub OAuth application client ID |
| `GITHUB_CLIENT_SECRET` | *(optional)* | GitHub OAuth application client secret |
| `CHROME_BIN` | `/usr/bin/chromium-browser` | Path to Chromium executable inside the runner container |

---

## 🧪 Testing & Verification

### Unit and Integration Tests
```bash
export PATH="/home/b14/go1.23/bin:$PATH"
go test -v ./...
```

### Live Smoke Test (`verify.sh`)
Validates 14 critical end-to-end paths against the live container cluster:
```bash
./verify.sh
```
Tests include:
1. Health check (`GET /healthz`)
2. Readiness check (`GET /readyz`)
3. Landing page render & title check
4. Search page query execution (`/search?q=react`)
5. Trending board retrieval (`/trending`)
6. Platform filtering (`/platform/cursor`)
7. Category filtering (`/category/rules`)
8. API skills array validation (`/api/v1/skills`)
9. API search JSON schema validation (`/api/v1/search?q=mcp`)
10. API trending skills validation (`/api/v1/trending`)
11. API platform metrics validation (`/api/v1/stats`)
12. Skill detail page view (`/skill/{slug}`)
13. Database port security check (host `127.0.0.1:5437` only, no `0.0.0.0` exposure)
14. Seed data volume threshold (>= 50 indexed items verified)

---

## 🤝 Contributing

1. Fork and clone the repository.
2. Ensure Go 1.22+ is available.
3. Create a feature branch: `git checkout -b feat/your-feature`.
4. Run tests: `go test -v ./...`.
5. Run lint & preflight checks: `./scripts/preflight.sh`.
6. Submit a GitHub Pull Request following Conventional Commits format (`feat(...)`, `fix(...)`).

---

## 📄 License

Licensed under the MIT License.
