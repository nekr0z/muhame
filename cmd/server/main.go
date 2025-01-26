package main

import (
	"log"
	"net/http"

	"github.com/nekr0z/muhame/internal/handlers"
	"github.com/nekr0z/muhame/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	mux := http.NewServeMux()

	update := handlers.UpdateHandleFunc(storage.NewMemStorage())
	mux.HandleFunc("/update/", update)

	return http.ListenAndServe(":8080", mux)
}
