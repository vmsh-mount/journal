package content

import (
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"

	"journal/internal/models"
	"journal/internal/render"
)

func LoadFragments() ([]models.Fragment, error) {
	files, err := filepath.Glob("internal/content/fragments/*.md")
	if err != nil {
		return nil, err
	}

	var fragments []models.Fragment

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
		date := extractDate(file, frontmatter)
		image := fetchImageUrl(frontmatter)

		fragments = append(fragments, models.Fragment{
			Slug:  slug,
			Title: title,
		    Image: image,
			Date:  date,
			HTML:  template.HTML(html),
		})
	}

	return fragments, nil
}

func fetchImageUrl(frontmatter map[string]string) string {
    return frontmatter["image"]
}
