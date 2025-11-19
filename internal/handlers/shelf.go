package handlers

import (
	"net/http"
	"sort"

	"journal/internal/content"
	"journal/internal/models"
	"journal/internal/render"
	"journal/internal/router"
)

type ShelfSection struct {
	Key   string
	Title string
	Items []models.ShelfItem
}

var shelfSections = []ShelfSection{
	{Key: "books", Title: "Bookshelf"},
	{Key: "papers", Title: "Papers"},
	{Key: "misc", Title: "Miscellany"},
}

func Shelf(w http.ResponseWriter, r *http.Request) {
	items, err := content.LoadShelfItems()
	if err != nil {
		HandleInternalError(w, r, err)
		return
	}

	sectionMap := make(map[string][]models.ShelfItem)
	for _, item := range items {
		key := item.Category
		if key == "" {
			key = "books"
		}
		sectionMap[key] = append(sectionMap[key], item)
	}

	var sections []ShelfSection
	for _, def := range shelfSections {
		entries := sectionMap[def.Key]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Date.After(entries[j].Date)
		})
		if len(entries) == 0 {
			continue
		}
		sections = append(sections, ShelfSection{
			Key:   def.Key,
			Title: def.Title,
			Items: entries,
		})
	}

	data := map[string]any{
		"Title":         "Shelf",
		"ShelfNav":      shelfSections,
		"ShelfSections": sections,
	}

	if err := render.Render(w, "shelf.html", data); err != nil {
		HandleInternalError(w, r, err)
	}
}

func ShelfDetail(w http.ResponseWriter, r *http.Request) {
	slug, ok := router.ExtractPathParam(r, "/shelf/")
	if !ok || slug == "" {
		HandleNotFound(w, r)
		return
	}

	items, err := content.LoadShelfItems()
	if err != nil {
		HandleInternalError(w, r, err)
		return
	}

	for _, item := range items {
		if item.Slug == slug {
			data := map[string]any{
				"Title":    item.Title,
				"Date":     item.Date,
				"Category": item.Category,
				"Content":  item.HTML,
			}
			if err := render.Render(w, "shelf_detail.html", data); err != nil {
				HandleInternalError(w, r, err)
			}
			return
		}
	}

	HandleNotFound(w, r)
}
