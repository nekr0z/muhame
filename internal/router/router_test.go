package router_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nekr0z/muhame/internal/handlers"
	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/router"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew_JSONUpdate(t *testing.T) {
	t.Parallel()

	log := zap.NewNop()

	tests := []struct {
		name       string
		in         string
		wantCode   int
		wantRes    string
		wantName   string
		wantMetric metrics.Metric
	}{
		{
			name: "gauge",
			in: `{
				"id": "test",
				"type": "gauge",
				"value": 1.2
			}`,
			wantCode: 200,
			wantRes: `{
				"id": "test",
				"type": "gauge",
				"value": 1.2
			}`,
			wantName:   "test",
			wantMetric: metrics.Gauge(1.2),
		},
		{
			name: "counter",
			in: `{
				"id": "test",
				"type": "counter",
				"delta": 1
				}`,
			wantRes: `{
				"id": "test",
				"type": "counter",
				"delta": 1
				}`,
			wantCode:   200,
			wantName:   "test",
			wantMetric: metrics.Counter(1),
		},
		{
			name: "counter with value",
			in: `{
				"id": "test",
				"type": "counter",
				"value": 1.2
			}`,
			wantCode: 400,
		},
		{
			name: "gauge with delta",
			in: `{
				"id": "test",
				"type": "gauge",
				"delta": 8
			}`,
			wantCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			st := &mockStorage{
				t:    t,
				name: tt.wantName,
				m:    tt.wantMetric,
			}

			r := router.New(log, st)

			req := httptest.NewRequest("POST", "/update/", strings.NewReader(tt.in))
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			assert.Equal(t, tt.wantCode, res.Code)

			if tt.wantRes != "" {
				assert.JSONEq(t, tt.wantRes, res.Body.String())
				assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
			}
		})
	}
}

func TestNew_JSONValue(t *testing.T) {
	t.Parallel()

	log := zap.NewNop()

	tests := []struct {
		name     string
		in       string
		m        metrics.Metric
		want     string
		wantCode int
	}{
		{
			name: "gauge",
			in: `{
				"id": "test",
				"type": "gauge"
			}`,
			m: metrics.Gauge(1.2),
			want: `{
				"id": "test",
				"type": "gauge",
				"value": 1.2
			}`,
			wantCode: 200,
		},
		{
			name: "counter",
			in: `{
				"id": "test",
				"type": "counter"
			}`,
			m: metrics.Counter(2),
			want: `{
				"id": "test",
				"type": "counter",
				"delta": 2
			}`,
			wantCode: 200,
		},
		{
			name: "counter",
			in: `{
				"id": "unexpected",
				"type": "counter"
			}`,
			m: metrics.Counter(2),
			want: `{
				"id": "unexpected",
				"type": "counter"
			}`,
			wantCode: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			st := &mockStorage{
				t:    t,
				name: "test",
				m:    tt.m,
			}

			r := router.New(log, st)

			req := httptest.NewRequest("POST", "/value/", strings.NewReader(tt.in))
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			assert.Equal(t, tt.wantCode, res.Code)

			assert.JSONEq(t, tt.want, res.Body.String())
		})
	}
}

var _ handlers.MetricsStorage = &mockStorage{}

type mockStorage struct {
	t    *testing.T
	name string
	m    metrics.Metric
}

func (m *mockStorage) Update(name string, metric metrics.Metric) error {
	m.t.Helper()
	assert.Equal(m.t, m.name, name)
	assert.Equal(m.t, m.m, metric)
	return nil
}

func (m *mockStorage) Get(metricType, name string) (metrics.Metric, error) {
	m.t.Helper()

	if name != m.name {
		return nil, handlers.ErrMetricNotFound
	}

	return m.m, nil
}

func (m *mockStorage) List() ([]string, []metrics.Metric, error) {
	m.t.Helper()
	return nil, nil, nil
}
