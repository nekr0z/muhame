package router_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/require"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/router"
)

const (
	trustedSubnet = "192.168.0.0/24"
	trustedIP     = "192.168.0.50"
	untrustedIP   = "192.168.4.50"

	badSubnet = "whatever"
	badIP     = ""

	name   = "test"
	metric = metrics.Counter(1)
)

func TestTrusted(t *testing.T) {
	t.Parallel()

	st := &mockStorage{t, name, metric}
	log := zap.NewNop()
	r := router.New(log, st, "", nil, trustedSubnet)

	srv := httptest.NewServer(r)
	defer srv.Close()

	t.Run("trusted", func(t *testing.T) {
		url := fmt.Sprintf("%s/update/counter/%s/%s", srv.URL, name, metric)

		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("X-Real-IP", trustedIP)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("untrusted", func(t *testing.T) {
		url := fmt.Sprintf("%s/update/counter/%s/%s", srv.URL, name, metric)

		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("X-Real-IP", untrustedIP)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("bad IP", func(t *testing.T) {
		url := fmt.Sprintf("%s/update/counter/%s/%s", srv.URL, name, metric)

		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("X-Real-IP", badIP)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestTrusted_BadSubnet(t *testing.T) {
	t.Parallel()

	st := &mockStorage{t, name, metric}
	log := zap.NewNop()
	r := router.New(log, st, "", nil, badSubnet)

	srv := httptest.NewServer(r)
	defer srv.Close()

	url := fmt.Sprintf("%s/update/counter/%s/%s", srv.URL, name, metric)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("X-Real-IP", trustedIP)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}
