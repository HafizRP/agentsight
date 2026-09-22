package handlers

import (
	"net/http"
	"strconv"

	"agentsight/internal/middleware"
	"agentsight/internal/repository"
	"agentsight/internal/templates"
	"github.com/go-chi/chi/v5"
)

type BookmarkHandler struct {
	bookmarkRepo repository.BookmarkRepository
}

func NewBookmarkHandler(bookmarkRepo repository.BookmarkRepository) *BookmarkHandler {
	return &BookmarkHandler{bookmarkRepo: bookmarkRepo}
}

func (h *BookmarkHandler) APIToggleBookmark(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	skillIDStr := chi.URLParam(r, "skill_id")
	skillID, err := strconv.Atoi(skillIDStr)
	if err != nil || skillID <= 0 {
		respondError(w, http.StatusBadRequest, "invalid skill_id")
		return
	}

	bookmarked, err := h.bookmarkRepo.Toggle(r.Context(), user.ID, skillID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to toggle bookmark: "+err.Error())
		return
	}

	if isHXRequest(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]interface{}{
			"SkillID":    skillID,
			"Bookmarked": bookmarked,
		}
		if err := templates.Default().RenderPartial(w, "_bookmark_button.html", data); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to render bookmark button: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"skill_id":   skillID,
		"bookmarked": bookmarked,
	})
}

func (h *BookmarkHandler) APIListBookmarks(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	skills, total, err := h.bookmarkRepo.ListByUser(r.Context(), user.ID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list bookmarks: "+err.Error())
		return
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"skills":      skills,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}
