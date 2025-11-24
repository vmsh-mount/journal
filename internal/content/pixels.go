package content

import (
    "html/template"
    "io/ioutil"
    "path/filepath"
    "strings"

    "journal/internal/models"
    "journal/internal/render"
)

func LoadPixels() ([]models.Pixel, error) {
    files, err := filepath.Glob("internal/content/pixels/*.md")
    if err != nil {
        return nil, err
    }

    var pixels []models.Pixel

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
        image := fetchImageUrl(frontmatter)
        date := extractDate(file, frontmatter)

        pixels = append(pixels, models.Pixel{
            Slug:  slug,
            Title: title,
            Image: image,
            Date:  date,
            HTML:  template.HTML(html),
        })
    }

    return pixels, nil

}
