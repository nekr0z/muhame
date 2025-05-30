package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/storage"
)

// ValueHandleFunc returns the handler for the /value/*/* endpoint.
func ValueHandleFunc(st getter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := st.Get(r.Context(), chi.URLParam(r, "type"), chi.URLParam(r, "name"))
		if err != nil {
			if errors.Is(err, storage.ErrMetricNotFound) {
				http.Error(w, "Metric not found.", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(m.String()))
		if err != nil {
			panic(err)
		}
	}
}

// ValueJSONHandleFunc returns the handler for the /value/ endpoint.
func ValueJSONHandleFunc(st getter) func(http.ResponseWriter, *http.Request) {
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
		t := jm.MType

		m, err := st.Get(r.Context(), t, name)
		if err != nil {
			if errors.Is(err, storage.ErrMetricNotFound) {
				respondJSONNotFound(w, t, name)
				return
			}

			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		_, err = w.Write(metrics.ToJSON(m, name))
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}
	}
}

func respondJSONNotFound(w http.ResponseWriter, t, name string) {
	bb, err := json.Marshal(
		metrics.JSONMetric{
			ID:    name,
			MType: t,
		},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNotFound)

	_, err = w.Write(bb)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
}

type getter interface {
	Get(context.Context, string, string) (metrics.Metric, error)
}
