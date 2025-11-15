package models

import (
    "html/template"
)

type Article struct {
    Slug string
    Title string
    HTML  template.HTML
}