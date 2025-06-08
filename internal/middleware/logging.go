package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type logResponseWriter struct {
	http.ResponseWriter
	status int
	// header http.Header
	body *bytes.Buffer
	size int
}

func (w *logResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *logResponseWriter) Write(b []byte) (int, error) {
	n, err1 := w.ResponseWriter.Write(b)
	_, err2 := w.body.Write(b)
	w.size += n
	return w.size, errors.Join(err1, err2)
}

func Logging(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()

				lw := &logResponseWriter{ResponseWriter: w, status: http.StatusOK, body: &bytes.Buffer{}}

				next.ServeHTTP(lw, r)
				duration := time.Since(start)

				log.Info("Http request",
					zap.String("method", r.Method),
					zap.String("uri", r.RequestURI),
					zap.Int("status", lw.status),
					zap.Int("size", lw.size),
					zap.String("duration", duration.String()),
				)
			},
		)
	}
}
