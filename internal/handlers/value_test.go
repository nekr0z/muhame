// Package handlers contains HTTP handlers used in the project.
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/metrics"
)

func TestValueHandleFunc(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantValue  string
	}{
		{
			name:       "gauge",
			method:     "GET",
			path:       "/gauge/test",
			wantStatus: http.StatusOK,
			wantValue:  "1.1",
		},
		{
			name:       "counter",
			method:     "GET",
			path:       "/counter/test",
			wantStatus: http.StatusOK,
			wantValue:  "11",
		},
		{
			name:       "non-existent",
			method:     "GET",
			path:       "/counter/none",
			wantStatus: http.StatusNotFound,
			wantValue:  "Metric not found.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			r := chi.NewRouter()
			r.Get("/{type}/{name}", ValueHandleFunc(oneMetricStorage{}))

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValue, string(body))
		})
	}
}

type oneMetricStorage struct{}

func (z oneMetricStorage) Update(_ string, _ metrics.Metric) error {
	return fmt.Errorf("not implemented")
}

func (oneMetricStorage) Get(t, n string) (metrics.Metric, error) {
	if t == "gauge" && n == "test" {
		return metrics.Gauge(1.1), nil
	}
	if t == "counter" && n == "test" {
		return metrics.Counter(11), nil
	}
	return nil, ErrMetricNotFound
}

func (oneMetricStorage) List() ([]string, []metrics.Metric, error) {
	return []string{
			"test",
			"test",
		},
		[]metrics.Metric{
			metrics.Gauge(1.1),
			metrics.Counter(11),
		}, nil
}
