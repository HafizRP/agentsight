-- Migrations: 002_seed_skills.sql
-- Curated seed skills dataset

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Next.js 14 App Router Cursor Rule',
  'nextjs-14-app-router-cursor-rule',
  'Best practices for Next.js 14 Server Actions, Route Handlers, and React Server Components.',
  'cursor',
  'rules',
  'frontend',
  'https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/nextjs-14.mdc',
  'https://github.com/PatrickJS/awesome-cursorrules',
  'PatrickJS',
  'awesome-cursorrules',
  4820,
  512,
  140,
  '# Next.js 14 Standards
- Favor React Server Components (RSC) over Client Components.
- Use Server Actions for data mutations with useActionState.
- Validate all inputs using Zod.
- Co-locate client and server logic cleanly.',
  'Favor React Server Components (RSC) over Client Components. Use Server Actions for data mutations with useActionState.',
  'curl -o .cursor/rules/nextjs-14.mdc https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/nextjs-14.mdc',
  '.cursor/rules/nextjs-14.mdc',
  'typescript',
  ARRAY['nextjs', 'react', 'typescript', 'cursor']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'FastAPI & Pydantic V2 Rule',
  'fastapi-pydantic-v2-rule',
  'Production patterns for FastAPI with async SQLAlchemy, dependency injection, and Pydantic v2 schemas.',
  'cursor',
  'rules',
  'backend',
  'https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/fastapi.mdc',
  'https://github.com/PatrickJS/awesome-cursorrules',
  'PatrickJS',
  'awesome-cursorrules',
  3210,
  280,
  95,
  '# FastAPI Production Guidelines
- Always define return types and response models explicitly.
- Utilize Depends for database sessions and auth claims.
- Handle errors via custom HTTPException subclasses.',
  'Always define return types and response models explicitly. Utilize Depends for database sessions and auth claims.',
  'curl -o .cursor/rules/fastapi.mdc https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/fastapi.mdc',
  '.cursor/rules/fastapi.mdc',
  'python',
  ARRAY['fastapi', 'python', 'pydantic', 'backend']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Go Chi & Clean Architecture Rule',
  'go-chi-clean-architecture-rule',
  'Clean domain-driven design principles for Go web services using Chi router and pgx.',
  'cursor',
  'rules',
  'backend',
  'https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/go-chi.mdc',
  'https://github.com/PatrickJS/awesome-cursorrules',
  'PatrickJS',
  'awesome-cursorrules',
  2980,
  190,
  85,
  '# Go Chi Service Guidelines
- Structure in cmd/, internal/handlers, internal/service, internal/repository.
- Never ignore errors; log with log/slog.
- Use context.Context as first argument in all I/O methods.',
  'Structure in cmd/, internal/handlers, internal/service, internal/repository. Never ignore errors; log with slog.',
  'curl -o .cursor/rules/go-chi.mdc https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/go-chi.mdc',
  '.cursor/rules/go-chi.mdc',
  'go',
  ARRAY['go', 'chi', 'backend', 'clean-architecture']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Tailwind CSS v4 & Lucide Icons Rule',
  'tailwind-css-v4-lucide-icons-rule',
  'Rules for responsive, dark-mode first component design with Tailwind CSS and Lucide icons.',
  'cursor',
  'rules',
  'frontend',
  'https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/tailwind.mdc',
  'https://github.com/PatrickJS/awesome-cursorrules',
  'PatrickJS',
  'awesome-cursorrules',
  2450,
  160,
  70,
  '# Tailwind CSS Guidelines
- Build dark-first with slate/zinc color scales.
- Ensure mobile responsiveness with min-w-0 on flex children.
- Extract reusable badge and button patterns.',
  'Build dark-first with slate/zinc color scales. Ensure mobile responsiveness with min-w-0 on flex children.',
  'curl -o .cursor/rules/tailwind.mdc https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/tailwind.mdc',
  '.cursor/rules/tailwind.mdc',
  'css',
  ARRAY['tailwind', 'css', 'ui', 'design']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Rust Axum & Tokio Async Architecture',
  'rust-axum-tokio-async-architecture',
  'Safe, concurrent HTTP services using Axum, Tokio, and SQLx with compile-time checked queries.',
  'cursor',
  'rules',
  'backend',
  'https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/rust-axum.mdc',
  'https://github.com/PatrickJS/awesome-cursorrules',
  'PatrickJS',
  'awesome-cursorrules',
  3150,
  210,
  110,
  '# Rust Axum Architecture
