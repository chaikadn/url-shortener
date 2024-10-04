package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestHandler_postURL(t *testing.T) {
	storage := memory.NewStorage()
	handler := New(storage, &config.Config{})

	r := chi.NewRouter()
	r.Post("/", handler.postURL)
	server := httptest.NewServer(r)
	defer server.Close()

	// TODO: mock random url generator
	tests := []struct {
		name         string
		method       string
		path         string
		body         string
		expectedCode int
		// expectedBody     string
	}{
		{
			name:         "positive test",
			method:       http.MethodPost,
			path:         "/",
			body:         "https://practicum.yandex.ru/",
			expectedCode: http.StatusCreated,
			// expectedBody: "fixedBody",
		},
		{
			name:         "empty body test",
			method:       http.MethodPost,
			path:         "/",
			body:         "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid url test",
			method:       http.MethodPost,
			path:         "/",
			body:         "invalid-url",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "wrong method test",
			method:       http.MethodGet,
			path:         "/",
			body:         "any-url",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.method
			req.URL = server.URL + tt.path
			req.Body = tt.body

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			// assert.Equal(t, tt.expectedBody, string(resp.Body()))
		})
	}
}

func TestHandler_getURL(t *testing.T) {
	storage := memory.NewStorage()
	handler := New(storage, &config.Config{})

	r := chi.NewRouter()
	r.Get("/{id}", handler.getURL)
	server := httptest.NewServer(r)
	defer server.Close()

	id, _ := storage.Add("https://practicum.yandex.ru/")

	tests := []struct {
		name             string
		method           string
		path             string
		expectedCode     int
		expectedBody     string
		expectedLocation string
	}{
		{
			name:             "positive test",
			method:           http.MethodGet,
			path:             "/" + id,
			expectedCode:     http.StatusTemporaryRedirect,
			expectedLocation: "https://practicum.yandex.ru/",
		},
		{
			name:         "non-existent URL",
			method:       http.MethodGet,
			path:         "/invalid-id",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Unable to get URL\n",
		},
		{
			name:         "wrong method",
			method:       http.MethodPost,
			path:         "/any-id",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().SetRedirectPolicy(resty.NoRedirectPolicy()).R() // disable autoredirect
			req.Method = tt.method
			req.URL = server.URL + tt.path

			resp, err := req.Send()

			// ignore redirect error
			if errors.Unwrap(err) == resty.ErrAutoRedirectDisabled {
				err = nil
			}
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			assert.Equal(t, tt.expectedLocation, resp.Header().Get("Location"))
			assert.Equal(t, tt.expectedBody, string(resp.Body()))
		})
	}
}
