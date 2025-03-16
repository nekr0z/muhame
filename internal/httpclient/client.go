package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/nekr0z/muhame/internal/hash"
)

const defaultRetries = 3

type Client struct {
	c   *http.Client
	key string
}

func New() Client {
	return Client{
		c: resty.New().
			SetRetryCount(defaultRetries).
			GetClient(),
	}
}

func (c Client) WithKey(key string) Client {
	c.key = key
	return c
}

func (c Client) Send(msg []byte, endpoint string) (int, error) {
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

	resp, err := c.c.Do(req)
	if err != nil {
		return 0, err
	}

	if resp == nil {
		return 0, fmt.Errorf("nil response")
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return resp.StatusCode, nil
}
