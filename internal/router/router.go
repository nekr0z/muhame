// Package router implements router for server.
package router

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/nekr0z/muhame/internal/handlers"
	"github.com/nekr0z/muhame/internal/storage"
)

func New(log *zap.Logger, st storage.Storage, key string) http.Handler {
	r := chi.NewRouter()

	r.Use(logger(log))

	if key != "" {
		log.Info("using key to verify messages", zap.String("key", key))
		r.Use(checkSig(key))
		r.Use(addSig(key))
	}

	r.Use(acceptGzip)
	r.Use(respondGzip)

	r.Post("/update/{type}/{name}/{value}", handlers.UpdateHandleFunc(st))
	r.Post("/update/", handlers.UpdateJSONHandleFunc(st))
	r.Post("/updates/", handlers.BulkUpdateHandleFunc(st))
	r.Post("/value/", handlers.ValueJSONHandleFunc(st))
	r.Get("/value/{type}/{name}", handlers.ValueHandleFunc(st))
	r.Get("/ping", handlers.PingHandleFunc(st))
	r.Get("/", handlers.RootHandleFunc(st))

	r.Handle("/debug/pprof/*", http.DefaultServeMux)

	return r
}

type middleware func(http.Handler) http.Handler
