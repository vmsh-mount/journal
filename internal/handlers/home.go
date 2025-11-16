package handlers

import (
	"net/http"

	"journal/internal/render"
)

func Home(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Welcome",
		"Intro": "This is the home page",
	}

	if err := render.Render(w, "home.html", data); err != nil {
		HandleInternalError(w, r, err)
	}
}
