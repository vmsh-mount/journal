package models

import (
	"html/template"
	"time"
)

type ShelfItem struct {
	Slug     string
	Title    string
	Category string
	Summary  string
	Date     time.Time
	HTML     template.HTML
}

func (s ShelfItem) Year() int {
	return s.Date.Year()
}
