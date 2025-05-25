package router

import (
	"compress/gzip"
	"net/http"
	"slices"
)

func acceptGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(r.Header.Values("Content-Encoding"), "gzip") {
			body, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Body = body
		}
		next.ServeHTTP(w, r)
	})
}

func respondGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(r.Header.Values("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")

			zw, err := gzip.NewWriterLevel(w, gzip.HuffmanOnly)
			if err != nil {
				panic(err)
			}
			defer func() {
				err = zw.Close()
				if err != nil {
					panic(err)
				}
			}()

			wr := &gzipWriter{w, zw}

			next.ServeHTTP(wr, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	z *gzip.Writer
}

// Write implements the io.Writer interface.
func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.z.Write(b)
}
