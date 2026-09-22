package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"agentsight/internal/models"
	"agentsight/internal/service"
)

type SearchHandler struct {
	searchService service.SearchService
}

func NewSearchHandler(searchService service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

func (h *SearchHandler) APISearch(w http.ResponseWriter, r *http.Request) {
	params := h.parseSearchParams(r)
	result, err := h.searchService.Search(r.Context(), params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "search failed: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *SearchHandler) HandleSearchPage(w http.ResponseWriter, r *http.Request) {
	params := h.parseSearchParams(r)
	result, err := h.searchService.Search(r.Context(), params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "search failed: "+err.Error())
		return
	}

	if wantsJSON(r) {
		respondJSON(w, http.StatusOK, result)
		return
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Search Results: %s - AgentSight</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-950 text-slate-100 min-h-screen">
  <header class="border-b border-slate-800 px-6 py-4 flex items-center justify-between">
    <a href="/" class="font-bold text-xl text-violet-400">AgentSight</a>
    <a href="/trending" class="hover:text-violet-300">Trending</a>
  </header>
  <main class="max-w-5xl mx-auto px-6 py-12">
    <div class="mb-8">
      <form action="/search" method="GET" class="flex gap-2">
        <input type="text" name="q" value="%s" placeholder="Search rules, skills, MCP servers..." class="flex-1 bg-slate-900 border border-slate-700 px-4 py-3 rounded text-white focus:outline-none focus:border-violet-500">
        <button type="submit" class="bg-violet-600 hover:bg-violet-700 px-6 py-3 rounded font-medium">Search</button>
      </form>
    </div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold">Results for "%s"</h1>
      <span class="text-slate-400 text-sm">Found %d results (page %d of %d)</span>
    </div>
    <div class="space-y-4">`,
		params.Query, params.Query, params.Query, result.Total, result.Page, result.TotalPages,
	)

	for _, s := range result.Skills {
		html += fmt.Sprintf(`
      <div class="p-4 bg-slate-900 border border-slate-800 rounded flex justify-between items-center">
        <div>
          <a href="/skill/%s" class="text-lg font-bold text-violet-300 hover:underline">%s</a>
          <span class="ml-2 text-xs px-2 py-1 bg-slate-800 rounded uppercase tracking-wider text-cyan-300">%s</span>
          <span class="ml-1 text-xs px-2 py-1 bg-slate-800 rounded text-slate-400">%s</span>
          <p class="text-sm text-slate-400 mt-1">%s</p>
        </div>
        <div class="text-right">
          <span class="text-cyan-400 font-mono">★ %d</span>
        </div>
      </div>`, s.Slug, s.Name, s.Platform, s.Category, s.ContentPreview, s.StarsCount)
	}

	if len(result.Skills) == 0 {
		html += `
      <div class="p-8 text-center bg-slate-900 border border-slate-800 rounded text-slate-400">
        No skills found matching your search query. Try different keywords or filters.
      </div>`
	}

	html += `
    </div>
  </main>
</body>
</html>`

	respondHTML(w, http.StatusOK, html)
}

func (h *SearchHandler) parseSearchParams(r *http.Request) models.SearchParams {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	return models.SearchParams{
		Query:       q.Get("q"),
		Platform:    q.Get("platform"),
		Category:    q.Get("category"),
		Subcategory: q.Get("subcategory"),
		Language:    q.Get("language"),
		Tag:         q.Get("tag"),
		Sort:        q.Get("sort"),
		Page:        page,
		Limit:       limit,
	}
}
