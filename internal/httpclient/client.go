// Package httpclient provides a client for HTTP requests.
package httpclient

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/nekr0z/muhame/internal/crypt"
	"github.com/nekr0z/muhame/internal/hash"
)

const (
	defaultRetries = 3

	HeaderRealIP = "X-Real-IP"
)

// Client is a client for HTTP requests.
type Client struct {
	c      *http.Client
	key    string
	pubKey *rsa.PublicKey
	ip     string
}

// New returns a new Client.
func New() Client {
	return Client{
		c: resty.New().
			SetRetryCount(defaultRetries).
			GetClient(),
		ip: getLocalIP(),
	}
}

// WithKey sets the signing key for the client.
func (c Client) WithKey(key string) Client {
	c.key = key
	return c
}

// WithCrypto sets the public key for the client.
func (c Client) WithCrypto(key *rsa.PublicKey) Client {
	c.pubKey = key
	return c
}

// getLocalIP returns the first non-loopback address of the machine.
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			return ipnet.IP.String()
		}
	}

	return ""
}

// Send sends a request to the given endpoint.
func (c Client) Send(msg []byte, endpoint string) (int, error) {
	if c.pubKey != nil {
		ciphertext, err := crypt.Encrypt(msg, c.pubKey)
		if err != nil {
			return 0, fmt.Errorf("failed to encrypt message: %w", err)
		}
		msg = ciphertext
	}

	b := bytes.NewBuffer(msg)
	req, err := http.NewRequest(http.MethodPost, endpoint, b)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if c.key != "" {
		sig := hash.Signature(msg, c.key)
		req.Header.Set(hash.Header, sig)
	}

	if c.ip != "" {
		req.Header.Set(HeaderRealIP, c.ip)
	}

	resp, err := c.c.Do(req)
	if err != nil {
		return 0, err
	}

	if resp == nil {
		return 0, fmt.Errorf("nil response")
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	err = resp.Body.Close()

	return resp.StatusCode, err
}
