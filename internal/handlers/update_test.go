// Package handlers contains HTTP handlers used in the project.
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{
			name:   "gauge",
			method: "POST",
			path:   "/update/gauge/test/1.1",
			want:   http.StatusOK,
		},
		{
			name:   "counter",
			method: "POST",
			path:   "/update/counter/test/11",
			want:   http.StatusOK,
		},
		{
			name:   "wrong value",
			method: "POST",
			path:   "/update/counter/test/1.1",
			want:   http.StatusBadRequest,
		},
		{
			name:   "wrong type",
			method: "POST",
			path:   "/update/hippopotamus/test/1.1",
			want:   http.StatusBadRequest,
		},
		{
			name:   "wrong method",
			method: "GET",
			path:   "/update/gauge/test/1.1",
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "no name",
			method: "POST",
			path:   "/update/gauge/",
			want:   http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", UpdateHandleFunc(zeroMetricStorage{}))

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want, res.StatusCode)
		})
	}
}

type zeroMetricStorage struct{}

func (z zeroMetricStorage) Update(_ string, _ metrics.Metric) error {
	return nil
}

func (zeroMetricStorage) Get(_, _ string) (metrics.Metric, error) {
	return nil, ErrMetricNotFound
}
