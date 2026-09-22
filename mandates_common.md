# [OPERATIONAL MANDATES & CODE INTEGRITY RULES]
# PROJECT: AgentSight — AI Agent Skill Discovery Engine
# TECH STACK: Go 1.22+ (Chi Router) + PostgreSQL 16 + HTMX + Tailwind CSS + chromedp + Docker

1. ZERO HUMAN INTERVENTION (IN-TASK):
   - Do NOT halt execution to request clarifications. Make technical decisions autonomously within the specification.
2. PRODUCTION READY (ZERO PLACEHOLDERS):
   - Strictly NO `TODO` comments, empty mock functions, unhandled errors, or stubs. All deliverables must be 100% complete and working.
3. WRITE DIRECTLY TO DISK:
   - Create and modify real files directly in the project working directory (`./`).
4. GO CONVENTIONS:
   - Use Go 1.22+ with `go-chi/chi/v5` for routing.
   - Database access via `jackc/pgx/v5/pgxpool` connection pool.
   - Project layout: `cmd/server/`, `internal/{handlers,repository,service,scraper,generator,templates,middleware}/`.
   - Structured logging via `log/slog`.
   - Use `html/template` for server-side rendering.
   - Scraping via `chromedp` with container-safe flags: `--no-sandbox`, `--disable-dev-shm-usage`, `--headless=new`, `--disable-gpu`.
5. PORT & MULTI-TENANT ISOLATION:
   - App port: `${APP_PORT:-8084}:8080` (avoids common port collisions).
   - PostgreSQL: `127.0.0.1:${DB_PORT:-5437}:5432` (NEVER bind host port 5432).
   - All Docker resources MUST use the scoped project namespace: `name: agentsight`.
   - Container names: `agentsight_app`, `agentsight_postgres`.
6. SELF-HEALING & VERIFICATION:
   - Run `go test -v ./...` before flagging a task as complete.
   - Automatically diagnose and fix errors until test exit code is 0.
7. GIT HYGIENE:
   - Use Conventional Commits format: `<type>(<scope>): <description>`.
   - Create `.gitignore` and `.hermesignore` at start (exclude: `node_modules/`, `.git/`, `bin/`, `tmp/`, `volumes/`, `logs/`, `.env`).
8. CHROMEDP & CONTAINER RESILIENCE:
   - Headless Chromium MUST include safe container flags: `--no-sandbox`, `--disable-dev-shm-usage`, `--headless=new`, `--disable-gpu`.
   - `CHROME_BIN` env var support (default: `/usr/bin/chromium-browser`).
   - 30-second timeout per scraping cycle with exponential backoff retry (max 3 attempts).
9. GITHUB API RATE LIMITING:
   - If `GITHUB_TOKEN` env var is set: use authenticated requests (5000 req/hr).
   - If absent: use unauthenticated requests (60 req/hr) with aggressive caching and backoff.
   - NEVER crash on rate limit — gracefully degrade to cached data.
10. EXTERNAL API RESILIENCE:
    - All external HTTP calls must have timeouts (max 30s), retries (max 3), and fallback to cached/seeded data.
    - If scraping fails entirely: seed from a curated local dataset (`migrations/002_seed_skills.sql`).
