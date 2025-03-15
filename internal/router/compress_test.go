package router_test

import (
	"bytes"
	"compress/gzip"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/router"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	log = zap.NewNop()

	testMetric     = metrics.Gauge(0.5)
	testMetricName = "test"

	in = `{"id": "test", "type": "gauge", "value": 0.5}`
)

func TestNew_GzippedRequest(t *testing.T) {
	t.Parallel()

	comp := compress(t, []byte(in))

	st := mockStorage{
		t:    t,
		name: testMetricName,
		m:    testMetric,
	}

	r := router.New(log, st, "")

	req := httptest.NewRequest("POST", "/value/", &comp)
	req.Header.Set("Content-Encoding", "gzip")

	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"id": "test", "type": "gauge", "value": 0.5}`, res.Body.String())
}

func TestNew_GzippedResponse(t *testing.T) {
	t.Parallel()

	st := mockStorage{
		t:    t,
		name: testMetricName,
		m:    testMetric,
	}

	r := router.New(log, st, "")

	req := httptest.NewRequest("POST", "/value/", strings.NewReader(in))
	req.Header.Set("Accept-Encoding", "gzip")

	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "gzip", res.Header().Get("Content-Encoding"))
	assert.JSONEq(t, `{"id": "test", "type": "gauge", "value": 0.5}`, uncompress(t, res.Body))
}

func compress(t *testing.T, b []byte) bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	assert.NoError(t, err)

	_, err = w.Write(b)
	assert.NoError(t, err)

	err = w.Close()
	assert.NoError(t, err)

	return buf
}

func uncompress(t *testing.T, b *bytes.Buffer) string {
	t.Helper()

	r, err := gzip.NewReader(b)
	assert.NoError(t, err)

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	assert.NoError(t, err)

	return buf.String()
}
