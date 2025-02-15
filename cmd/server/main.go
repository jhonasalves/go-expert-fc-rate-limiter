package main

import (
	"log"
	"net/http"

	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/infra/webserver"
)

func main() {
	srv := webserver.NewServer()

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", srv.Router))
}
