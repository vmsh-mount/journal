package content

import (
    "io/ioutil"
    "path/filepath"
    "strings"
    "html/template"

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

        html, err := render.MarkdownToHTML(md)
        if err != nil {
            return nil, err
        }

        slug := strings.TrimSuffix(filepath.Base(file), ".md")
        title := extractTitle(md)

        articles = append(articles, models.Article{
            Slug: slug,
            Title: title,
            HTML: template.HTML(html),
        })
    }
    return articles, nil
}

func extractTitle(md string) string {
    lines := strings.Split(md, "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "# ") {
            return strings.TrimPrefix(line, "# ")
        }
    }
    return "Untitled"
}