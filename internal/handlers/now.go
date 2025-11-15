package handlers

import (
    "net/http"
    "journal/internal/render"
)

func Now(w http.ResponseWriter, r *http.Request) {
    data := map[string]any{
        "Title": "Now",
    }

    if err := render.Render(w, "now.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}