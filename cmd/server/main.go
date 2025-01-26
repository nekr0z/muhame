package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/handlers"
	"github.com/nekr0z/muhame/internal/storage"
)

func main() {
	flag.Parse()

	if err := run(flagNetAddress); err != nil {
		log.Fatal(err)
	}
}

func run(addr addr.NetAddress) error {
	st := storage.NewMemStorage()

	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandleFunc(st))
	r.Get("/value/{type}/{name}", handlers.ValueHandleFunc(st))
	r.Get("/", handlers.RootHandleFunc(st))

	log.Printf("running server on %s", addr.String())
	return http.ListenAndServe(addr.String(), r)
}
