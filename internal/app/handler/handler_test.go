package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ShortenURL(t *testing.T) {
	s := memory.NewStorage()
	h := New(s)

	tests := []struct {
		name     string
		method   string
		body     string
		wantCode int
		wantBody string
	}{
		{
			name:     "positive test",
			method:   http.MethodPost,
			body:     "https://www.youtube.com",
			wantCode: http.StatusCreated,
			wantBody: "", // random URL
		},
		{
			name:     "wrong method test",
			method:   http.MethodGet,
			body:     "",
			wantCode: http.StatusBadRequest,
			wantBody: "method not alowed\n",
		},
		{
			name:     "empty request body test",
			method:   http.MethodPost,
			body:     "",
			wantCode: http.StatusBadRequest,
			wantBody: "request body can not be empty\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.ShortenURL(w, r)

			assert.Equal(t, tt.wantCode, w.Code, "status code not match")
			if tt.wantBody != "" {
				assert.Equal(t, w.Body.String(), tt.wantBody)
			}
			if w.Code == http.StatusCreated {
				url, _ := strings.CutPrefix(w.Body.String(), "http://localhost:8080/")
				assert.Equal(t, tt.body, s.Storage[url])
			}
		})
	}
}

func TestHandler_GetURL(t *testing.T) {
	s := memory.NewStorage()
	s.Storage["test"] = "https://www.youtube.com"
	h := New(s)

	tests := []struct {
		name         string
		method       string
		path         string
		wantCode     int
		wantBody     string
		wantLocation string
	}{
		{
			name:         "positive test",
			method:       http.MethodGet,
			path:         "/test",
			wantCode:     http.StatusTemporaryRedirect,
			wantLocation: "https://www.youtube.com/",
		},
		{
			name:     "wrong method test",
			method:   http.MethodPost,
			path:     "/test",
			wantCode: http.StatusBadRequest,
			wantBody: "method not alowed\n",
		},
		{
			name:     "wrong url test",
			method:   http.MethodGet,
			path:     "/noTest",
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			h.GetURL(w, r)

			assert.Equal(t, tt.wantCode, w.Code, "status code not match")
			if tt.wantBody != "" {
				assert.Equal(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
