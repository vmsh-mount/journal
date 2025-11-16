package handlers

import (
	"net/http"

	"journal/internal/render"
)

func Shelf(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Shelf",
	}

	if err := render.Render(w, "shelf.html", data); err != nil {
		HandleInternalError(w, r, err)
	}
}