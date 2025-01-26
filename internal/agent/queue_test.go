package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/metrics"
)

func TestEndpoint(t *testing.T) {
	type args struct {
		addr       string
		metricType string
		name       string
		value      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{"http://localhost:8080", "counter", "test", "1"},
			want: "http://localhost:8080/update/counter/test/1",
		},
		{
			name: "trailing slash",
			args: args{"http://localhost:8080/", "gauge", "test", "1.1"},
			want: "http://localhost:8080/update/gauge/test/1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := endpoint(tt.args.addr, tt.args.metricType, tt.args.name, tt.args.value); got != tt.want {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

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
			want: "/update/gauge/test/1.2",
		},
		{
			name: "counter",
			m: queuedMetric{
				name: "another",
				val:  metrics.Counter(2),
			},
			want: "/update/counter/another/2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.want, r.URL.Path)
			}))
			defer srv.Close()

			sendMetric(tt.m, srv.URL)
		})
	}
}