- Use State extractor with Arc<AppState>.
- Return Result<T, AppError> implementing IntoResponse.
- Prefer non-blocking I/O throughout.',
  'Use State extractor with Arc<AppState>. Return Result<T, AppError> implementing IntoResponse.',
  'curl -o .cursor/rules/rust-axum.mdc https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/rust-axum.mdc',
  '.cursor/rules/rust-axum.mdc',
  'rust',
  ARRAY['rust', 'axum', 'backend', 'tokio']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'PostgreSQL Context MCP Server',
  'postgresql-context-mcp-server',
  'Model Context Protocol server providing schema discovery, read-only querying, and query optimization for PostgreSQL databases.',
  'mcp',
  'mcp_servers',
  'backend',
  'https://github.com/modelcontextprotocol/servers/tree/main/src/postgres',
  'https://github.com/modelcontextprotocol/servers',
  'modelcontextprotocol',
  'servers',
  12500,
  1420,
  450,
  '# Postgres MCP Server
Connects Claude and other LLMs securely to PostgreSQL databases. Features schema inspection, EXPLAIN analysis, and parameterized queries.',
  'Connects Claude and other LLMs securely to PostgreSQL databases. Features schema inspection and EXPLAIN analysis.',
  'npx -y @modelcontextprotocol/server-postgres "postgresql://user:password@localhost:5432/mydb"',
  'src/postgres/index.ts',
  'typescript',
  ARRAY['mcp', 'postgres', 'database', 'sql']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Filesystem Workspace MCP Server',
  'filesystem-workspace-mcp-server',
  'Official MCP server for sandboxed local directory reads, edits, diffs, and workspace searching.',
  'mcp',
  'mcp_servers',
  'tools',
  'https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem',
  'https://github.com/modelcontextprotocol/servers',
  'modelcontextprotocol',
  'servers',
  12500,
  1420,
  380,
  '# Filesystem MCP Server
Provides secure, sandboxed file operations for LLM agent hosts like Claude Desktop.',
  'Provides secure, sandboxed file operations for LLM agent hosts like Claude Desktop.',
  'npx -y @modelcontextprotocol/server-filesystem /path/to/allowed/dir',
  'src/filesystem/index.ts',
  'typescript',
  ARRAY['mcp', 'filesystem', 'tools', 'sandboxed']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Playwright Browser Automation MCP Server',
  'playwright-browser-automation-mcp-server',
  'Headless browser control MCP server supporting web navigation, DOM inspection, clicking, and full-page screenshot auditing.',
  'mcp',
  'mcp_servers',
  'tools',
  'https://github.com/executeautomation/mcp-playwright',
  'https://github.com/executeautomation/mcp-playwright',
  'executeautomation',
  'mcp-playwright',
  1820,
  195,
  160,
  '# Playwright MCP Server
Enables AI agents to interact with web pages, fill forms, run end-to-end tests, and capture screenshots via Playwright.',
  'Enables AI agents to interact with web pages, fill forms, run end-to-end tests, and capture screenshots.',
  'npx -y @executeautomation/playwright-mcp-server',
  'index.ts',
  'typescript',
  ARRAY['mcp', 'playwright', 'browser', 'automation']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'GitHub API & Pull Request MCP Server',
  'github-api-pull-request-mcp-server',
  'Seamless GitHub repository management, issue creation, branch commits, and PR reviews via MCP.',
  'mcp',
  'mcp_servers',
  'devops',
  'https://github.com/modelcontextprotocol/servers/tree/main/src/github',
  'https://github.com/modelcontextprotocol/servers',
  'modelcontextprotocol',
  'servers',
  12500,
  1420,
  410,
  '# GitHub MCP Server
