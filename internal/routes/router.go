package routes

import (
    "net/http"
    "journal/internal/handlers"
)

func NewRouter() *http.ServeMux {
    mux := http.NewServeMux() // Mux = Multiplexer — something that can route multiple URLs (paths) to different handlers.

    mux.HandleFunc("/", handlers.HomeHandler)
    mux.HandleFunc("/health", handlers.HealthHandler)

    return mux
}
