package content

import (
    "html/template"
    "io/ioutil"
    "path/filepath"
    "strings"

    "journal/internal/models"
    "journal/internal/render"
)

func LoadAbout() (models.About, error) {
    files, err := filepath.Glob("internal/content/about/about.md")
    if err != nil || len(files) == 0 {
        return models.About{}, err
    }

    data, err := ioutil.ReadFile(files[0])
    if err != nil {
        return models.About{}, err
    }

    md := string(data)
    md, frontmatter := parseFrontmatter(md)

    html, _, err := render.MarkdownToHTML(md)
    if err != nil {
        return models.About{}, err
    }

    slug := strings.TrimSuffix(filepath.Base(files[0]), ".md")
    image := frontmatter["image"]
    date := extractDate(files[0], frontmatter)

    return models.About{
        Slug:  slug,
        Image: image,
        Date:  date,
        HTML:  template.HTML(html),
    }, nil
}
