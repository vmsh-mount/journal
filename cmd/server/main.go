package main

import (
    "log"
    "net/http"
    "journal/internal/routes"
)

func main() {
    mux := routes.NewRouter()

    err := http.ListenAndServe(":8080", mux)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Server running at http://localhost:8080")
}