package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
)

//go:embed *.html partials/*.html auth/*.html
var TemplateFS embed.FS

var (
	defaultRenderer *Renderer
	rendererOnce    sync.Once
)

// Renderer manages and renders HTML templates.
type Renderer struct {
	pages    map[string]*template.Template
	partials *template.Template
	funcMap  template.FuncMap
	mu       sync.RWMutex
}

// Default returns the package-level singleton renderer.
func Default() *Renderer {
	rendererOnce.Do(func() {
		r, err := NewRenderer()
		if err != nil {
			panic(fmt.Sprintf("failed to initialize default template renderer: %v", err))
		}
		defaultRenderer = r
	})
	return defaultRenderer
}

// NewRenderer initializes a new template renderer with all templates pre-parsed.
func NewRenderer() (*Renderer, error) {
	r := &Renderer{
		pages: make(map[string]*template.Template),
	}
	r.funcMap = r.FuncMap()

	if err := r.parseTemplates(); err != nil {
		return nil, err
	}
	return r, nil
}

// FuncMap returns the helper functions available to all templates.
func (r *Renderer) FuncMap() template.FuncMap {
	return template.FuncMap{
		"slugify":       Slugify,
		"truncate":      Truncate,
		"timeago":       TimeAgo,
		"formatNumber":  FormatNumber,
		"platformColor": PlatformColor,
		"categoryIcon":  CategoryIcon,
		"dict":          Dict,
		"add":           Add,
		"sub":           Sub,
		"safeHTML":      SafeHTML,
		"hasTag":        HasTag,
	}
}

func (r *Renderer) parseTemplates() error {
	pageFiles := []string{
		"index.html",
		"search.html",
		"skill_detail.html",
		"trending.html",
		"category.html",
		"platform.html",
		"auth/login.html",
	}

	partialFiles := []string{
		"partials/_bookmark_button.html",
		"partials/_skill_card.html",
		"partials/_skill_list.html",
		"partials/_search_results.html",
	}

	// 1. Parse shared partials template
	partialsTmpl := template.New("partials").Funcs(r.funcMap)
	parsedPartials, err := partialsTmpl.ParseFS(TemplateFS, partialFiles...)
	if err != nil {
		return fmt.Errorf("failed to parse partials: %w", err)
	}
	r.partials = parsedPartials

	// 2. Parse pages with base.html and all partials
	for _, page := range pageFiles {
		filesToParse := append([]string{"base.html", page}, partialFiles...)
		tmpl, err := template.New("base.html").Funcs(r.funcMap).ParseFS(TemplateFS, filesToParse...)
		if err != nil {
			return fmt.Errorf("failed to parse page template %s: %w", page, err)
		}
		r.pages[page] = tmpl

		// Also register short name without directory (e.g. "login.html" for "auth/login.html")
		if strings.Contains(page, "/") {
			short := page[strings.LastIndex(page, "/")+1:]
			r.pages[short] = tmpl
		}
	}

	return nil
}

// Render writes the rendered full page (inside base.html) to w.
func (r *Renderer) Render(w io.Writer, page string, data interface{}) error {
	r.mu.RLock()
	tmpl, ok := r.pages[page]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("template page %q not found", page)
	}

	return tmpl.ExecuteTemplate(w, "base.html", data)
}

// RenderPartial writes a rendered partial template directly to w.
func (r *Renderer) RenderPartial(w io.Writer, name string, data interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.partials == nil {
		return fmt.Errorf("no partials template initialized")
	}

	// Try lookup by candidate names, prioritizing defined template name
	clean := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(name, "partials/"), "/"), "_"), ".html")
	candidates := []string{
		clean,
		name,
		strings.TrimSuffix(strings.TrimPrefix(name, "partials/"), ".html"),
		strings.TrimPrefix(name, "partials/"),
	}

	for _, cand := range candidates {
		if r.partials.Lookup(cand) != nil {
			return r.partials.ExecuteTemplate(w, cand, data)
		}
	}

	return fmt.Errorf("partial template %q not found", name)
}

