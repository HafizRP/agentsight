package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"agentsight/internal/middleware"
	"agentsight/internal/models"
	"agentsight/internal/service"
	"github.com/go-chi/chi/v5"
)

type SkillHandler struct {
	skillService    service.SkillService
	trendingService service.TrendingService
}

func NewSkillHandler(skillService service.SkillService, trendingService service.TrendingService) *SkillHandler {
	return &SkillHandler{
		skillService:    skillService,
		trendingService: trendingService,
	}
}

func (h *SkillHandler) APIListSkills(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := models.SkillFilter{
		Platform:    q.Get("platform"),
		Category:    q.Get("category"),
		Subcategory: q.Get("subcategory"),
		Language:    q.Get("language"),
		Tag:         q.Get("tag"),
		Sort:        q.Get("sort"),
		Page:        page,
		Limit:       limit,
	}

	skills, total, err := h.skillService.List(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list skills: "+err.Error())
		return
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + filter.Limit - 1) / filter.Limit
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"skills":      skills,
		"total":       total,
		"page":        filter.Page,
		"limit":       filter.Limit,
		"total_pages": totalPages,
	})
}

func (h *SkillHandler) APIGetSkill(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondError(w, http.StatusBadRequest, "slug is required")
		return
	}

	var userID int
	if u := middleware.UserFromContext(r.Context()); u != nil {
		userID = u.ID
	}

	detail, err := h.skillService.GetBySlug(r.Context(), slug, userID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			respondError(w, http.StatusNotFound, "skill not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get skill: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, detail)
}

func (h *SkillHandler) APIGetTrending(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	skills, err := h.trendingService.GetTrending(r.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get trending skills: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"skills": skills,
		"total":  len(skills),
	})
}

func (h *SkillHandler) APIGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.skillService.GetStats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get stats: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

