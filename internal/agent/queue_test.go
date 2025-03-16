package agent

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/httpclient"
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

				b := r.Body
				defer b.Close()

				if r.Header.Get("Content-Encoding") == "gzip" {
					var err error
					b, err = gzip.NewReader(b)
					assert.NoError(t, err)
				}

				bb, err := io.ReadAll(b)
				assert.NoError(t, err)

				assert.JSONEq(t, tt.want, string(bb))
			}))
			defer srv.Close()

			sendMetric(httpclient.New(), tt.m, srv.URL)
		})
	}
}

func TestSendBulk(t *testing.T) {
	mm := []queuedMetric{
		{
			name: "test",
			val:  metrics.Gauge(1.2),
		},
		{
			name: "another",
			val:  metrics.Counter(2),
		},
	}

	want := `[
{"id": "test", "type": "gauge", "value": 1.2},
{"id": "another", "type": "counter", "delta": 2}
]`

	q := queue{}

	for _, m := range mm {
		q.push(m)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/updates/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		b := r.Body
		defer b.Close()

		var err error
		b, err = gzip.NewReader(b)
		assert.NoError(t, err)

		bb, err := io.ReadAll(b)
		assert.NoError(t, err)

		assert.JSONEq(t, want, string(bb))
	}))
	defer srv.Close()

	q.sendMetrics(httpclient.New(), srv.URL)
}

func TestSendBulk_Fallback(t *testing.T) {
	mm := []queuedMetric{
		{
			name: "test",
			val:  metrics.Gauge(1.2),
		},
		{
			name: "another",
			val:  metrics.Counter(2),
		},
	}

	q := queue{}

	for _, m := range mm {
		q.push(m)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/updates/" {
			http.Error(w, "storage does not support bulk updates", http.StatusConflict)
			return
		}

		assert.Equal(t, "/update/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
	}))
	defer srv.Close()

	q.sendMetrics(httpclient.New(), srv.URL)
}

func TestSendBulk_EmptyQueue(t *testing.T) {
	q := queue{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not send metrics from empty queue")
	}))
	defer srv.Close()

	q.sendMetrics(httpclient.New(), srv.URL)
}
