// Package handlers contains HTTP handlers used in the project.
package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nekr0z/muhame/internal/metrics"
)

// UpdateHandleFunc returns the handler for the /update/ endpoint.
func UpdateHandleFunc(st MetricsStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err error
			m   metrics.Metric
		)

		value := chi.URLParam(r, "value")

		switch chi.URLParam(r, "type") {
		case "gauge":
			m, err = metrics.ParseGauge(value)
		case "counter":
			m, err = metrics.ParseCounter(value)
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request: %s", err), http.StatusBadRequest)
			return
		}

		if err := st.Update(chi.URLParam(r, "name"), m); err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}
	}
}
