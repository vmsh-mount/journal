package content

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"journal/internal/models"
	"journal/internal/render"
)

func LoadArticles() ([]models.Article, error) {
	files, err := filepath.Glob("internal/content/articles/*.md")
	if err != nil {
		return nil, err
	}

	var articles []models.Article

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		md := string(data)
		md, frontmatter := parseFrontmatter(md)

		html, toc, err := render.MarkdownToHTML(md)
		if err != nil {
			return nil, err
		}

		slug := strings.TrimSuffix(filepath.Base(file), ".md")
		title := extractTitle(md, frontmatter)
		summary := extractSummary(md, frontmatter)
		date := extractDate(file, frontmatter)

		articles = append(articles, models.Article{
			Slug:            slug,
			Title:           title,
			Summary:         summary,
			Date:            date,
			HTML:            template.HTML(html),
			TableOfContents: template.HTML(toc),
		})
	}
	return articles, nil
}

// parseFrontmatter extracts YAML frontmatter from markdown
// Returns markdown without frontmatter and a map of frontmatter fields
func parseFrontmatter(md string) (string, map[string]string) {
	frontmatter := make(map[string]string)

	// Check for frontmatter delimiter (---)
	if !strings.HasPrefix(md, "---") {
		return md, frontmatter
	}

	// Find the end of frontmatter (second ---)
	firstDelim := strings.Index(md, "---")
	if firstDelim == -1 {
		return md, frontmatter
	}

	// Find second delimiter
	remaining := md[firstDelim+3:]
	// Skip whitespace/newlines
	remaining = strings.TrimLeft(remaining, " \n\r\t")
	secondDelim := strings.Index(remaining, "---")
	if secondDelim == -1 {
		return md, frontmatter
	}

	// Extract frontmatter content
	frontmatterContent := remaining[:secondDelim]
	markdownContent := remaining[secondDelim+3:]

	// Parse frontmatter lines
	lines := strings.Split(frontmatterContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, `"'`)
			frontmatter[key] = value
		}
	}

	// Return markdown without frontmatter
	return strings.TrimLeft(markdownContent, " \n\r\t"), frontmatter
}

func extractTitle(md string, frontmatter map[string]string) string {
	// Check frontmatter first
	if title, ok := frontmatter["title"]; ok && title != "" {
		return title
	}

	// Fallback to first H1 in markdown
	lines := strings.Split(md, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "Untitled"
}

func extractSummary(md string, frontmatter map[string]string) string {
	if summary, ok := frontmatter["summary"]; ok && summary != "" {
		return summary
	}

	// Remove markdown headers and trim
	paragraphs := strings.Split(md, "\n\n")
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Remove markdown formatting for summary
		plain := trimmed
		plain = strings.ReplaceAll(plain, "**", "")
		plain = strings.ReplaceAll(plain, "*", "")
		plain = strings.ReplaceAll(plain, "_", "")

		// Remove inline links [text](url)
		linkRegex := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
		plain = linkRegex.ReplaceAllString(plain, "$1")

		if len(plain) > 220 {
			return plain[:220] + "..."
		}
		return plain
	}

	return ""
}

func extractDate(file string, frontmatter map[string]string) time.Time {
	// Check frontmatter first
	if dateStr, ok := frontmatter["date"]; ok && dateStr != "" {
		// Try common date formats
		formats := []string{
			"2006-01-02",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, dateStr); err == nil {
				return t
			}
		}
	}

	// Fallback to file modification time
	if info, err := os.Stat(file); err == nil {
		return info.ModTime()
	}

	// Last resort: current time
	return time.Now()
}
