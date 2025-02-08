package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	gzipw *gzip.Writer
}

func newGzipWriter(w http.ResponseWriter) (*gzipWriter, error) {
	gw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	return &gzipWriter{
		ResponseWriter: w,
		gzipw:          gw,
	}, nil
}

func (gw *gzipWriter) Write(b []byte) (int, error) {
	return gw.gzipw.Write(b)
}

func (gw *gzipWriter) WriteHeader(statusCode int) {
	gw.Header().Del("Content-Length")
	gw.Header().Set("Content-Encoding", "gzip")
	gw.ResponseWriter.WriteHeader(statusCode)
}

func (gw *gzipWriter) Close() error {
	return gw.gzipw.Close()
}

type gzipReader struct {
	io.ReadCloser
	gzipr *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{
		ReadCloser: r,
		gzipr:      gr,
	}, nil
}

func (gr *gzipReader) Read(p []byte) (n int, err error) {
	return gr.gzipr.Read(p)
}

func (gr *gzipReader) Close() error {
	if err := gr.gzipr.Close(); err != nil {
		return err
	}
	return gr.ReadCloser.Close()
}

func WithGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// проверяем, поддерживает ли клиент сжатие gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := parseHeader(acceptEncoding, "gzip")
		if supportsGzip {
			gw, err := newGzipWriter(w)
			if err != nil {
				http.Error(w, "cannot encode gzip", http.StatusInternalServerError)
				return
			}
			w = gw
			defer gw.Close()
		}

		//проверяем, отправил ли клиент сжатые в gzip данные
		contentEncoding := r.Header.Get("Content-Encoding")
		// исправить
		sendsGzip := parseHeader(contentEncoding, "gzip")
		if sendsGzip {
			gr, err := newGzipReader(r.Body)
			if err != nil {
				http.Error(w, "cannot decode gzip", http.StatusInternalServerError)
				return
			}
			r.Body = gr
			defer gr.Close()
		}

		next.ServeHTTP(w, r)
	})
}

// доработать (может быть значенрие, например zip;q=0.8)
func parseHeader(header, value string) bool {
	values := strings.Split(header, ",")
	for _, val := range values {
		if val == strings.TrimSpace(value) {
			return true
		}
	}
	return false
}
