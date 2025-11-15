package handlers

import (
    "net/http"
    "journal/internal/render"
)

func Home(w http.ResponseWriter, r *http.Request) {
    data := map[string]any {
        "Title": "Welcome",
        "Intro": "This is the home page",
    }

    err := render.Render(w, "home.html", data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
