package handlers

import (
	"net/http"

	"journal/internal/content"
    "journal/internal/render"
)

func About(w http.ResponseWriter, r *http.Request) {
    about, err := content.LoadAbout()
    if err != nil {
        HandleInternalError(w, r, err)
        return
    }

	data := map[string]any{
		"Title": "About",
		"About": about,
	}

	if err := render.Render(w, "about.html", data); err != nil {
		HandleInternalError(w, r, err)
	}
}