package handlers

import (
	"context"
	"net/http"

	"github.com/nekr0z/muhame/internal/storage"
)

// PingHandleFunc returns the handler for the /ping/ endpoint.
func PingHandleFunc(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := st.(pingable)
		if !ok {
			http.Error(w, "pinging the used storage makes no sense", http.StatusConflict)
			return
		}

		err := p.Ping(r.Context())

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type pingable interface {
	Ping(context.Context) error
}
