package main

import (
	"log"

	"myproject/internal/server"
)

const version = "1.0.0"

func main() {
	// Create and start server
	srv := server.NewServer(":8080", version)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
