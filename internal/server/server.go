package server

import (
    "net/http"
    "journal/internal/handlers"
)

type Server struct {
    mux *http.ServeMux
}

// It creates and configures the HTTP server
func New() *Server {
    s := &Server {
        mux: http.NewServeMux(),
    }
    s.registerRoutes()
    return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", handlers.Home)

	s.mux.HandleFunc("/articles", handlers.Articles)
	s.mux.HandleFunc("/fragments", handlers.Fragments)
	s.mux.HandleFunc("/shelf", handlers.Shelf)
	s.mux.HandleFunc("/pixels", handlers.Pixels)
	s.mux.HandleFunc("/now", handlers.Now)
	s.mux.HandleFunc("/uses", handlers.Uses)

	// serve the static files
	fileServer := http.FileServer(http.Dir("static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
}

// Start runs the HTTP server
func (s *Server) Start(port string) error {
    return http.ListenAndServe(port, s.mux)
}

/** What & Why's:
  - "net/http":  provides everything needed to build HTTP servers and clients
  - mux: a pointer to an http.ServeMux
  - What's ServeMux:
    - It's Go's built-in HTTP request multiplexer (router)
    - It maps URL paths to handler functions
    So, mux is the core router that determines which handler serves each incoming request

   - http.FileServer reads files from your filesystem
   - We point it to static/ directory.
*/