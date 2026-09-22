# AgentSight — Execution & Handover Summary

**Execution Date**: September 22, 2026  
**Pipeline Stage**: TASK_6 (Release, Verification & Handover)  
**Target Branch**: `main`  
**Deployment Namespace**: `agentsight`  
**Host Application Endpoint**: `http://localhost:8084` (`http://100.108.204.127:8084` via Tailscale)  
**Database Binding**: `127.0.0.1:5437` (isolated to localhost)  

---

## 1. Deployment Status

| Container Name | Service | Image | Status | Host Port Mapping | Healthcheck |
|---|---|---|---|---|---|
| `agentsight_app` | `app` | `agentsight-app` | `Up (healthy)` | `0.0.0.0:8084->8080/tcp` | `wget -qO- /healthz` (200 OK) |
| `agentsight_postgres` | `postgres` | `postgres:16-alpine` | `Up (healthy)` | `127.0.0.1:5437->5432/tcp` | `pg_isready` (connected) |

- **Docker Compose Project**: `agentsight`
- **Network**: `agentsight_net` (bridge)
- **Persistent Volume**: `agentsight_pgdata`
- **Zero Port Conflict**: Preflight verified ports `8084` and `5437` free before starting.
- **Port Exposure Audit**: DB port bound strictly to `127.0.0.1:5437` — zero external exposure (`0.0.0.0` or `::`).

---

## 2. Build & Test Timing

| Step | Duration | Result | Notes |
|---|---|---|---|
| **Preflight Port Verification** | ~0.5s | `SUCCESS` | Verified `APP_PORT=8084` & `DB_PORT=5437` available |
| **Go Unit & Component Tests** | ~0.20s | `14/14 PASS` | Handlers, middleware, scrapers, services, templates |
| **Docker Multi-Stage Build** | ~101.2s | `SUCCESS` | Stage 1 (Go 1.23 Alpine) + Stage 2 (Alpine 3.19 + Chromium) |
| **Database Migrations** | ~0.48s | `SUCCESS` | Embedded migrations 001, 002, 003 applied automatically |
| **Liveness & Readiness Probes** | < 1ms | `HTTP 200` | `/healthz` and `/readyz` returned healthy immediately |
| **verify.sh Automated Smoke Test** | ~2.1s | `14/14 PASS` | All 14 integration assertions passed with exit code 0 |

---

## 3. Smoke Test Verification Matrix (`./verify.sh`)

| # | Test Assertion | Target Route | Protocol | Expected | Status |
|---|---|---|---|---|---|
| 1 | Health check probe | `GET /healthz` | HTTP | 200 OK, `{"status":"ok"}` | ✅ PASS |
| 2 | Readiness check probe | `GET /readyz` | HTTP | 200 OK, `database: connected` | ✅ PASS |
| 3 | Landing page view | `GET /` | HTTP | 200 OK, contains "AgentSight" | ✅ PASS |
| 4 | Search engine execution | `GET /search?q=react` | HTTP | 200 OK, contains skill cards | ✅ PASS |
| 5 | Trending board view | `GET /trending` | HTTP | 200 OK, contains skill items | ✅ PASS |
| 6 | Platform filter view | `GET /platform/cursor` | HTTP | 200 OK, platform skills rendered | ✅ PASS |
| 7 | Category filter view | `GET /category/rules` | HTTP | 200 OK, category skills rendered | ✅ PASS |
| 8 | API skills list | `GET /api/v1/skills` | REST JSON | 200 OK, array of skills | ✅ PASS |
| 9 | API search endpoint | `GET /api/v1/search?q=mcp` | REST JSON | 200 OK, search results array | ✅ PASS |
| 10 | API trending endpoint | `GET /api/v1/trending` | REST JSON | 200 OK, trending skills array | ✅ PASS |
| 11 | API platform stats | `GET /api/v1/stats` | REST JSON | 200 OK, platform counts map | ✅ PASS |
| 12 | Skill detail page view | `GET /skill/{slug}` | HTTP | 200 OK, rule snippet rendered | ✅ PASS |
| 13 | Host port security audit | `docker port postgres 5432` | Host | Bound to `127.0.0.1`, not `0.0.0.0` | ✅ PASS |
| 14 | Seed data volume threshold | `GET /api/v1/skills` | REST JSON | Total skills indexed >= 50 | ✅ PASS |

**Final Smoke Test Outcome**: **14 Passed, 0 Failed (Exit Code: 0)**

---

## 4. Skills & Platform Index Metrics

As of deployment completion, the background ingestion engine has scraped and indexed live repositories in addition to seed migrations:

- **Total Skills Indexed**: `501` (exceeds seed threshold of 50 by 10x)
- **Total Stars Monitored**: `16,506,988`
- **Total Ingestion Sources**: `9`

### Platform Breakdown
- **Cursor (`cursor`)**: `254` skills / rules
- **Generic AI Agents (`generic`)**: `155` skills / frameworks
- **Model Context Protocol (`mcp`)**: `51` server implementations
- **Claude (`claude`)**: `15` AGENTS.md configurations
- **Gemini (`gemini`)**: `15` SKILL.md modules
- **GitHub Copilot (`copilot`)**: `11` prompt rules

### Category Breakdown
- **Rules (`rules`)**: `265` items
- **Skills (`skills`)**: `168` items
- **MCP Servers (`mcp_servers`)**: `51` items
- **Frameworks (`frameworks`)**: `8` items
- **Tools (`tools`)**: `6` items
- **Prompts (`prompts`)**: `3` items

---

## 5. Scraper Sources Status

| ID | Source Name | Type | Target URL / Repo | Sync Status |
|---|---|---|---|---|
| 1 | Awesome Cursor Rules | `github_repo` | `PatrickJS/awesome-cursorrules` | `Active (Synced)` |
| 2 | Awesome MCP Servers | `github_repo` | `punkpeye/awesome-mcp-servers` | `Active (Synced)` |
| 3 | Official MCP Servers | `github_repo` | `modelcontextprotocol/servers` | `Active (Synced)` |
| 4 | AGENTS.md Spec | `github_repo` | `agentsmd/agents.md` | `Active (Synced)` |
| 5 | Cursor Directory | `website` | `https://cursor.directory` | `Active (Synced)` |
| 6 | MCP Servers Org | `website` | `https://mcpservers.org` | `Active (Synced)` |
| 7 | GitHub Search - cursorrules | `github_search` | `filename:.cursorrules` | `Active (Rate-Limited Safe)` |
| 8 | GitHub Search - AGENTS.md | `github_search` | `filename:AGENTS.md` | `Active (Rate-Limited Safe)` |
| 9 | GitHub Search - SKILL.md | `github_search` | `filename:SKILL.md` | `Active (Rate-Limited Safe)` |

---

## 6. Artifacts Delivered

- `verify.sh`: Executable automated smoke test script with 14 end-to-end integration checks.
- `README.md`: Full architectural, API, scraper, environment, and operations documentation.
- `EXECUTION_SUMMARY.md`: Deployment audit, test matrix, and indexing statistics handover report.
- Active Docker runtime with zero host port leaks.
