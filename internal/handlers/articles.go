package handlers

import (
	"net/http"
	"sort"

	"journal/internal/content"
	"journal/internal/models"
	"journal/internal/render"
	"journal/internal/router"
)

// ArticlesByYear groups articles by year
type ArticlesByYear struct {
	Year     int
	Articles []models.Article
}

func Articles(w http.ResponseWriter, r *http.Request) {
	articles, err := content.LoadArticles()
	if err != nil {
		HandleInternalError(w, r, err)
		return
	}

	// Sort articles by date (newest first)
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date.After(articles[j].Date)
	})

	// Group by year
	yearMap := make(map[int][]models.Article)
	for _, article := range articles {
		year := article.Year()
		yearMap[year] = append(yearMap[year], article)
	}

	// Convert to slice and sort years (newest first)
	var articlesByYear []ArticlesByYear
	for year, yearArticles := range yearMap {
		articlesByYear = append(articlesByYear, ArticlesByYear{
			Year:     year,
			Articles: yearArticles,
		})
	}

	sort.Slice(articlesByYear, func(i, j int) bool {
		return articlesByYear[i].Year > articlesByYear[j].Year
	})

	// Get all unique years for navigation
	years := make([]int, 0, len(yearMap))
	for year := range yearMap {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))

	data := map[string]any{
		"Title":          "Articles",
		"ArticlesByYear": articlesByYear,
		"Years":          years,
	}

	if err := render.Render(w, "articles.html", data); err != nil {
		HandleInternalError(w, r, err)
	}
}

func ArticleDetail(w http.ResponseWriter, r *http.Request) {
	slug, ok := router.ExtractPathParam(r, "/articles/")
	if !ok || slug == "" {
		HandleNotFound(w, r)
		return
	}

	articles, err := content.LoadArticles()
	if err != nil {
		HandleInternalError(w, r, err)
		return
	}

	// Find article by slug
	for _, a := range articles {
		if a.Slug == slug {
			data := map[string]any{
				"Title":           a.Title,
				"Date":            a.Date,
				"Content":         a.HTML,
				"TableOfContents": a.TableOfContents,
			}
			if err := render.Render(w, "article_detail.html", data); err != nil {
				HandleInternalError(w, r, err)
			}
			return
		}
	}

	HandleNotFound(w, r)
}