Allows agent hosts to query repositories, create commits, inspect branches, and open pull requests autonomously.',
  'Allows agent hosts to query repositories, create commits, inspect branches, and open pull requests autonomously.',
  'GITHUB_PERSONAL_ACCESS_TOKEN="ghp_xxx" npx -y @modelcontextprotocol/server-github',
  'src/github/index.ts',
  'typescript',
  ARRAY['mcp', 'github', 'git', 'devops']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Fullstack TDD Autonomous Agent',
  'fullstack-tdd-autonomous-agent',
  'Specification for autonomous coding agent following RED-GREEN-REFACTOR with automated regression test execution.',
  'claude',
  'skills',
  'ai-ml',
  'https://raw.githubusercontent.com/agentsmd/agents.md/main/examples/fullstack-tdd.md',
  'https://github.com/agentsmd/agents.md',
  'agentsmd',
  'agents.md',
  5600,
  430,
  290,
  '# Agent Persona: Fullstack TDD Specialist
- Always write unit tests before touching application code.
- Verify test failure (RED), implement minimum code (GREEN), then refactor.
- Never delete tests without justification.',
  'Always write unit tests before touching application code. Verify test failure (RED), implement code (GREEN), then refactor.',
  'curl -o AGENTS.md https://raw.githubusercontent.com/agentsmd/agents.md/main/examples/fullstack-tdd.md',
  'AGENTS.md',
  'markdown',
  ARRAY['claude', 'tdd', 'testing', 'agent-spec']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Security Code Auditor & Pentest Guard',
  'security-code-auditor-pentest-guard',
  'Automated security review skill identifying OWASP Top 10 vulnerabilities, CWE anti-patterns, and timing attacks.',
  'claude',
  'skills',
  'devops',
  'https://raw.githubusercontent.com/agentsmd/agents.md/main/examples/security-auditor.md',
  'https://github.com/agentsmd/agents.md',
  'agentsmd',
  'agents.md',
  4100,
  310,
  175,
  '# Security Auditor
- Inspect all database queries for SQL injection.
- Validate proper CSRF and CORS protections.
- Enforce secure cookie attributes: HttpOnly, Secure, SameSite.',
  'Inspect all database queries for SQL injection. Validate CSRF/CORS protections. Enforce secure cookie attributes.',
  'curl -o CLAUDE.md https://raw.githubusercontent.com/agentsmd/agents.md/main/examples/security-auditor.md',
  'CLAUDE.md',
  'markdown',
  ARRAY['claude', 'security', 'owasp', 'audit']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Cloud Run & Dockerfile Deployer Skill',
  'cloud-run-dockerfile-deployer-skill',
  'Agent workflow for optimizing multi-stage Dockerfiles and zero-downtime deployment to Google Cloud Run.',
  'gemini',
  'skills',
  'devops',
  'https://raw.githubusercontent.com/google-gemini/gemini-skills/main/cloud-run/SKILL.md',
  'https://github.com/google-gemini/gemini-skills',
  'google-gemini',
  'gemini-skills',
  2340,
  210,
  90,
  '# Cloud Run Deployment Skill
- Multi-stage build with unprivileged runner user.
- Environment parameter mapping via GCP Secret Manager.
- Automatic readiness probing on /readyz.',
  'Multi-stage build with unprivileged runner user. Environment parameter mapping via GCP Secret Manager.',
  'mkdir -p .gemini/skills && curl -o .gemini/skills/SKILL.md https://raw.githubusercontent.com/google-gemini/gemini-skills/main/cloud-run/SKILL.md',
  '.gemini/skills/SKILL.md',
  'dockerfile',
  ARRAY['gemini', 'docker', 'cloudrun', 'devops']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Gemini Multimodal Data Extraction Skill',
  'gemini-multimodal-data-extraction-skill',
  'Structured JSON schema extraction from invoices, receipts, and complex PDF tables using Gemini 1.5 Pro.',
  'gemini',
  'skills',
  'ai-ml',
  'https://raw.githubusercontent.com/google-gemini/gemini-skills/main/data-extraction/SKILL.md',
  'https://github.com/google-gemini/gemini-skills',
  'google-gemini',
  'gemini-skills',
  1890,
  140,
  80,
  '# Multimodal Extraction
