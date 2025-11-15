package main

import (
    "log"
    "journal/internal/server"
)

func main() {
    s := server.New()
    err := s.Start(":8080")
    if err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
    log.Println("Server running at http://localhost:8080")
}