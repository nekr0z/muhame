// Package router implements router for server.
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nekr0z/muhame/internal/handlers"
)

func New(st handlers.MetricsStorage) http.Handler {
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandleFunc(st))
	r.Get("/value/{type}/{name}", handlers.ValueHandleFunc(st))
	r.Get("/", handlers.RootHandleFunc(st))

	return r
}