// Web Page Handlers
func (h *SkillHandler) HandleLanding(w http.ResponseWriter, r *http.Request) {
	if wantsJSON(r) {
		h.APIGetTrending(w, r)
		return
	}

	skills, _ := h.trendingService.GetTrending(r.Context(), 8)
	stats, _ := h.skillService.GetStats(r.Context())

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>AgentSight - AI Agent Skill Discovery Engine</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-950 text-slate-100 min-h-screen">
  <header class="border-b border-slate-800 px-6 py-4 flex items-center justify-between">
    <div class="font-bold text-xl text-violet-400">AgentSight</div>
    <nav class="space-x-4">
      <a href="/trending" class="hover:text-violet-300">Trending</a>
      <a href="/search" class="hover:text-violet-300">Search</a>
      <a href="/auth/github" class="bg-violet-600 hover:bg-violet-700 px-4 py-2 rounded font-medium">Sign in with GitHub</a>
    </nav>
  </header>
  <main class="max-w-5xl mx-auto px-6 py-12">
    <h1 class="text-4xl font-bold mb-4">Discover the best AI agent skills</h1>
    <p class="text-slate-400 mb-8">Aggregating and ranking cursor rules, Claude AGENTS.md, Gemini skills, and MCP servers.</p>
    <div class="mb-10">
      <form action="/search" method="GET" class="flex gap-2">
        <input type="text" name="q" placeholder="Search rules, skills, MCP servers..." class="flex-1 bg-slate-900 border border-slate-700 px-4 py-3 rounded text-white focus:outline-none focus:border-violet-500">
        <button type="submit" class="bg-violet-600 hover:bg-violet-700 px-6 py-3 rounded font-medium">Search</button>
      </form>
    </div>
    <div class="grid grid-cols-3 gap-4 mb-10 text-center">
      <div class="bg-slate-900 p-4 rounded border border-slate-800">
        <div class="text-2xl font-bold text-cyan-400">%d</div>
        <div class="text-sm text-slate-400">Skills Indexed</div>
      </div>
      <div class="bg-slate-900 p-4 rounded border border-slate-800">
        <div class="text-2xl font-bold text-violet-400">%d</div>
        <div class="text-sm text-slate-400">Total Stars</div>
      </div>
      <div class="bg-slate-900 p-4 rounded border border-slate-800">
        <div class="text-2xl font-bold text-emerald-400">%d</div>
        <div class="text-sm text-slate-400">Sources Monitored</div>
      </div>
    </div>
    <h2 class="text-2xl font-semibold mb-4">Trending Skills (%d)</h2>
    <div class="space-y-4">`,
		safeInt(stats, func(s *models.Stats) int { return s.TotalSkills }),
		safeInt(stats, func(s *models.Stats) int { return s.TotalStars }),
		safeInt(stats, func(s *models.Stats) int { return s.TotalSources }),
		len(skills),
	)

	for _, s := range skills {
		html += fmt.Sprintf(`
      <div class="p-4 bg-slate-900 border border-slate-800 rounded flex justify-between items-center">
        <div>
          <a href="/skill/%s" class="text-lg font-bold text-violet-300 hover:underline">%s</a>
          <span class="ml-2 text-xs px-2 py-1 bg-slate-800 rounded uppercase tracking-wider text-cyan-300">%s</span>
          <p class="text-sm text-slate-400 mt-1">%s</p>
        </div>
        <div class="text-right">
          <span class="text-cyan-400 font-mono">★ %d</span>
          <div class="text-xs text-slate-500">velocity +%d/wk</div>
        </div>
      </div>`, s.Slug, s.Name, s.Platform, s.ContentPreview, s.StarsCount, s.StarsVelocity)
	}

	html += `
    </div>
  </main>
</body>
</html>`

	respondHTML(w, http.StatusOK, html)
}

func (h *SkillHandler) HandleSkillDetail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var userID int
	if u := middleware.UserFromContext(r.Context()); u != nil {
		userID = u.ID
	}

	detail, err := h.skillService.GetBySlug(r.Context(), slug, userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "skill not found")
		return
	}

	if wantsJSON(r) {
		respondJSON(w, http.StatusOK, detail)
		return
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>%s - AgentSight</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-950 text-slate-100 min-h-screen">
  <header class="border-b border-slate-800 px-6 py-4 flex items-center justify-between">
    <a href="/" class="font-bold text-xl text-violet-400">AgentSight</a>
    <a href="/search" class="hover:text-violet-300">Search</a>
  </header>
  <main class="max-w-4xl mx-auto px-6 py-12">
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-3xl font-bold">%s</h1>
      <span class="text-cyan-400 font-mono text-xl">★ %d</span>
    </div>
    <div class="flex gap-2 mb-6">
      <span class="text-xs px-2 py-1 bg-violet-900/40 text-violet-300 rounded">%s</span>
      <span class="text-xs px-2 py-1 bg-slate-800 text-slate-300 rounded">%s</span>
    </div>
    <p class="text-slate-300 text-lg mb-6">%s</p>
    <div class="bg-slate-900 border border-slate-800 p-4 rounded mb-6">
      <h3 class="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-2">Install Snippet</h3>
      <pre class="bg-slate-950 p-3 rounded text-sm text-cyan-300 font-mono overflow-x-auto">%s</pre>
    </div>
    <div class="bg-slate-900 border border-slate-800 p-4 rounded mb-8">
      <h3 class="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-2">Content Raw</h3>
      <pre class="bg-slate-950 p-3 rounded text-xs text-slate-300 font-mono overflow-x-auto whitespace-pre-wrap">%s</pre>
    </div>
    <a href="/" class="text-violet-400 hover:underline">← Back to discovery</a>
  </main>
</body>
</html>`,
		detail.Skill.Name, detail.Skill.Name, detail.Skill.StarsCount,
		detail.Skill.Platform, detail.Skill.Category, detail.Skill.Description,
		detail.Skill.InstallSnippet, detail.Skill.ContentRaw,
	)

	respondHTML(w, http.StatusOK, html)
}

