package main

import (
	"log"
	"journal/internal/server"
)

func main() {
	s, err := server.New()
	if err != nil {
		log.Fatalf("Error initializing server: %v", err)
	}

	if err := s.Start(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}