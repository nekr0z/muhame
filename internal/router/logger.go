package router

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func logger(log *zap.Logger) middleware {
	sugar := *log.Sugar()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   &responseData{},
			}

			next.ServeHTTP(&lw, r)

			duration := time.Since(start)

			sugar.Infoln(
				"URI", uri,
				"method", method,
				"duration", duration,
				"status", lw.responseData.status,
				"size", lw.responseData.size,
			)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write implements the io.Writer interface.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if r.responseData.status == 0 {
		r.WriteHeader(http.StatusOK)
	}

	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader implements the http.ResponseWriter interface.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

type responseData struct {
	status int
	size   int
}
