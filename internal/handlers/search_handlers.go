package handlers

import (
	"net/http"
	"strconv"

	"agentsight/internal/middleware"
	"agentsight/internal/models"
	"agentsight/internal/service"
	"agentsight/internal/templates"
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

	if isHXRequest(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]interface{}{
			"Skills":     result.Skills,
			"Total":      result.Total,
			"Page":       result.Page,
			"Limit":      result.Limit,
			"TotalPages": result.TotalPages,
			"Query":      params.Query,
			"Platform":   params.Platform,
			"Category":   params.Category,
			"Sort":       params.Sort,
		}
		if err := templates.Default().RenderPartial(w, "_search_results.html", data); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to render search results: "+err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":     "Search",
		"ActiveNav": "browse",
		"Query":     params.Query,
		"Params":    params,
		"Result":    result,
		"User":      middleware.UserFromContext(r.Context()),
	}

	if err := templates.Default().Render(w, "search.html", data); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to render search page: "+err.Error())
		return
	}
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
