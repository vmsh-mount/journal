package handlers

import (
	"net/http"
	"sort"

	"journal/internal/content"
	"journal/internal/models"
	"journal/internal/render"
	"journal/internal/router"
)

type FragmentsByYear struct {
	Year      int
	Fragments []models.Fragment
}

func Fragments(w http.ResponseWriter, r *http.Request) {
	fragments, err := content.LoadFragments()
	if err != nil {
		HandleInternalError(w, r, err)
		return
	}

	sort.Slice(fragments, func(i, j int) bool {
		return fragments[i].Date.After(fragments[j].Date)
	})

	yearMap := make(map[int][]models.Fragment)
	for _, fragment := range fragments {
		year := fragment.Year()
		yearMap[year] = append(yearMap[year], fragment)
	}

	var fragmentsByYear []FragmentsByYear
	for year, yearFragments := range yearMap {
		fragmentsByYear = append(fragmentsByYear, FragmentsByYear{
			Year:      year,
			Fragments: yearFragments,
		})
	}

	sort.Slice(fragmentsByYear, func(i, j int) bool {
		return fragmentsByYear[i].Year > fragmentsByYear[j].Year
	})

	years := make([]int, 0, len(yearMap))
	for year := range yearMap {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))

	data := map[string]any{
		"Title":           "Fragments",
		"FragmentsByYear": fragmentsByYear,
		"Years":           years,
	}

	if err := render.Render(w, "fragments.html", data); err != nil {
		HandleInternalError(w, r, err)
	}
}

func FragmentDetail(w http.ResponseWriter, r *http.Request) {
	slug, ok := router.ExtractPathParam(r, "/fragments/")
	if !ok || slug == "" {
		HandleNotFound(w, r)
		return
	}

	fragments, err := content.LoadFragments()
	if err != nil {
		HandleInternalError(w, r, err)
		return
	}

	for _, f := range fragments {
		if f.Slug == slug {
			data := map[string]any{
				"Title":   f.Title,
				"Date":    f.Date,
				"Image":   f.Image,
				"Content": f.HTML,
			}
			if err := render.Render(w, "fragment_detail.html", data); err != nil {
				HandleInternalError(w, r, err)
			}
			return
		}
	}

	HandleNotFound(w, r)
}
