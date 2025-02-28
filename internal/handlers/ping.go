package handlers

import (
	"net/http"

	"github.com/nekr0z/muhame/internal/storage"
)

func PingHandleFunc(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := st.Ping(r.Context())
		switch err {
		case storage.ErrNotADatabase:
			http.Error(w, err.Error(), http.StatusConflict)
			return
		case nil:
			w.WriteHeader(http.StatusOK)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
