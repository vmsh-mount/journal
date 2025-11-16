package handlers

import (
	"net/http"

	"journal/internal/content"
	"journal/internal/render"
	"journal/internal/router"
)

func Articles(w http.ResponseWriter, r *http.Request) {
	articles, err := content.LoadArticles()
	if err != nil {
		HandleInternalError(w, r, err)
		return
	}

	data := map[string]any{
		"Title":    "Articles",
		"Articles": articles,
	}

	if err := render.Render(w, "articles.html", data); err != nil {
		HandleInternalError(w, r, err)
	}
}

func ArticleDetail(w http.ResponseWriter, r *http.Request) {
	slug, ok := router.ExtractPathParam(r, "/articles/")
	if !ok || slug == "" {
		HandleNotFound(w, r)
		return
	}

	articles, err := content.LoadArticles()
	if err != nil {
		HandleInternalError(w, r, err)
		return
	}

	// Find article by slug
	for _, a := range articles {
		if a.Slug == slug {
			data := map[string]any{
				"Title":   a.Title,
				"Content": a.HTML,
			}
			if err := render.Render(w, "article_detail.html", data); err != nil {
				HandleInternalError(w, r, err)
			}
			return
		}
	}

	HandleNotFound(w, r)
}
