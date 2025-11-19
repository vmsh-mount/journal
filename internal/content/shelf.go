package content

import (
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"

	"journal/internal/models"
	"journal/internal/render"
)

func LoadShelfItems() ([]models.ShelfItem, error) {
	files, err := filepath.Glob("internal/content/shelf/*/*.md")
	if err != nil {
		return nil, err
	}

	var items []models.ShelfItem

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		md := string(data)
		md, frontmatter := parseFrontmatter(md)

		html, _, err := render.MarkdownToHTML(md)
		if err != nil {
			return nil, err
		}

		slug := strings.TrimSuffix(filepath.Base(file), ".md")
		title := extractTitle(md, frontmatter)
		summary := extractSummary(md, frontmatter)
		date := extractDate(file, frontmatter)
		category := frontmatter["category"]
		if category == "" {
			category = "books"
		}

		items = append(items, models.ShelfItem{
			Slug:     slug,
			Title:    title,
			Summary:  summary,
			Date:     date,
			Category: strings.ToLower(category),
			HTML:     template.HTML(html),
		})
	}

	return items, nil
}
