package httpclient_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/hash"
	"github.com/nekr0z/muhame/internal/httpclient"
)

func TestSend_Signed(t *testing.T) {
	msg := "test message"
	key := "testkey"

	sig := sha256.Sum256([]byte(msg + key))
	want := hex.EncodeToString(sig[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get(hash.Header)
		assert.Equal(t, want, got)
		m, _ := io.ReadAll(r.Body)
		assert.Equal(t, msg, string(m))
	}))

	c := httpclient.New().WithKey(key)
	_, err := c.Send([]byte(msg), srv.URL)
	assert.NoError(t, err)
}

func TestSend_RealIP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-IP")
		assert.NotEmpty(t, ip, "X-Real-IP header should not be empty")
		assert.NotEqual(t, "127.0.0.1", ip, "X-Real-IP should not be loopback")
		assert.NotEqual(t, "::1", ip, "X-Real-IP should not be IPv6 loopback")

		m, _ := io.ReadAll(r.Body)
		assert.Equal(t, "test message", string(m))
	}))

	c := httpclient.New()
	_, err := c.Send([]byte("test message"), srv.URL)
	assert.NoError(t, err)
}
