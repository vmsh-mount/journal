package render

import (
	"html/template"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// TemplateCache holds parsed templates
type TemplateCache struct {
	templates map[string]*template.Template
	mu        sync.RWMutex
}

var cache *TemplateCache

// InitTemplates parses all templates once at startup
func InitTemplates() error {
	cache = &TemplateCache{
		templates: make(map[string]*template.Template),
	}

	templateFiles := []string{
		"base.html",
		"home.html",
		"articles.html",
		"article_detail.html",
		"fragments.html",
		"fragment_detail.html",
		"shelf.html",
		"pixels.html",
		"now.html",
		"uses.html",
	}

	basePath := filepath.Join("internal", "templates", "base.html")

	for _, tmpl := range templateFiles {
		pagePath := filepath.Join("internal", "templates", tmpl)

		// Parse base + page template
		ts, err := template.ParseFiles(basePath, pagePath)
		if err != nil {
			return err
		}

		cache.templates[tmpl] = ts
	}

	return nil
}

// Render executes a cached template
func Render(w http.ResponseWriter, tmpl string, data map[string]any) error {
	if data == nil {
		data = map[string]any{}
	}
	data["Year"] = time.Now().Year()

	cache.mu.RLock()
	ts, ok := cache.templates[tmpl]
	cache.mu.RUnlock()

	if !ok {
		// Fallback: parse on the fly if template not found (shouldn't happen)
		return renderFallback(w, tmpl, data)
	}

	return ts.ExecuteTemplate(w, "base.html", data)
}

// renderFallback is a safety net if template not in cache
func renderFallback(w http.ResponseWriter, tmpl string, data map[string]any) error {
	basePath := filepath.Join("internal", "templates", "base.html")
	pagePath := filepath.Join("internal", "templates", tmpl)

	ts, err := template.ParseFiles(basePath, pagePath)
	if err != nil {
		return err
	}

	return ts.ExecuteTemplate(w, "base.html", data)
}