- Define Pydantic output schema.
- Utilize temperature=0.0 for deterministic serialization.
- Validate returned fields with retry fallback.',
  'Define Pydantic output schema. Utilize temperature=0.0 for deterministic serialization.',
  'mkdir -p .gemini/skills && curl -o .gemini/skills/extraction.md https://raw.githubusercontent.com/google-gemini/gemini-skills/main/data-extraction/SKILL.md',
  '.gemini/skills/extraction.md',
  'python',
  ARRAY['gemini', 'multimodal', 'pdf', 'extraction']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'TypeScript Strict & Functional Core',
  'typescript-strict-functional-core',
  'GitHub Copilot instructions enforcing immutable data flow, Discriminated Unions, and zero ''any'' casts.',
  'copilot',
  'rules',
  'frontend',
  'https://raw.githubusercontent.com/github/copilot-instructions/main/typescript.md',
  'https://github.com/github/copilot-instructions',
  'github',
  'copilot-instructions',
  6700,
  520,
  210,
  '# Copilot Instructions: TypeScript
- Strict null checks enabled; never use ''any'' or non-null assertion ''!''.
- Prefer Discriminated Unions over type inheritance.
- Use readonly arrays and properties for immutability.',
  'Strict null checks enabled; never use any or non-null assertion !. Prefer Discriminated Unions.',
  'curl -o .github/copilot-instructions.md https://raw.githubusercontent.com/github/copilot-instructions/main/typescript.md',
  '.github/copilot-instructions.md',
  'typescript',
  ARRAY['copilot', 'typescript', 'functional', 'strict']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'ReAct Reasoning Framework Prompt',
  'react-reasoning-framework-prompt',
  'Standardized Thought-Action-Observation loop prompt template for complex multi-step tool use.',
  'generic',
  'frameworks',
  'ai-ml',
  'https://raw.githubusercontent.com/prompt-engineering/awesome-prompts/main/react-loop.md',
  'https://github.com/prompt-engineering/awesome-prompts',
  'prompt-engineering',
  'awesome-prompts',
  8900,
  910,
  310,
  '# ReAct Framework Prompt
Solve problems by alternating between Thought, Action, and Observation until the final deliverable is verified.',
  'Solve problems by alternating between Thought, Action, and Observation until final deliverable is verified.',
  'curl -o prompts/react-loop.md https://raw.githubusercontent.com/prompt-engineering/awesome-prompts/main/react-loop.md',
  'prompts/react-loop.md',
  'markdown',
  ARRAY['react-loop', 'prompts', 'frameworks', 'agents']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'SQLite Schema & Migration Tools Rule',
  'sqlite-schema-migration-tools-rule',
  'Best practices for embedded SQLite database access, WAL mode, pragmas, and schema evolution.',
  'cursor',
  'rules',
  'backend',
  'https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/sqlite.mdc',
  'https://github.com/PatrickJS/awesome-cursorrules',
  'PatrickJS',
  'awesome-cursorrules',
  2100,
  130,
  65,
  '# SQLite Guidelines
- Always enable PRAGMA journal_mode=WAL and PRAGMA busy_timeout=5000.
- Handle idempotent columns carefully with ALTER TABLE before CREATE INDEX.
- Store timestamps in UTC ISO8601 strings.',
  'Always enable PRAGMA journal_mode=WAL. Store timestamps in UTC ISO8601 strings.',
  'curl -o .cursor/rules/sqlite.mdc https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/sqlite.mdc',
  '.cursor/rules/sqlite.mdc',
  'sql',
  ARRAY['sqlite', 'sql', 'database', 'cursor']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Claude Code Reviewer & Quality Gate',
  'claude-code-reviewer-quality-gate',
  'Pre-commit and pull request automated code review bot configured with static analysis checks.',
  'claude',
  'tools',
  'devops',
  'https://raw.githubusercontent.com/agentsmd/agents.md/main/examples/code-reviewer.md',
  'https://github.com/agentsmd/agents.md',
  'agentsmd',
  'agents.md',
  3400,
  260,
  130,
  '# Code Reviewer
