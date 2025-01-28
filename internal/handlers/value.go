package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ValueHandleFunc returns the handler for the /value/ endpoint.
func ValueHandleFunc(st MetricsStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := st.Get(chi.URLParam(r, "type"), chi.URLParam(r, "name"))
		if err != nil {
			if err == ErrMetricNotFound {
				http.Error(w, "Metric not found.", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(m.String()))
	}
}
