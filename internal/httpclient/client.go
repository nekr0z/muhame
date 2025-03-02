package httpclient

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

func New() *http.Client {
	return resty.New().
		SetRetryCount(3).
		GetClient()
}
