package middleware

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func Compression(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentEncoding := r.Header.Get("Content-Encoding")
			if strings.Contains(contentEncoding, "gzip") {
				gr, err := newGzipReader(r.Body)
				if err != nil {
					log.Error("failed to decode request", zap.Error(err))
					http.Error(w, "Failed to decompress request", http.StatusInternalServerError)
					return
				}
				defer gr.Close()
				r.Body = gr
			}

			acceptEncoding := r.Header.Get("Accept-Encoding")
			if strings.Contains(acceptEncoding, "gzip") {
				gw, err := newGzipWriter(w)
				if err != nil {
					// sending without comptrssion
					log.Error("failed to encode response", zap.Error(err))
				} else {
					w = gw
				}
				defer gw.Close()
			}

			next.ServeHTTP(w, r)
		})
	}
}
