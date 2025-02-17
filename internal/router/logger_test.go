package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogger(t *testing.T) {
	respString := "It works!"
	uri := "http://example.com/foo"

	core, observed := observer.New(zap.DebugLevel)
	log := zap.New(core)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, respString)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", uri, nil)
	rr := httptest.NewRecorder()

	logger(log)(handler).ServeHTTP(rr, req)
	log.Sync()

	logs := observed.All()
	require.Len(t, logs, 1)

	msg := logs[0].Message
	require.Contains(t, msg, "method GET")
	require.Contains(t, msg, fmt.Sprintf("URI %s", uri))
	require.Contains(t, msg, "duration")
	require.Contains(t, msg, "status 200")
	require.Contains(t, msg, fmt.Sprintf("size %d", len(respString)))
}
