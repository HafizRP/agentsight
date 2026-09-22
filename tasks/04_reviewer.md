# ROLE SPECIFICATION: STAFF ENGINEER & AUDITOR (AGENT 4)
# TASK ID: TASK_4
# TARGET BRANCH: main
# DEPENDENCIES: [TASK_2, TASK_3]

## 1. OBJECTIVE
Execute an independent code audit, run integration tests, and safely merge PRs into `main`.

## 2. AUDIT CHECKLIST
- [ ] Architecture Isolation: Handlers, services, repositories, scraper adapters, and templates remain decoupled. No circular imports.
- [ ] Zero Placeholders: Confirm no `TODO` annotations, blank mocks, or unhandled errors exist.
- [ ] chromedp Container Safety: Verify flags include `--no-sandbox`, `--disable-dev-shm-usage`, `--headless=new`, `--disable-gpu`.
- [ ] GitHub API Resilience: Rate limit handling, ETag caching, graceful degradation to seed data.
- [ ] SQL Injection Prevention: All queries use parameterized statements via pgx — no string concatenation.
- [ ] Template Security: All user-generated content escaped in HTML templates.
- [ ] Test Suite Validation: Execute `go test -v ./...` — all GREEN.
- [ ] Search Quality: Full-text search returns relevant results, filters work correctly.
- [ ] HTMX Endpoint Consistency: All `hx-get`/`hx-post` attributes reference real handler routes.

## 3. EXECUTION PROTOCOL
1. If authenticated via GitHub CLI: audit PR diffs → `gh pr review --approve` → `gh pr merge --squash --delete-branch`.
2. If local fallback: squash-merge feature branches directly into `main`.
3. After all merges: run `go build ./...` and `go test -v ./...` on `main` to catch integration issues.
4. Fix any issues found — commit with `fix(<scope>): <description>`.
