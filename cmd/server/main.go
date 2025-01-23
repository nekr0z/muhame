package main

import (
	"log"
	"net/http"

	"github.com/nekr0z/muhame/internal/handlers"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.UpdateHandler)

	return http.ListenAndServe(":8080", mux)
}
