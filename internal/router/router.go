// Package router implements router for server.
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nekr0z/muhame/internal/handlers"
	"github.com/nekr0z/muhame/internal/storage"
	"go.uber.org/zap"
)

func New(log *zap.Logger, st storage.Storage) http.Handler {
	r := chi.NewRouter()

	r.Use(logger(log))
	r.Use(acceptGzip)
	r.Use(respondGzip)

	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandleFunc(st))
	r.Post("/update/", handlers.UpdateJSONHandleFunc(st))
	r.Post("/updates/", handlers.BulkUpdateHandleFunc(st))
	r.Post("/value/", handlers.ValueJSONHandleFunc(st))
	r.Get("/value/{type}/{name}", handlers.ValueHandleFunc(st))
	r.Get("/ping", handlers.PingHandleFunc(st))
	r.Get("/", handlers.RootHandleFunc(st))

	return r
}

type middleware func(http.Handler) http.Handler
