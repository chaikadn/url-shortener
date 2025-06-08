package middleware

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
)

type gzipWriter struct {
	http.ResponseWriter
	gzw *gzip.Writer
}

func newGzipWriter(hw http.ResponseWriter) (*gzipWriter, error) {
	gzw, err := gzip.NewWriterLevel(hw, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	return &gzipWriter{
		ResponseWriter: hw,
		gzw:            gzw,
	}, err
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.gzw.Write(b)
}

func (w *gzipWriter) WriteHeader(statusCode int) {
	w.Header().Set("Content-encoding", "gzip")
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipWriter) Close() error {
	return w.gzw.Close()
}

type gzipReader struct {
	io.ReadCloser
	gzr *gzip.Reader
}

func newGzipReader(body io.ReadCloser) (*gzipReader, error) {
	gzr, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}
	return &gzipReader{
		ReadCloser: body,
		gzr:        gzr,
	}, nil
}

func (r *gzipReader) Read(p []byte) (int, error) {
	return r.gzr.Read(p)
}

func (r *gzipReader) Close() error {
	return errors.Join(r.gzr.Close(), r.ReadCloser.Close())
}
