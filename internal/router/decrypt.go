package router

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/nekr0z/muhame/internal/crypt"
)

func decrypt(privateKey *rsa.PrivateKey) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			body, err := crypt.Decrypt(b, privateKey)
			if err != nil {
				r.Body = io.NopCloser(bytes.NewReader(b))
			} else {
				r.Body = io.NopCloser(bytes.NewReader(body))
			}

			next.ServeHTTP(w, r)
		})
	}
}
