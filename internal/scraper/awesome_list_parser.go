package scraper

import (
	"regexp"
	"strings"
)

var (
	// Markdown heading regex: matches ## Category and ### Subcategory
	h2Regex = regexp.MustCompile(`^##\s+(?:[^\w\s]*\s*)?([^\n\r#]+)`)
	h3Regex = regexp.MustCompile(`^###\s+(?:[^\w\s]*\s*)?([^\n\r#]+)`)

	// Bullet list link regex: - [Name](URL) or * [Name](URL) or 1. [Name](URL)
	bulletLinkRegex = regexp.MustCompile(`^\s*(?:[-*+]|\d+\.)\s+(?:\*\*)?\[([^\]]+)\]\(([^)]+)\)(?:\*\*)?(.*)$`)

	// Link inside table cell: [Name](URL)
	cellLinkRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	// Description prefixes to clean
	descPrefixRegex = regexp.MustCompile(`^[\s:\-–—\s*]+`)

	// GitHub repo URL matcher
	githubRepoRegex = regexp.MustCompile(`^https?://(?:www\.)?github\.com/([^/]+)/([^/#?]+)`)
)

// AwesomeEntry represents a single item parsed from an awesome list.
type AwesomeEntry struct {
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	RawLine     string `json:"raw_line"`
}

// AwesomeListParser parses Markdown awesome lists, tables, and curated registries.
type AwesomeListParser struct{}

// NewAwesomeListParser creates a new AwesomeListParser instance.
func NewAwesomeListParser() *AwesomeListParser {
	return &AwesomeListParser{}
}

// Parse extracts all entries from a markdown string.
func (p *AwesomeListParser) Parse(content string) []AwesomeEntry {
	lines := strings.Split(content, "\n")
	var entries []AwesomeEntry

	currentCategory := "General"
	currentSubcategory := "General"

	inTable := false
	tableColName := -1
	tableColDesc := -1
	tableColLink := -1

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		// Check for Category heading (##)
		if matches := h2Regex.FindStringSubmatch(line); len(matches) > 1 {
			inTable = false
			cat := strings.TrimSpace(matches[1])
			if !strings.EqualFold(cat, "Contents") && !strings.EqualFold(cat, "Table of Contents") && !strings.EqualFold(cat, "Contributing") && !strings.EqualFold(cat, "License") {
				currentCategory = cleanHeading(cat)
				currentSubcategory = currentCategory
			}
			continue
		}

		// Check for Subcategory heading (###)
		if matches := h3Regex.FindStringSubmatch(line); len(matches) > 1 {
			inTable = false
			subcat := strings.TrimSpace(matches[1])
			if !strings.EqualFold(subcat, "Contents") && !strings.EqualFold(subcat, "Table of Contents") {
				currentSubcategory = cleanHeading(subcat)
			}
			continue
		}

		// Check for Markdown Table Header or Row
		if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
			cols := splitTableCells(line)
			if len(cols) >= 2 {
				if isTableDivider(cols) {
					inTable = true
					continue
				}

				if !inTable {
					tableColName, tableColDesc, tableColLink = findTableColumns(cols)
					continue
				}

				// If in table, parse data row
				entry := parseTableRow(cols, tableColName, tableColDesc, tableColLink, currentCategory, currentSubcategory, line)
				if entry.Name != "" && entry.URL != "" {
					entries = append(entries, entry)
					continue
				}
			}
		} else {
			inTable = false
		}

		// Check for Bullet List Item
		if matches := bulletLinkRegex.FindStringSubmatch(line); len(matches) > 3 {
			name := strings.TrimSpace(matches[1])
			url := strings.TrimSpace(matches[2])
			desc := strings.TrimSpace(matches[3])

			desc = descPrefixRegex.ReplaceAllString(desc, "")
			desc = strings.TrimSpace(desc)

			// Skip navigation anchors like #contents, #installation
			if strings.HasPrefix(url, "#") {
				continue
			}

			// Clean bold/italics from name
			name = strings.Trim(name, "*_`")

			entries = append(entries, AwesomeEntry{
				Category:    currentCategory,
				Subcategory: currentSubcategory,
				Name:        name,
				URL:         url,
				Description: desc,
				RawLine:     line,
			})
		}
	}

	return entries
}

