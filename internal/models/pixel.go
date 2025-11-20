package models

import  (
    "html/template"
    "time"
)

type Pixel struct {
    Slug  string
    Title string
    Image string
    Date  time.Time
    HTML  template.HTML
}
