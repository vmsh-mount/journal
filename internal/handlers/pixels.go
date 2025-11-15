package handlers

import (
    "net/http"
    "journal/internal/render"
)

func Pixels(w http.ResponseWriter, r *http.Request) {
    data := map[string]any{
        "Title": "Pixels",
    }

    if err := render.Render(w, "pixels.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}