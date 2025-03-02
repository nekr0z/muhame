package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/storage"
)

func BulkUpdateHandleFunc(st storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bu, ok := st.(bulkUpdater)
		if !ok {
			http.Error(w, "storage does not support bulk updates", http.StatusConflict)
			return
		}

		var jms []metrics.JSONMetric
		if err := json.NewDecoder(r.Body).Decode(&jms); err != nil {
			http.Error(w, fmt.Sprintf("Bad request: %s", err), http.StatusBadRequest)
			return
		}

		nms := toNamed(jms)

		if len(nms) == 0 {
			http.Error(w, "No valid metrics supplied", http.StatusBadRequest)
			return
		}

		err := bu.BulkUpdate(r.Context(), nms)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func toNamed(jms []metrics.JSONMetric) []metrics.Named {
	nms := make([]metrics.Named, 0, len(jms))

	for _, jm := range jms {
		nm, err := jm.Named()
		if err != nil {
			continue
		}

		nms = append(nms, nm)
	}

	return nms
}

type bulkUpdater interface {
	BulkUpdate(context.Context, []metrics.Named) error
}