func (h *SkillHandler) HandleTrendingPage(w http.ResponseWriter, r *http.Request) {
	if wantsJSON(r) {
		h.APIGetTrending(w, r)
		return
	}

	skills, err := h.trendingService.GetTrending(r.Context(), 50)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get trending skills")
		return
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Trending Skills - AgentSight</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-950 text-slate-100 min-h-screen">
  <header class="border-b border-slate-800 px-6 py-4 flex items-center justify-between">
    <a href="/" class="font-bold text-xl text-violet-400">AgentSight</a>
    <a href="/search" class="hover:text-violet-300">Search</a>
  </header>
  <main class="max-w-4xl mx-auto px-6 py-12">
    <h1 class="text-3xl font-bold mb-6">Trending Skills Board</h1>
    <div class="space-y-4">`)

	for i, s := range skills {
		html += fmt.Sprintf(`
      <div class="p-4 bg-slate-900 border border-slate-800 rounded flex justify-between items-center">
        <div class="flex items-center gap-4">
          <span class="text-slate-500 font-mono text-lg w-6">#%d</span>
          <div>
            <a href="/skill/%s" class="text-lg font-bold text-violet-300 hover:underline">%s</a>
            <span class="ml-2 text-xs px-2 py-1 bg-slate-800 rounded uppercase tracking-wider text-cyan-300">%s</span>
            <p class="text-sm text-slate-400 mt-1">%s</p>
          </div>
        </div>
        <div class="text-right">
          <span class="text-cyan-400 font-mono">★ %d</span>
          <div class="text-xs text-slate-500">velocity +%d/wk</div>
        </div>
      </div>`, i+1, s.Slug, s.Name, s.Platform, s.ContentPreview, s.StarsCount, s.StarsVelocity)
	}

	html += `
    </div>
  </main>
</body>
</html>`

	respondHTML(w, http.StatusOK, html)
}

func (h *SkillHandler) HandleCategoryPage(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	skills, total, err := h.skillService.ListByCategory(r.Context(), category, 1, 20)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to browse category: "+err.Error())
		return
	}

	if wantsJSON(r) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"category": category,
			"skills":   skills,
			"total":    total,
		})
		return
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Category: %s - AgentSight</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-950 text-slate-100 min-h-screen">
  <header class="border-b border-slate-800 px-6 py-4 flex items-center justify-between">
    <a href="/" class="font-bold text-xl text-violet-400">AgentSight</a>
    <a href="/search" class="hover:text-violet-300">Search</a>
  </header>
  <main class="max-w-4xl mx-auto px-6 py-12">
    <h1 class="text-3xl font-bold mb-2">Category: %s</h1>
    <p class="text-slate-400 mb-6">Found %d skills</p>
    <div class="space-y-4">`, category, category, total)

	for _, s := range skills {
		html += fmt.Sprintf(`
      <div class="p-4 bg-slate-900 border border-slate-800 rounded flex justify-between items-center">
        <div>
          <a href="/skill/%s" class="text-lg font-bold text-violet-300 hover:underline">%s</a>
          <span class="ml-2 text-xs px-2 py-1 bg-slate-800 rounded uppercase tracking-wider text-cyan-300">%s</span>
          <p class="text-sm text-slate-400 mt-1">%s</p>
        </div>
        <div class="text-right">
          <span class="text-cyan-400 font-mono">★ %d</span>
        </div>
      </div>`, s.Slug, s.Name, s.Platform, s.ContentPreview, s.StarsCount)
	}

	html += `
    </div>
  </main>
</body>
</html>`

	respondHTML(w, http.StatusOK, html)
}

func (h *SkillHandler) HandlePlatformPage(w http.ResponseWriter, r *http.Request) {
	platform := chi.URLParam(r, "platform")
	skills, total, err := h.skillService.ListByPlatform(r.Context(), platform, 1, 20)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to browse platform: "+err.Error())
		return
	}

	if wantsJSON(r) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"platform": platform,
			"skills":   skills,
			"total":    total,
		})
		return
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Platform: %s - AgentSight</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-950 text-slate-100 min-h-screen">
  <header class="border-b border-slate-800 px-6 py-4 flex items-center justify-between">
    <a href="/" class="font-bold text-xl text-violet-400">AgentSight</a>
    <a href="/search" class="hover:text-violet-300">Search</a>
  </header>
  <main class="max-w-4xl mx-auto px-6 py-12">
    <h1 class="text-3xl font-bold mb-2">Platform: %s</h1>
    <p class="text-slate-400 mb-6">Found %d skills</p>
    <div class="space-y-4">`, platform, platform, total)

	for _, s := range skills {
		html += fmt.Sprintf(`
      <div class="p-4 bg-slate-900 border border-slate-800 rounded flex justify-between items-center">
        <div>
          <a href="/skill/%s" class="text-lg font-bold text-violet-300 hover:underline">%s</a>
          <span class="ml-2 text-xs px-2 py-1 bg-slate-800 rounded uppercase tracking-wider text-cyan-300">%s</span>
          <p class="text-sm text-slate-400 mt-1">%s</p>
        </div>
        <div class="text-right">
          <span class="text-cyan-400 font-mono">★ %d</span>
        </div>
      </div>`, s.Slug, s.Name, s.Category, s.ContentPreview, s.StarsCount)
	}

	html += `
    </div>
  </main>
</body>
</html>`

	respondHTML(w, http.StatusOK, html)
}

func safeInt(s *models.Stats, getter func(s *models.Stats) int) int {
	if s == nil {
		return 0
	}
	return getter(s)
}
