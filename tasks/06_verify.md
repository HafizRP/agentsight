# ROLE SPECIFICATION: RELEASE & QA ENGINEER (AGENT 6)
# TASK ID: TASK_6
# TARGET BRANCH: main
# DEPENDENCIES: [TASK_5]

## 1. OBJECTIVE
Conduct end-to-end smoke testing against running containers, audit port exposure, and generate handover documentation.

## 2. SCOPE OF IMPLEMENTATION

### Deployment
1. Generate `.env` from `.env.example` with secure defaults.
2. Run `scripts/preflight.sh` to verify port availability.
3. Execute `docker compose -p agentsight up -d --build`.
4. Poll `http://localhost:${APP_PORT:-8084}/healthz` every 5s until HTTP 200 (max 120s).
5. Poll `http://localhost:${APP_PORT:-8084}/readyz` until HTTP 200.

### Automated Smoke Test (`verify.sh`)
```bash
#!/usr/bin/env bash
set -euo pipefail
BASE_URL="http://127.0.0.1:${APP_PORT:-8084}"
PASS=0; FAIL=0

# Test 1: Health check — GET /healthz → 200
# Test 2: Readiness check — GET /readyz → 200
# Test 3: Landing page — GET / → 200, contains "AgentSight" in body
# Test 4: Search — GET /search?q=react → 200, contains skill results
# Test 5: Trending — GET /trending → 200, contains skill cards
# Test 6: Browse by platform — GET /platform/cursor → 200
# Test 7: Browse by category — GET /category/rules → 200
# Test 8: API skills list — GET /api/v1/skills → 200, valid JSON array
# Test 9: API search — GET /api/v1/search?q=mcp → 200, valid JSON
# Test 10: API trending — GET /api/v1/trending → 200, valid JSON
# Test 11: API stats — GET /api/v1/stats → 200, JSON with platform counts
# Test 12: Skill detail — GET /skill/{first-slug-from-api} → 200
# Test 13: Port security — DB port bound to 127.0.0.1, NOT 0.0.0.0
# Test 14: Seed data check — API returns >= 50 skills total

echo "Results: ${PASS} passed, ${FAIL} failed"
[ ${FAIL} -eq 0 ] && exit 0 || exit 1
```
- `chmod +x verify.sh && ./verify.sh` must return exit 0.
- If any test fails: debug, fix, rebuild, re-run.

### Documentation
- **`README.md`**: Architecture overview, feature list, tech stack, API reference, scraper documentation, Docker setup guide, environment variables, contributing guide.
- **`EXECUTION_SUMMARY.md`**: Build timing, test results, deployment status, skills indexed count, scraper source count.

## 3. VERIFICATION CRITERIA
1. `./verify.sh` returns exit code 0.
2. No port exposure warnings in `docker compose logs`.
3. Seed data is visible on the landing page.

## 4. EXIT PROTOCOL
1. Commit `verify.sh`, `README.md`, and `EXECUTION_SUMMARY.md` to `main`.
2. Push to remote if authenticated.
