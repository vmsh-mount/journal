package models

import (
	"html/template"
	"time"
)

type Fragment struct {
	Slug  string
	Title string
	Image string
	Date  time.Time
	HTML  template.HTML
}

func (f Fragment) Year() int {
	return f.Date.Year()
}
