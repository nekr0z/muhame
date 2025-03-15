package router

import (
	"bytes"
	"io"
	"net/http"

	"github.com/nekr0z/muhame/internal/hash"
)

func checkSig(key string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			body := r.Body
			defer body.Close()

			bb, err := io.ReadAll(body)
			if err != nil {
				http.Error(w, "failed to read the body", http.StatusBadRequest)
				return
			}

			calculated := hash.Signature(bb, key)

			sig := r.Header.Get(hash.Header)

			if sig != "" && calculated != sig {
				http.Error(w, "signature does not match", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(bb))
			next.ServeHTTP(w, r)
		})
	}
}

func addSig(key string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var wb bytes.Buffer
			rw := &responseWriter{
				w:      &wb,
				header: make(http.Header),
			}

			next.ServeHTTP(rw, r)

			bb, err := io.ReadAll(&wb)
			if err != nil {
				http.Error(w, "failed to read the response", http.StatusInternalServerError)
				return
			}

			for k, values := range rw.header {
				for _, value := range values {
					w.Header().Add(k, value)
				}
			}

			w.Header().Set(hash.Header, hash.Signature(bb, key))

			if rw.code != 0 {
				w.WriteHeader(rw.code)
			}

			_, err = w.Write(bb)
			if err != nil {
				http.Error(w, "failed to write the response", http.StatusInternalServerError)
			}
		})
	}
}

var _ http.ResponseWriter = &responseWriter{}

type responseWriter struct {
	w      io.Writer
	code   int
	header http.Header
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.w.Write(b)
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
}
