package handlers

import (
	"net/http"
	"sort"
	"strings"

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

// getAllTags returns a map of all tags and their counts
func getAllTags(articles []models.Article) map[string]int {
	tags := make(map[string]int)
	for _, article := range articles {
		for _, tag := range article.Tags {
			tags[tag]++
		}
	}
	return tags
}

// filterArticlesByTag returns articles that have all the specified tags
func filterArticlesByTag(articles []models.Article, tagQuery string) []models.Article {
	if tagQuery == "" {
		return articles
	}

	// Split the query into individual tags
	tagFilters := strings.Split(tagQuery, "+")
	for i := range tagFilters {
		tagFilters[i] = strings.TrimSpace(tagFilters[i])
	}

	var filtered []models.Article
	for _, article := range articles {
		tagSet := make(map[string]bool)
		for _, tag := range article.Tags {
			tagSet[tag] = true
		}

		allTagsPresent := true
		for _, filterTag := range tagFilters {
			if !tagSet[filterTag] {
				allTagsPresent = false
				break
			}
		}

		if allTagsPresent {
			filtered = append(filtered, article)
		}
	}

	return filtered
}

func Articles(w http.ResponseWriter, r *http.Request) {
	articles, err := content.LoadArticles()
	if err != nil {
		HandleInternalError(w, r, err)
		return
	}

	// Get tag filters from query parameter
	tagQuery := r.URL.Query().Get("tags")
	var currentTags []string
	if tagQuery != "" {
		currentTags = strings.Split(tagQuery, ",")
		// Filter articles by tags (OR logic - article must have at least one of the selected tags)
		filteredArticles := make([]models.Article, 0, len(articles))
		for _, article := range articles {
			tagSet := make(map[string]bool)
			for _, t := range article.Tags {
				tagSet[t] = true
			}

			// Check if article has any of the selected tags
			for _, tag := range currentTags {
				if tagSet[tag] {
					filteredArticles = append(filteredArticles, article)
					break
				}
			}
		}
		articles = filteredArticles
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

	// Get all unique tags from all articles (not just filtered ones)
	allArticles, err := content.LoadArticles()
	if err != nil {
		HandleInternalError(w, r, err)
		return
	}
	allTags := getAllTags(allArticles)

	data := map[string]any{
		"Title":          "Articles",
		"ArticlesByYear": articlesByYear,
		"Years":          years,
		"AllTags":        allTags,
		"CurrentTags":    currentTags,
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
			// Load all articles to find related ones
			allArticles, err := content.LoadArticles()
			if err != nil {
				HandleInternalError(w, r, err)
				return
			}

			data := map[string]any{
				"Title":           a.Title,
				"Date":            a.Date,
				"Content":         a.HTML,
				"TableOfContents": a.TableOfContents,
				"Tags":            a.Tags,
			}
			if err := render.Render(w, "article_detail.html", data); err != nil {
				HandleInternalError(w, r, err)
			}
			return
		}
	}

	HandleNotFound(w, r)
}
