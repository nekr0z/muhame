package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nekr0z/muhame/internal/handlers"
	"github.com/nekr0z/muhame/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandleFunc(storage.NewMemStorage()))

	return http.ListenAndServe(":8080", r)
}