- Review changes for cyclomatic complexity, missing test coverage, and naming consistency.
- Provide actionable inline patch suggestions.',
  'Review changes for cyclomatic complexity, test coverage, and naming consistency. Provide actionable patch suggestions.',
  'curl -o .claude/review.md https://raw.githubusercontent.com/agentsmd/agents.md/main/examples/code-reviewer.md',
  '.claude/review.md',
  'markdown',
  ARRAY['claude', 'code-review', 'quality', 'tools']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Vue 3 & Pinia Composition API Rule',
  'vue-3-pinia-composition-api-rule',
  'Composition API standards, TypeScript props, Pinia state stores, and Vue Router guard patterns.',
  'cursor',
  'rules',
  'frontend',
  'https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/vue3.mdc',
  'https://github.com/PatrickJS/awesome-cursorrules',
  'PatrickJS',
  'awesome-cursorrules',
  1820,
  110,
  45,
  '# Vue 3 Rules
- Strictly use <script setup lang="ts">.
- Use defineProps and defineEmits with TypeScript types.
- Keep Pinia stores modular and strongly typed.',
  'Strictly use script setup lang=ts. Use defineProps and defineEmits with TypeScript types. Keep Pinia stores modular.',
  'curl -o .cursor/rules/vue3.mdc https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/vue3.mdc',
  '.cursor/rules/vue3.mdc',
  'vue',
  ARRAY['vue', 'pinia', 'typescript', 'frontend']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Django 5 & Ninja Async API Rule',
  'django-5-ninja-async-api-rule',
  'Clean Django 5 patterns using Django-Ninja async router, Pydantic schemas, and optimized QuerySets.',
  'cursor',
  'rules',
  'backend',
  'https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/django.mdc',
  'https://github.com/PatrickJS/awesome-cursorrules',
  'PatrickJS',
  'awesome-cursorrules',
  2150,
  140,
  55,
  '# Django Guidelines
- Use select_related and prefetch_related to eliminate N+1 queries.
- Use Django Ninja for OpenAPI-compliant async endpoints.',
  'Use select_related and prefetch_related to eliminate N+1 queries. Use Django Ninja for OpenAPI endpoints.',
  'curl -o .cursor/rules/django.mdc https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/django.mdc',
  '.cursor/rules/django.mdc',
  'python',
  ARRAY['django', 'python', 'backend', 'api']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Docker & Multi-Stage Production Containers',
  'docker-multi-stage-production-containers',
  'Minimal, non-root, alpine and scratch base images for secure container deployments.',
  'cursor',
  'rules',
  'devops',
  'https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/docker.mdc',
  'https://github.com/PatrickJS/awesome-cursorrules',
  'PatrickJS',
  'awesome-cursorrules',
  3900,
  320,
  120,
  '# Docker Standards
- Run as unprivileged non-root user.
- Use builder stages to isolate compile-time dependencies.
- Ensure .dockerignore excludes node_modules, git, and sensitive envs.',
  'Run as unprivileged non-root user. Use builder stages to isolate dependencies. Ensure .dockerignore excludes secret files.',
  'curl -o .cursor/rules/docker.mdc https://raw.githubusercontent.com/PatrickJS/awesome-cursorrules/main/rules/docker.mdc',
  '.cursor/rules/docker.mdc',
  'dockerfile',
  ARRAY['docker', 'devops', 'containers', 'security']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Brave Web Search MCP Server',
  'brave-web-search-mcp-server',
  'Real-time web search and local POI query engine for AI assistants powered by Brave Search API.',
  'mcp',
  'mcp_servers',
  'tools',
  'https://github.com/modelcontextprotocol/servers/tree/main/src/brave-search',
  'https://github.com/modelcontextprotocol/servers',
  'modelcontextprotocol',
  'servers',
  12500,
  1420,
  490,
  '# Brave Search MCP Server
