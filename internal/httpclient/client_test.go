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
	c.Send([]byte(msg), srv.URL)
}