// --- Helper Functions ---

// Slugify converts text into a clean URL-friendly slug.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// Truncate shortens a string to maxLen runes and appends "..." if truncated.
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// TimeAgo formats a timestamp relative to now (e.g., "5m ago", "2d ago").
func TimeAgo(v interface{}) string {
	if v == nil {
		return "never"
	}
	var t time.Time
	switch val := v.(type) {
	case time.Time:
		t = val
	case *time.Time:
		if val == nil {
			return "never"
		}
		t = *val
	case string:
		if val == "" {
			return "never"
		}
		parsed, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return val
		}
		t = parsed
	default:
		return "never"
	}

	if t.IsZero() {
		return "never"
	}

	d := time.Since(t)
	if d < 0 {
		return "just now"
	}
	seconds := int(d.Seconds())
	if seconds < 60 {
		return "just now"
	}
	minutes := int(d.Minutes())
	if minutes < 60 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	hours := int(d.Hours())
	if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	}
	days := hours / 24
	if days < 7 {
		return fmt.Sprintf("%dd ago", days)
	}
	weeks := days / 7
	if weeks < 4 {
		return fmt.Sprintf("%dw ago", weeks)
	}
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%dmo ago", months)
	}
	years := days / 365
	return fmt.Sprintf("%dy ago", years)
}

// FormatNumber formats integers into clean compact forms (e.g. 1.2k, 1M).
func FormatNumber(n int) string {
	if n < 0 {
		return fmt.Sprintf("%d", n)
	}
	if n >= 1000000 {
		val := float64(n) / 1000000.0
		if val == math.Floor(val) {
			return fmt.Sprintf("%.0fM", val)
		}
		return fmt.Sprintf("%.1fM", val)
	}
	if n >= 1000 {
		val := float64(n) / 1000.0
		if val == math.Floor(val) {
			return fmt.Sprintf("%.0fk", val)
		}
		return fmt.Sprintf("%.1fk", val)
	}
	return fmt.Sprintf("%d", n)
}

// PlatformColor returns the tailwind classes for platform badges.
func PlatformColor(v interface{}) string {
	platform := fmt.Sprintf("%v", v)
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "cursor":
		return "bg-blue-500/20 text-blue-400 border border-blue-500/30"
	case "claude":
		return "bg-orange-500/20 text-orange-400 border border-orange-500/30"
	case "gemini":
		return "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30"
	case "mcp":
		return "bg-purple-500/20 text-purple-400 border border-purple-500/30"
	case "copilot":
		return "bg-sky-500/20 text-sky-400 border border-sky-500/30"
	default:
		return "bg-slate-500/20 text-slate-400 border border-slate-500/30"
	}
}

// CategoryIcon returns inline SVG icon markup for categories.
func CategoryIcon(v interface{}) template.HTML {
	category := fmt.Sprintf("%v", v)
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "rules":
		return template.HTML(`<svg class="w-4 h-4 inline-block" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>`)
	case "skills":
		return template.HTML(`<svg class="w-4 h-4 inline-block" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>`)
	case "mcp_servers":
		return template.HTML(`<svg class="w-4 h-4 inline-block" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"></path></svg>`)
	case "prompts":
		return template.HTML(`<svg class="w-4 h-4 inline-block" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"></path></svg>`)
	case "frameworks":
		return template.HTML(`<svg class="w-4 h-4 inline-block" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>`)
	case "tools":
		return template.HTML(`<svg class="w-4 h-4 inline-block" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>`)
	default:
		return template.HTML(`<svg class="w-4 h-4 inline-block" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"></path></svg>`)
	}
}

// Dict creates a map from key-value pairs inside templates.
func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires even number of arguments")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict key must be a string")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

// Add sums two integers in templates.
func Add(a, b int) int {
	return a + b
}

// Sub subtracts b from a in templates.
func Sub(a, b int) int {
	return a - b
}

// SafeHTML marks a string as safe HTML.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// HasTag checks if a string slice contains a given tag (case-insensitive).
func HasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}
