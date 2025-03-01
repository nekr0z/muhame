// Package handlers contains HTTP handlers used in the project.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nekr0z/muhame/internal/metrics"
)

// UpdateHandleFunc returns the handler for the /update/ endpoint.
func UpdateHandleFunc(st updater) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		value := chi.URLParam(r, "value")
		t := chi.URLParam(r, "type")

		m, err := metrics.Parse(t, value)
		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request: %s", err), http.StatusBadRequest)
			return
		}

		if err := st.Update(r.Context(), chi.URLParam(r, "name"), m); err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}
	}
}

func UpdateJSONHandleFunc(st getUpdater) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var jm metrics.JSONMetric
		if err := json.NewDecoder(r.Body).Decode(&jm); err != nil {
			http.Error(w, fmt.Sprintf("Bad request: %s", err), http.StatusBadRequest)
			return
		}

		name := jm.ID

		m, err := jm.Metric()
		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request: %s", err), http.StatusBadRequest)
			return
		}

		if err := st.Update(r.Context(), name, m); err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		m, err = st.Get(r.Context(), m.Type(), name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		_, err = w.Write(metrics.ToJSON(m, name))
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
		}
	}
}

type updater interface {
	Update(context.Context, string, metrics.Metric) error
}

type getUpdater interface {
	getter
	updater
}
