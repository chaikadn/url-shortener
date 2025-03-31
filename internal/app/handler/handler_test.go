package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/storage"
	"github.com/chaikadn/url-shortener/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TODO: сделать мок для RandStr(length int) string
func TestHandler_handlePing(t *testing.T) {
	tests := []struct {
		name       string
		mockSetup  func(*mocks.MockStorage)
		wantStatus int
		wantBody   string
	}{
		{
			name: "successful ping",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().Ping(gomock.Any()).Return(nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name: "storage unavailable",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().Ping(gomock.Any()).Return(assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "failed to connect to storage\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mst := mocks.NewMockStorage(ctrl)
			tt.mockSetup(mst)

			h := &Handler{
				storage: mst,
				config:  nil,
			}

			r := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			h.handlePing(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			if tt.wantBody != "" {
				body, _ := io.ReadAll(res.Body)
				assert.Equal(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestHandler_handleRedirect(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mocks.MockStorage)
		method       string
		uri          string
		wantStatus   int
		wantLocation string
		wantBody     string
	}{
		{
			name: "succsessful get",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().
					GetOriginal(gomock.Any(), "test").
					Return(&storage.URLEntry{ShortURL: "test", OriginalURL: "https://practicum.yandex.ru"}, nil)
			},
			method:       http.MethodGet,
			uri:          "/test",
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "https://practicum.yandex.ru",
			wantBody:     "",
		},
		{
			name: "wrong short url",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().GetOriginal(gomock.Any(), "wrong").Return(nil, storage.ErrNotFound)
			},
			method:       http.MethodGet,
			uri:          "/wrong",
			wantStatus:   http.StatusNotFound,
			wantLocation: "",
			wantBody:     "url not found\n",
		},
		{
			name: "storage error",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().GetOriginal(gomock.Any(), "error").Return(nil, assert.AnError)
			},
			method:       http.MethodGet,
			uri:          "/error",
			wantStatus:   http.StatusInternalServerError,
			wantLocation: "",
			wantBody:     "failed to get url\n",
		},
		{
			name:         "wrong method",
			mockSetup:    func(ms *mocks.MockStorage) {},
			method:       http.MethodPost,
			uri:          "/any",
			wantStatus:   http.StatusMethodNotAllowed,
			wantLocation: "",
			wantBody:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mst := mocks.NewMockStorage(ctrl)
			tt.mockSetup(mst)

			h := &Handler{
				storage: mst,
				config:  nil,
			}

			r := httptest.NewRequest(tt.method, tt.uri, nil)
			w := httptest.NewRecorder()

			h.Route().ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			assert.Equal(t, tt.wantLocation, res.Header.Get("Location"))
			if tt.wantBody != "" {
				body, _ := io.ReadAll(res.Body)
				assert.Equal(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestHandler_handleShortenText(t *testing.T) {
	tests := []struct {
		name       string
		mockSetup  func(*mocks.MockStorage)
		method     string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name: "succsessful",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().
					Add(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			method:     http.MethodPost,
			body:       "https://practicum.yandex.ru",
			wantStatus: http.StatusCreated,
			wantBody:   "http://localhost:8080/.*",
		},
		{
			name: "storage error",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().
					Add(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			method:     http.MethodPost,
			body:       "https://any.ru",
			wantStatus: http.StatusBadRequest,
			wantBody:   "failed to shorten url",
		},
		{
			name:       "invalid url",
			mockSetup:  func(ms *mocks.MockStorage) {},
			method:     http.MethodPost,
			body:       "invalid-url",
			wantStatus: http.StatusBadRequest,
			wantBody:   "failed to shorten url",
		},
		{
			name:       "wrong method",
			mockSetup:  func(ms *mocks.MockStorage) {},
			method:     http.MethodGet,
			body:       "any-url",
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mst := mocks.NewMockStorage(ctrl)
			tt.mockSetup(mst)

			h := &Handler{
				storage: mst,
				config:  &config.Config{BaseURL: "http://localhost:8080"},
			}

			r := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.Route().ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			if tt.wantBody != "" {
				body, _ := io.ReadAll(res.Body)
				assert.Regexp(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestHandler_handleShortenJSON(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(*mocks.MockStorage)
		method          string
		body            string
		wantContentType string
		wantStatus      int
		wantBody        string
	}{
		{
			name: "succsessful",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil)
			},
			method:          http.MethodPost,
			body:            `{"url": "https://practicum.yandex.ru"}`,
			wantContentType: "application/json",
			wantStatus:      http.StatusCreated,
			wantBody:        `{"result":"http://localhost:8080/.*"}`,
		},
		{
			name: "storage error",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().Add(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			method:     http.MethodPost,
			body:       `{"url": "https://error.ru"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "failed to shorten url",
		},
		{
			name:       "invalid url",
			mockSetup:  func(ms *mocks.MockStorage) {},
			method:     http.MethodPost,
			body:       `{"url": "invalid-url"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "failed to shorten url",
		},
		{
			name:       "invalid json",
			mockSetup:  func(ms *mocks.MockStorage) {},
			method:     http.MethodPost,
			body:       `invalid-json`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "failed to decode json body",
		},
		{
			name:       "wrong method",
			mockSetup:  func(ms *mocks.MockStorage) {},
			method:     http.MethodGet,
			body:       "any-url",
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mst := mocks.NewMockStorage(ctrl)
			tt.mockSetup(mst)

			h := &Handler{
				storage: mst,
				config:  &config.Config{BaseURL: "http://localhost:8080"},
			}

			r := httptest.NewRequest(tt.method, "/api/shorten", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.Route().ServeHTTP(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			if res.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.wantContentType, res.Header.Get("Content-Type"))
			}
			if tt.wantBody != "" {
				body, _ := io.ReadAll(res.Body)
				assert.Regexp(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestHandler_shorten(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockStorage)
		originalURL string
		wantErr     error
	}{
		{
			name: "succsess",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil)
			},
			originalURL: "https://practicum.yandex.ru",
			wantErr:     nil,
		},
		{
			name: "storage error",
			mockSetup: func(ms *mocks.MockStorage) {
				ms.EXPECT().Add(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			originalURL: "https://error.ru",
			wantErr:     assert.AnError,
		},
		{
			name:        "invalid url",
			mockSetup:   func(ms *mocks.MockStorage) {},
			originalURL: "invalid-url",
			wantErr:     fmt.Errorf("invalid url 'invalid-url'"),
		},
		{
			name:        "empty url",
			mockSetup:   func(ms *mocks.MockStorage) {},
			originalURL: "",
			wantErr:     fmt.Errorf("invalid url ''"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mst := mocks.NewMockStorage(ctrl)
			tt.mockSetup(mst)

			h := &Handler{
				storage: mst,
				config:  &config.Config{BaseURL: "http://localhost:8080"},
			}

			shortURL, err := h.shorten(context.Background(), tt.originalURL)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
			}

			if err != nil {
				assert.Equal(t, "", shortURL)
			} else {
				assert.Regexp(t, fmt.Sprintf(`^%s/[a-zA-Z0-9]{8}$`, h.config.BaseURL), shortURL)
			}
		})
	}
}
