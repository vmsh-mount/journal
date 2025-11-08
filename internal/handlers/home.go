package handlers

import (
    "fmt"
    "net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Welcome to vamsi's personal blog")
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "OK")
}