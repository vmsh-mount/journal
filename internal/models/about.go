package models

import (
    "html/template"
    "time"
)

type About struct {
    Slug  string
    Image string
    Date  time.Time
    HTML  template.HTML
}