Enables agent hosts to perform live web queries, fetch news, and verify facts with fresh web index citations.',
  'Enables agent hosts to perform live web queries, fetch news, and verify facts with fresh citations.',
  'BRAVE_API_KEY="BSA-xxx" npx -y @modelcontextprotocol/server-brave-search',
  'src/brave-search/index.ts',
  'typescript',
  ARRAY['mcp', 'search', 'web', 'brave']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Slack Workspace Operations MCP Server',
  'slack-workspace-operations-mcp-server',
  'Send notifications, triage channel messages, search conversation history, and post alerts to Slack.',
  'mcp',
  'mcp_servers',
  'tools',
  'https://github.com/modelcontextprotocol/servers/tree/main/src/slack',
  'https://github.com/modelcontextprotocol/servers',
  'modelcontextprotocol',
  'servers',
  12500,
  1420,
  340,
  '# Slack MCP Server
Integrates AI workflows directly into Slack channels for automated team notifications and conversational triage.',
  'Integrates AI workflows directly into Slack channels for automated team notifications.',
  'SLACK_BOT_TOKEN="xoxb-xxx" npx -y @modelcontextprotocol/server-slack',
  'src/slack/index.ts',
  'typescript',
  ARRAY['mcp', 'slack', 'collaboration', 'chat']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Git & Branch Version Control MCP Server',
  'git-branch-version-control-mcp-server',
  'Direct local git repository manipulation: diffs, stage, commit, branch switching, and log inspection.',
  'mcp',
  'mcp_servers',
  'devops',
  'https://github.com/modelcontextprotocol/servers/tree/main/src/git',
  'https://github.com/modelcontextprotocol/servers',
  'modelcontextprotocol',
  'servers',
  12500,
  1420,
  310,
  '# Git MCP Server
Execute git operations locally with strict sandbox enforcement and detailed unified diff output.',
  'Execute git operations locally with strict sandbox enforcement and detailed unified diff output.',
  'npx -y @modelcontextprotocol/server-git --repository /path/to/repo',
  'src/git/index.ts',
  'typescript',
  ARRAY['mcp', 'git', 'version-control', 'devops']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Autonomous Technical Writer AGENTS.md',
  'autonomous-technical-writer-agents-md',
  'Specialized agent configuration for generating clear, diagram-backed API documentation and tutorials.',
  'claude',
  'prompts',
  'frontend',
  'https://raw.githubusercontent.com/agentsmd/agents.md/main/examples/tech-writer.md',
  'https://github.com/agentsmd/agents.md',
  'agentsmd',
  'agents.md',
  2750,
  180,
  95,
  '# Tech Writer Agent
- Document every public endpoint with request/response JSON samples.
- Keep explanations concise, bulleted, and actionable.
- Use Mermaid.js or ASCII diagrams for architecture workflows.',
  'Document every public endpoint with request/response JSON samples. Keep explanations concise, bulleted, and actionable.',
  'curl -o AGENTS.md https://raw.githubusercontent.com/agentsmd/agents.md/main/examples/tech-writer.md',
  'AGENTS.md',
  'markdown',
  ARRAY['claude', 'docs', 'writing', 'documentation']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'System Architecture & Threat Modeling Prompt',
  'system-architecture-threat-modeling-prompt',
  'STRIDE threat modeling and architectural validation prompt for mission-critical cloud deployments.',
  'generic',
  'prompts',
  'devops',
  'https://raw.githubusercontent.com/prompt-engineering/awesome-prompts/main/threat-modeling.md',
  'https://github.com/prompt-engineering/awesome-prompts',
  'prompt-engineering',
  'awesome-prompts',
  5100,
  420,
  150,
  '# Threat Modeling Agent
Perform STRIDE evaluation on each trust boundary: Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege.',
  'Perform STRIDE evaluation on each trust boundary: Spoofing, Tampering, Repudiation, Info Disclosure, DoS, Elevation.',
  'curl -o prompts/threat-modeling.md https://raw.githubusercontent.com/prompt-engineering/awesome-prompts/main/threat-modeling.md',
  'prompts/threat-modeling.md',
  'markdown',
  ARRAY['prompts', 'security', 'stride', 'architecture']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'BigQuery SQL Query Optimization Skill',
  'bigquery-sql-query-optimization-skill',
  'Analyzes BigQuery query execution plans, partition pruning, and cluster keys to slash byte consumption.',
  'gemini',
  'skills',
  'backend',
  'https://raw.githubusercontent.com/google-gemini/gemini-skills/main/bigquery/SKILL.md',
  'https://github.com/google-gemini/gemini-skills',
  'google-gemini',
  'gemini-skills',
  1640,
  115,
  60,
  '# BigQuery Optimizer
