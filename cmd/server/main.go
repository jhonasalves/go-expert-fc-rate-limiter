package main

import (
	"log"
	"net/http"

	webserver "github.com/jhonasalves/go-expert-fc-rate-limiter/internal/server"
)

func main() {
	srv := webserver.NewServer()

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", srv.Router))
}
