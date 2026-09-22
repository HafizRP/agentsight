# ROLE SPECIFICATION: FRONTEND ENGINEER (AGENT 3)
# TASK ID: TASK_3
# TARGET BRANCH: feat/frontend-ui
# DEPENDENCIES: [TASK_1]

## 1. OBJECTIVE
Build a responsive, interactive discovery UI for AgentSight using Go HTML templates, HTMX, Tailwind CSS, and minimal vanilla JS.

## 2. SCOPE OF IMPLEMENTATION

### Design System
- **Dark-first theme** with accent colors:
  - Background: Slate 900/950 (`#0f172a`, `#020617`)
  - Cards: Slate 800 (`#1e293b`) with subtle border (`border-slate-700`)
  - Primary accent: Violet 500 (`#8b5cf6`) — for CTAs, active states, highlights
  - Secondary: Cyan 400 (`#22d3ee`) — for star counts, badges
  - Text: Slate 50/200 (`#f8fafc`, `#e2e8f0`)
- **Platform badges** with distinct colors:
  - Cursor: Blue pill (`bg-blue-500/20 text-blue-400`)
  - Claude: Orange pill (`bg-orange-500/20 text-orange-400`)
  - Gemini: Emerald pill (`bg-emerald-500/20 text-emerald-400`)
  - MCP: Purple pill (`bg-purple-500/20 text-purple-400`)
  - Copilot: Sky pill (`bg-sky-500/20 text-sky-400`)
- **Typography**: Inter (Google Fonts) for body, JetBrains Mono for code snippets.
- Tailwind CSS via CDN. HTMX via CDN.

### Templates (`internal/templates/`)

**`base.html`** — Root layout:
- `<head>`: Tailwind CDN, HTMX CDN, Inter + JetBrains Mono fonts, custom styles, meta tags.
- Navbar: AgentSight logo/wordmark, search bar (expandable), nav links (Trending, Browse, Platforms), GitHub OAuth login button / user avatar dropdown.
- Footer: GitHub link, "Built for the agentic era" tagline.

**`index.html`** — Landing / Home:
- Hero section: Headline "Discover the best AI agent skills" + subtitle + search bar (large, centered).
- Stats row: Total skills indexed, platforms covered, sources monitored (animated counters).
- Trending Skills: Top 8 trending skill cards in a responsive grid.
- Browse by Platform: Horizontal row of platform filter pills (Cursor, Claude, Gemini, MCP, Copilot) → HTMX partial load.
- Browse by Category: Category cards (Rules, Skills, MCP Servers, Prompts, Frameworks, Tools) with icon and count.

**`search.html`** — Search Results:
- Search bar (pre-filled with query).
- Filter sidebar: Platform checkboxes, Category checkboxes, Language/Framework tags, Sort dropdown (Relevance, Stars, Newest, Trending).
- Results grid: Skill cards with HTMX infinite scroll (`hx-get="/search?page=2" hx-trigger="revealed"`).
- Results count and pagination info.

**`skill_detail.html`** — Skill Detail Page:
- Skill header: Name, platform badge, category badge, star/fork counts, last synced.
- Tabs (HTMX-swapped):
  - **Overview**: Description, tags, repo link, install snippet with copy button.
  - **Content**: Full raw content rendered in a code block with syntax highlighting (Prism.js or highlight.js via CDN).
  - **Related**: Grid of related skills (same category/tags).
- Bookmark button (HTMX `hx-post`, toggles heart icon).
- "Open on GitHub" external link button.

**`trending.html`** — Trending Board:
- Time filter: This Week / This Month (HTMX toggle).
- Leaderboard-style list: Rank badge, skill card, stars velocity sparkline indicator, platform badge.

**`category.html`** / **`platform.html`** — Browse Pages:
- Filter header with active filter highlighted.
- Skill card grid with HTMX pagination.

**`auth/login.html`** — Login prompt:
- "Sign in with GitHub" button (dark, GitHub-branded).
- Benefits list: "Bookmark skills", "Personalized feed", "Track new additions".

**`partials/`** — HTMX partial templates:
- `_skill_card.html`: Reusable skill card component (name, description preview, platform badge, category badge, stars, tags, bookmark icon).
- `_skill_list.html`: Grid of skill cards (used by search, trending, browse).
- `_search_results.html`: Search results with count header + skill list.
- `_bookmark_button.html`: Toggle state for bookmark heart.

### Static Assets (`static/`)
- `style.css`: Custom styles beyond Tailwind — code block theming, card hover effects, search bar animation, bookmark pulse, smooth transitions.
- `app.js`: Copy-to-clipboard for install snippets, search debounce, animated counter on homepage, keyboard shortcuts (/ to focus search).

### Template Renderer (`internal/templates/renderer.go`)
- Template parsing helper with template functions: `slugify`, `truncate`, `timeago`, `formatNumber`, `platformColor`, `categoryIcon`.

## 3. VERIFICATION CRITERIA
1. All templates parse without errors: `go build ./...`
2. No missing template references or undefined functions.
3. All HTMX interactions reference existing API endpoints from `contracts/api_contract.json`.

## 4. EXIT PROTOCOL
1. Commit changes to `feat/frontend-ui`.
2. Push branch and open PR via `gh pr create --title "feat(frontend): HTMX Discovery UI & Skill Browser" --body "..."`.
