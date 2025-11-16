package handlers

import (
	"log"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string
	Code    int
}

// HandleError writes an error response and logs it
func HandleError(w http.ResponseWriter, r *http.Request, err error, code int) {
	log.Printf("Error [%s %s]: %v", r.Method, r.URL.Path, err)
	http.Error(w, err.Error(), code)
}

// HandleNotFound returns a 404 response
func HandleNotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

// HandleInternalError returns a 500 response
func HandleInternalError(w http.ResponseWriter, r *http.Request, err error) {
	HandleError(w, r, err, http.StatusInternalServerError)
}