// ParseToRawSkills converts markdown content into standardized RawSkill structs.
func (p *AwesomeListParser) ParseToRawSkills(content string, defaultOwner, defaultRepo string, stars, forks int) []RawSkill {
	entries := p.Parse(content)
	skills := make([]RawSkill, 0, len(entries))

	for _, e := range entries {
		owner := defaultOwner
		repo := defaultRepo
		repoURL := ""

		if ghMatches := githubRepoRegex.FindStringSubmatch(e.URL); len(ghMatches) > 2 {
			owner = ghMatches[1]
			repo = ghMatches[2]
			repo = strings.TrimSuffix(repo, ".git")
			repoURL = "https://github.com/" + owner + "/" + repo
		} else if defaultOwner != "" && defaultRepo != "" {
			repoURL = "https://github.com/" + defaultOwner + "/" + defaultRepo
		}

		skills = append(skills, RawSkill{
			Name:           e.Name,
			Description:    e.Description,
			SourceURL:      e.URL,
			RepoURL:        repoURL,
			RepoOwner:      owner,
			RepoName:       repo,
			StarsCount:     stars,
			ForksCount:     forks,
			Category:       e.Category,
			Subcategory:    e.Subcategory,
			ContentPreview: e.Description,
		})
	}

	return skills
}

func cleanHeading(heading string) string {
	heading = strings.TrimSpace(heading)
	heading = strings.Trim(heading, "#*`_:")
	return strings.TrimSpace(heading)
}

func splitTableCells(line string) []string {
	trimmed := strings.Trim(line, "|")
	parts := strings.Split(trimmed, "|")
	res := make([]string, len(parts))
	for i, p := range parts {
		res[i] = strings.TrimSpace(p)
	}
	return res
}

func isTableDivider(cols []string) bool {
	for _, c := range cols {
		clean := strings.ReplaceAll(c, "-", "")
		clean = strings.ReplaceAll(clean, ":", "")
		clean = strings.ReplaceAll(clean, " ", "")
		if clean != "" {
			return false
		}
	}
	return true
}

func isTableHeader(cols []string) bool {
	headerWords := []string{"name", "server", "package", "description", "link", "url", "repo", "author", "type", "details"}
	for _, c := range cols {
		lower := strings.ToLower(c)
		for _, w := range headerWords {
			if strings.Contains(lower, w) {
				return true
			}
		}
	}
	return false
}

func findTableColumns(cols []string) (int, int, int) {
	nameCol, descCol, linkCol := -1, -1, -1
	for i, c := range cols {
		lower := strings.ToLower(c)
		switch {
		case strings.Contains(lower, "name") || strings.Contains(lower, "server") || strings.Contains(lower, "package") || strings.Contains(lower, "rule"):
			if nameCol == -1 {
				nameCol = i
			}
		case strings.Contains(lower, "desc") || strings.Contains(lower, "detail") || strings.Contains(lower, "summary") || strings.Contains(lower, "info"):
			if descCol == -1 {
				descCol = i
			}
		case strings.Contains(lower, "link") || strings.Contains(lower, "url") || strings.Contains(lower, "repo") || strings.Contains(lower, "github"):
			if linkCol == -1 {
				linkCol = i
			}
		}
	}
	if nameCol == -1 && len(cols) > 0 {
		nameCol = 0
	}
	if descCol == -1 && len(cols) > 1 {
		descCol = 1
	}
	return nameCol, descCol, linkCol
}

func parseTableRow(cols []string, nameIdx, descIdx, linkIdx int, cat, subcat, rawLine string) AwesomeEntry {
	var name, url, desc string

	// Try extracting from name cell (e.g. [Postgres](https://...))
	if nameIdx >= 0 && nameIdx < len(cols) {
		nameCell := cols[nameIdx]
		if match := cellLinkRegex.FindStringSubmatch(nameCell); len(match) > 2 {
			name = strings.TrimSpace(match[1])
			url = strings.TrimSpace(match[2])
		} else {
			name = strings.Trim(nameCell, "*_`")
		}
	}

	// Try extracting URL from link column if URL not found yet
	if url == "" && linkIdx >= 0 && linkIdx < len(cols) {
		linkCell := cols[linkIdx]
		if match := cellLinkRegex.FindStringSubmatch(linkCell); len(match) > 2 {
			url = strings.TrimSpace(match[2])
		} else if strings.HasPrefix(linkCell, "http://") || strings.HasPrefix(linkCell, "https://") {
			url = strings.TrimSpace(linkCell)
		}
	}

	// Try extracting description
	if descIdx >= 0 && descIdx < len(cols) {
		desc = strings.TrimSpace(cols[descIdx])
	}

	return AwesomeEntry{
		Category:    cat,
		Subcategory: subcat,
		Name:        name,
		URL:         url,
		Description: desc,
		RawLine:     rawLine,
	}
}
