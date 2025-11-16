package handlers

import (
	"net/http"

	"journal/internal/render"
)

func Uses(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Uses",
	}

	if err := render.Render(w, "uses.html", data); err != nil {
		HandleInternalError(w, r, err)
	}
}