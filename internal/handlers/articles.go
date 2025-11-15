package handlers

import (
    "net/http"
    "journal/internal/render"
)

func Articles(w http.ResponseWriter, r *http.Request) {
    data := map[string]any{
        "Title": "Articles",
    }

    if err := render.Render(w, "articles.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}