package models

import (
	"html/template"
	"time"
)

type Article struct {
	Slug            string
	Title           string
	Summary         string
	Date            time.Time
	HTML            template.HTML
	TableOfContents template.HTML
	Tags            []string
}

// Year returns the year the article was published
func (a Article) Year() int {
	return a.Date.Year()
}
