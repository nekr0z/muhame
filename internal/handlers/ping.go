package handlers

import (
	"errors"
	"net/http"

	"github.com/nekr0z/muhame/internal/storage"
)

func PingHandleFunc(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := st.Ping(r.Context())

		if err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.Is(err, storage.ErrNotADatabase) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