- Avoid SELECT *; project only needed columns.
- Enforce partition filter requirements.
- Use approximate aggregation functions (APPROX_COUNT_DISTINCT) for large datasets.',
  'Avoid SELECT *; project only needed columns. Enforce partition filter requirements. Use approximate aggregation.',
  'mkdir -p .gemini/skills && curl -o .gemini/skills/bigquery.md https://raw.githubusercontent.com/google-gemini/gemini-skills/main/bigquery/SKILL.md',
  '.gemini/skills/bigquery.md',
  'sql',
  ARRAY['gemini', 'bigquery', 'sql', 'data']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'Go Chi & Clean Domain Model Copilot Rule',
  'go-chi-clean-domain-model-copilot-rule',
  'Copilot workspace configuration for idiomatic Go error handling, interfaces, and struct pointers.',
  'copilot',
  'rules',
  'backend',
  'https://raw.githubusercontent.com/github/copilot-instructions/main/go.md',
  'https://github.com/github/copilot-instructions',
  'github',
  'copilot-instructions',
  4800,
  390,
  170,
  '# Go Copilot Guidelines
- Accept interfaces, return structs.
- Wrap errors with fmt.Errorf("...: %w", err).
- Never use panic() in HTTP handlers.',
  'Accept interfaces, return structs. Wrap errors with %w. Never use panic() in HTTP handlers.',
  'curl -o .github/copilot-instructions.md https://raw.githubusercontent.com/github/copilot-instructions/main/go.md',
  '.github/copilot-instructions.md',
  'go',
  ARRAY['copilot', 'go', 'backend', 'idiomatic']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'LangGraph StateGraph Workflow Orchestrator',
  'langgraph-stategraph-workflow-orchestrator',
  'Cyclic multi-agent graph specification with conditional routing, checkpoints, and human-in-the-loop breakpoints.',
  'generic',
  'frameworks',
  'ai-ml',
  'https://github.com/langchain-ai/langgraph',
  'https://github.com/langchain-ai/langgraph',
  'langchain-ai',
  'langgraph',
  14200,
  1850,
  620,
  '# LangGraph Orchestration
- Define TypedDict state schema.
- Connect nodes via conditional edges.
- Persist state checkpointing using PostgresSaver.',
  'Define TypedDict state schema. Connect nodes via conditional edges. Persist state checkpointing.',
  'pip install -U langgraph langchain-core',
  'README.md',
  'python',
  ARRAY['langgraph', 'agents', 'frameworks', 'python']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();

INSERT INTO skills (
  name, slug, description, platform, category, subcategory,
  source_url, repo_url, repo_owner, repo_name,
  stars_count, forks_count, stars_velocity,
  content_raw, content_preview, install_snippet, file_path, language, tags,
  last_synced_at
) VALUES (
  'OpenAPI Spec to Agent Tool Generator',
  'openapi-spec-to-agent-tool-generator',
  'Automated tool schema generator converting Swagger 3.0 / OpenAPI specs into LLM function call definitions.',
  'generic',
  'tools',
  'tools',
  'https://github.com/open-agent-tools/openapi-generator',
  'https://github.com/open-agent-tools/openapi-generator',
  'open-agent-tools',
  'openapi-generator',
  3100,
  270,
  140,
  '# OpenAPI Tool Generator
Parses JSON/YAML OpenAPI specifications into standardized JSON Schema functions for Claude, OpenAI, and Gemini models.',
  'Parses JSON/YAML OpenAPI specifications into standardized JSON Schema functions for AI agents.',
  'npx -y openapi-agent-generator --spec swagger.json',
  'cli.ts',
  'typescript',
  ARRAY['openapi', 'tools', 'generator', 'json-schema']::TEXT[],
  NOW()
) ON CONFLICT (slug) DO UPDATE SET
  stars_count = EXCLUDED.stars_count,
  stars_velocity = EXCLUDED.stars_velocity,
  updated_at = NOW();
