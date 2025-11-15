package render

import (
    "html/template"
    "net/http"
    "path/filepath"
    "time"
)

func Render(w http.ResponseWriter, tmpl string, data map[string]any) error {
    if data == nil {
        data = map[string]any{}
    }
    data["Year"] = time.Now().Year()

    basePath := filepath.Join("internal", "templates", "base.html")
    pagePath := filepath.Join("internal", "templates", tmpl)

    templates, err := template.ParseFiles(basePath, pagePath)
    if err != nil {
        return err
    }

    return templates.ExecuteTemplate(w, "base.html", data)
}

/** What & Why's:
    - It avoids duplicating template code in every handler.
    - It loads both layout + page templates.
*/