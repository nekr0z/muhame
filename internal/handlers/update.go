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

// UpdateHandleFunc returns the handler for the /update/*/* endpoint.
func UpdateHandleFunc(st updater) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		value := chi.URLParam(r, "value")
		t := chi.URLParam(r, "type")

		m, err := metrics.Parse(t, value)
		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request: %s", err), http.StatusBadRequest)
			return
		}

		if err := st.Update(r.Context(), metrics.Named{
			Name:   chi.URLParam(r, "name"),
			Metric: m,
		}); err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}
	}
}

// UpdateJSONHandleFunc returns the handler for the /update/ endpoint.
func UpdateJSONHandleFunc(st getUpdater) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()

		var jm metrics.JSONMetric
		if err := json.NewDecoder(r.Body).Decode(&jm); err != nil {
			http.Error(w, fmt.Sprintf("Bad request: %s", err), http.StatusBadRequest)
			return
		}

		name := jm.ID

		nm, err := jm.Named()
		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request: %s", err), http.StatusBadRequest)
			return
		}

		err = st.Update(r.Context(), nm)
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		m, err := st.Get(r.Context(), nm.Type(), nm.Name)
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
	Update(context.Context, metrics.Named) error
}

type getUpdater interface {
	getter
	updater
}
