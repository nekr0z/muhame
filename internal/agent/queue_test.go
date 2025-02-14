package agent

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/metrics"
)

func TestSendMetric(t *testing.T) {
	tests := []struct {
		name string
		m    queuedMetric
		want string
	}{
		{
			name: "gauge",
			m: queuedMetric{
				name: "test",
				val:  metrics.Gauge(1.2),
			},
			want: `{"id": "test", "type": "gauge", "value": 1.2}`,
		},
		{
			name: "counter",
			m: queuedMetric{
				name: "another",
				val:  metrics.Counter(2),
			},
			want: `{"id": "another", "type": "counter", "delta": 2}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/update/", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				b, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, tt.want, string(b))
			}))
			defer srv.Close()

			sendMetric(tt.m, srv.URL)
		})
	}
}
