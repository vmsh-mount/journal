package router

import (
	"net/http"
	"strings"
)

// ExtractSlug extracts the slug from a URL path
// Example: "/articles/my-post" -> "my-post"
func ExtractSlug(r *http.Request, prefix string) string {
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	slug := strings.TrimPrefix(path, prefix)
	slug = strings.Trim(slug, "/")
	return slug
}

// ExtractPathParam extracts a path parameter from URL
// Example: "/articles/my-post" with prefix "/articles/" returns "my-post"
func ExtractPathParam(r *http.Request, prefix string) (string, bool) {
	slug := ExtractSlug(r, prefix)
	if slug == "" {
		return "", false
	}
	return slug, true
}

