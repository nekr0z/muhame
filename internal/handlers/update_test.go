// Package handlers contains HTTP handlers used in the project.
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/metrics"
)

func TestUpdateHandleFunc(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{
			name:   "gauge",
			method: "POST",
			path:   "/gauge/test/1.1",
			want:   http.StatusOK,
		},
		{
			name:   "counter",
			method: "POST",
			path:   "/counter/test/11",
			want:   http.StatusOK,
		},
		{
			name:   "wrong value",
			method: "POST",
			path:   "/counter/test/1.1",
			want:   http.StatusBadRequest,
		},
		{
			name:   "wrong type",
			method: "POST",
			path:   "/hippopotamus/test/1.1",
			want:   http.StatusBadRequest,
		},
		{
			name:   "wrong method",
			method: "GET",
			path:   "/gauge/test/1.1",
			want:   http.StatusMethodNotAllowed,
		},
		{
			name:   "no name",
			method: "POST",
			path:   "/gauge/",
			want:   http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			r := chi.NewRouter()
			r.Post("/{type}/{name}/{value}", UpdateHandleFunc(zeroMetricStorage{}))

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
