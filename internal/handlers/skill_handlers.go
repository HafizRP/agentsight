package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"agentsight/internal/middleware"
	"agentsight/internal/models"
	"agentsight/internal/service"
	"agentsight/internal/templates"
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

	skills, err := h.trendingService.GetTrending(r.Context(), 8)
	if err != nil {
		skills = []models.Skill{}
	}
	stats, err := h.skillService.GetStats(r.Context())
	if err != nil {
		stats = &models.Stats{}
	}

	if isHXRequest(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]interface{}{
			"Skills":     skills,
			"Total":      len(skills),
			"Page":       1,
			"TotalPages": 1,
		}
		if err := templates.Default().RenderPartial(w, "_skill_list.html", data); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to render template: "+err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":     "Discover AI Agent Skills",
		"ActiveNav": "",
		"Skills":    skills,
		"Stats":     stats,
		"User":      middleware.UserFromContext(r.Context()),
	}

	if err := templates.Default().Render(w, "index.html", data); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to render landing page: "+err.Error())
		return
	}
}

func (h *SkillHandler) HandleSkillDetail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
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

	if wantsJSON(r) {
		respondJSON(w, http.StatusOK, detail)
		return
	}

	if isHXRequest(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.Default().RenderPartial(w, "_skill_card.html", detail.Skill); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to render template: "+err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":         detail.Skill.Name,
		"Skill":         detail.Skill,
		"IsBookmarked":  detail.IsBookmarked,
		"RelatedSkills": detail.RelatedSkills,
		"User":          middleware.UserFromContext(r.Context()),
	}

	if err := templates.Default().Render(w, "skill_detail.html", data); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to render skill detail: "+err.Error())
		return
	}
}

func (h *SkillHandler) HandleTrendingPage(w http.ResponseWriter, r *http.Request) {
	if wantsJSON(r) {
		h.APIGetTrending(w, r)
		return
	}

	timeframe := r.URL.Query().Get("timeframe")
	if timeframe == "" {
		timeframe = "week"
	}

	skills, err := h.trendingService.GetTrending(r.Context(), 50)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get trending skills: "+err.Error())
		return
	}

	if isHXRequest(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]interface{}{
			"Skills":    skills,
			"Total":     len(skills),
			"Timeframe": timeframe,
		}
		if err := templates.Default().RenderPartial(w, "_skill_list.html", data); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to render template: "+err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":     "Trending Skills",
		"ActiveNav": "trending",
		"Skills":    skills,
		"Timeframe": timeframe,
		"User":      middleware.UserFromContext(r.Context()),
	}

	if err := templates.Default().Render(w, "trending.html", data); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to render trending page: "+err.Error())
		return
	}
}

func (h *SkillHandler) HandleCategoryPage(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit := 20

	skills, total, err := h.skillService.ListByCategory(r.Context(), category, page, limit)
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

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	if isHXRequest(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]interface{}{
			"Skills":      skills,
			"Total":       total,
			"Page":        page,
			"TotalPages":  totalPages,
			"Category":    category,
			"InfiniteURL": fmt.Sprintf("/category/%s?page=%d", category, page+1),
		}
		if err := templates.Default().RenderPartial(w, "_skill_list.html", data); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to render template: "+err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":      category + " Skills",
		"Category":   category,
		"Skills":     skills,
		"Total":      total,
		"Page":       page,
		"TotalPages": totalPages,
		"User":       middleware.UserFromContext(r.Context()),
	}

	if err := templates.Default().Render(w, "category.html", data); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to render category page: "+err.Error())
		return
	}
}

func (h *SkillHandler) HandlePlatformPage(w http.ResponseWriter, r *http.Request) {
	platform := chi.URLParam(r, "platform")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit := 20

	skills, total, err := h.skillService.ListByPlatform(r.Context(), platform, page, limit)
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

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	if isHXRequest(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]interface{}{
			"Skills":      skills,
			"Total":       total,
			"Page":        page,
			"TotalPages":  totalPages,
			"Platform":    platform,
			"InfiniteURL": fmt.Sprintf("/platform/%s?page=%d", platform, page+1),
		}
		if err := templates.Default().RenderPartial(w, "_skill_list.html", data); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to render template: "+err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"Title":      platform + " Skills",
		"Platform":   platform,
		"Skills":     skills,
		"Total":      total,
		"Page":       page,
		"TotalPages": totalPages,
		"User":       middleware.UserFromContext(r.Context()),
	}

	if err := templates.Default().Render(w, "platform.html", data); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to render platform page: "+err.Error())
		return
	}
}
