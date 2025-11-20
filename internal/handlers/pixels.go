package handlers

import (
	"net/http"
	"sort"

	"journal/internal/content"
    "journal/internal/render"
)

func Pixels(w http.ResponseWriter, r *http.Request) {
    pixels, err := content.LoadPixels()
    if err != nil {
        HandleInternalError(w, r, err)
        return
    }

    sort.Slice(pixels, func(i, j int) bool {
        return pixels[i].Date.After(pixels[j].Date)
    })

    data := map[string]any{
        "Title" : "Pixels",
        "Pixels": pixels,
    }

    if err := render.Render(w, "pixels.html", data); err != nil {
        HandleInternalError(w, r, err)
    }
}