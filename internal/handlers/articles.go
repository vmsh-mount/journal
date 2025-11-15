package handlers

import (
    "net/http"

    "journal/internal/render"
    "journal/internal/content"
)

func Articles(w http.ResponseWriter, r *http.Request) {
    articles, err := content.LoadArticles()
    if err != nil {
        http.Error(w, "Failed to load articles", http.StatusInternalServerError)
        return
    }

    data := map[string]any {
        "Title": "Articles",
        "Articles": articles,
    }

    if err := render.Render(w, "articles.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func ArticleDetail(w http.ResponseWriter, r *http.Request) {
    slug := r.URL.Path[len("/articles/"):]

    articles, err := content.LoadArticles()
    if err != nil {
        http.Error(w, "Failed to load articles", http.StatusInternalServerError)
        return
    }

    // todo: optimise this later
    for _, a := range articles {
        if a.Slug == slug {
            data := map[string]any {
                "Content": a.HTML,
            }
            if err := render.Render(w, "article_detail.html", data); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
            }
            return
        }
    }

    http.NotFound(w, r)
}
