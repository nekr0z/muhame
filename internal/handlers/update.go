// Package handlers contains HTTP handlers used in the project.
package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/nekr0z/muhame/internal/metrics"
)

// UpdateHandleFunc is the handler for the /update/ endpoint.
func UpdateHandleFunc(st MetricsStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		p := strings.Split(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/update/"), "/"), "/")

		if len(p) < 2 {
			http.Error(w, "Bad request, but the spec dictates a 404.", http.StatusNotFound)
			return
		}

		if len(p) != 3 {
			return
		}

		var (
			err error
			m   metrics.Metric
		)

		switch p[0] {
		case "gauge":
			m, err = metrics.ParseGauge(p[2])
		case "counter":
			m, err = metrics.ParseCounter(p[2])
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request: %s", err), http.StatusBadRequest)
			return
		}

		if err := st.Update(p[1], m); err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
