package handlers

import (
	"net/http"

	"journal/internal/render"
)

func Fragments(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Fragments",
	}

	if err := render.Render(w, "fragments.html", data); err != nil {
		HandleInternalError(w, r, err)
	}
}