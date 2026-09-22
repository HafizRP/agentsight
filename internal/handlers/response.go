package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error": message,
	})
}

func respondHTML(w http.ResponseWriter, status int, htmlContent string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(htmlContent))
}

func wantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/json") || strings.HasPrefix(r.URL.Path, "/api/")
}

func isHXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
