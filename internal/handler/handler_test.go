package handler

// import (
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/chaikadn/url-shortener/internal/config"
// 	"github.com/chaikadn/url-shortener/internal/storage"
// 	"github.com/chaikadn/url-shortener/mocks"
// 	"github.com/go-chi/chi/v5"
// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/zap"
// )

// func TestHandler_HandlePingStorage(t *testing.T) {
// 	const endpoint = "/ping"
// 	tests := []struct {
// 		name            string
// 		mockSetup       func(*mocks.MockStorage)
// 		wantStatus      int
// 		wantContentType string
// 		wantBody        string
// 	}{
// 		{
// 			name: "successful ping",
// 			mockSetup: func(ms *mocks.MockStorage) {
// 				ms.EXPECT().Ping().Return(nil)
// 			},
// 			wantStatus:      http.StatusOK,
// 			wantContentType: "",
// 			wantBody:        "",
// 		},
// 		{
// 			name: "storage error",
// 			mockSetup: func(ms *mocks.MockStorage) {
// 				ms.EXPECT().Ping().Return(assert.AnError)
// 			},
// 			wantStatus:      http.StatusInternalServerError,
// 			wantContentType: "application/json",
// 			wantBody:        `{"error":"ping failed"}`,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			ms := mocks.NewMockStorage(ctrl)
// 			tt.mockSetup(ms)

// 			h := New(zap.NewNop(), &config.Config{}, nil, ms)
// 			router := chi.NewRouter()
// 			router.Get(endpoint, h.HandlePingStorage)

// 			r := httptest.NewRequest(http.MethodGet, endpoint, http.NoBody)
// 			w := httptest.NewRecorder()

// 			router.ServeHTTP(w, r)

// 			res := w.Result()
// 			body, _ := io.ReadAll(res.Body)
// 			defer res.Body.Close()

// 			assert.Equal(t, tt.wantStatus, res.StatusCode)
// 			assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
// 			if tt.wantBody != "" {
// 				assert.JSONEq(t, tt.wantBody, string(body))
// 			}
// 		})
// 	}
// }

// func TestHandler_HandleRedirect(t *testing.T) {
// 	tests := []struct {
// 		name            string
// 		mockSetup       func(*mocks.MockStorage)
// 		url             string
// 		wantStatus      int
// 		wantContentType string
// 		wantLocation    string
// 		wantBody        string
// 	}{
// 		{
// 			name: "successful redirect",
// 			mockSetup: func(ms *mocks.MockStorage) {
// 				ms.EXPECT().Get("test123").Return("http://redirect.com", nil)
// 			},
// 			url:             "http://localhost:8080/test123",
// 			wantStatus:      http.StatusTemporaryRedirect,
// 			wantContentType: "",
// 			wantLocation:    "http://redirect.com",
// 			wantBody:        "",
// 		},
// 		{
// 			name: "url not found",
// 			mockSetup: func(ms *mocks.MockStorage) {
// 				ms.EXPECT().Get("missing").Return("", storage.ErrNotFound)
// 			},
// 			url:             "http://localhost:8080/missing",
// 			wantStatus:      http.StatusNotFound,
// 			wantContentType: "application/json",
// 			wantLocation:    "",
// 			wantBody:        `{"error":"url not found"}`,
// 		},
// 		{
// 			name: "failed to redirect",
// 			mockSetup: func(ms *mocks.MockStorage) {
// 				ms.EXPECT().Get("fail").Return("", assert.AnError)
// 			},
// 			url:             "http://localhost:8080/fail",
// 			wantStatus:      http.StatusInternalServerError,
// 			wantContentType: "application/json",
// 			wantLocation:    "",
// 			wantBody:        `{"error":"failed to redirect"}`,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			ms := mocks.NewMockStorage(ctrl)
// 			tt.mockSetup(ms)

// 			h := New(zap.NewNop(), &config.Config{}, nil, ms)
// 			router := chi.NewRouter()
// 			router.Get("/{key}", h.HandleRedirect)

// 			r := httptest.NewRequest(http.MethodGet, tt.url, http.NoBody)
// 			w := httptest.NewRecorder()

// 			router.ServeHTTP(w, r)

// 			res := w.Result()

// 			body, _ := io.ReadAll(res.Body)
// 			defer res.Body.Close()

// 			assert.Equal(t, tt.wantStatus, res.StatusCode)
// 			if tt.wantLocation != "" {
// 				assert.Equal(t, tt.wantLocation, res.Header.Get("Location"))
// 			}
// 			if tt.wantContentType != "" {
// 				assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
// 			}
// 			if tt.wantBody != "" {
// 				assert.JSONEq(t, tt.wantBody, string(body))
// 			}
// 		})
// 	}
// }

// func TestHandler_HandleShorten(t *testing.T) {
// 	const endpoint = "/shorten"
// 	var config = &config.Config{
// 		Host: "localhost",
// 		Port: "8080",
// 	}

// 	tests := []struct {
// 		name            string
// 		mockSetup       func(*mocks.MockStorage, *mocks.MockGenerator)
// 		reqContentType  string
// 		reqBody         string
// 		wantStatus      int
// 		wantContentType string
// 		wantBody        string
// 	}{
// 		{
// 			name: "successful save",
// 			mockSetup: func(ms *mocks.MockStorage, mg *mocks.MockGenerator) {
// 				ms.EXPECT().Save("http://example.com", "key123").Return(nil)
// 				mg.EXPECT().Generate("http://example.com").Return("key123", nil)
// 			},
// 			reqContentType:  "application/json",
// 			reqBody:         `{"url":"http://example.com"}`,
// 			wantStatus:      http.StatusCreated,
// 			wantContentType: "application/json",
// 			wantBody:        `{"result":"http://localhost:8080/key123"}`,
// 		},
// 		{
// 			name: "storage error",
// 			mockSetup: func(ms *mocks.MockStorage, mg *mocks.MockGenerator) {
// 				ms.EXPECT().Save("http://error.com", "key123").Return(assert.AnError)
// 				mg.EXPECT().Generate("http://error.com").Return("key123", nil)
// 			},
// 			reqContentType:  "application/json",
// 			reqBody:         `{"url":"http://error.com"}`,
// 			wantStatus:      http.StatusInternalServerError,
// 			wantContentType: "application/json",
// 			wantBody:        `{"error":"failed to save url"}`,
// 		},
// 		{
// 			name: "generator error",
// 			mockSetup: func(ms *mocks.MockStorage, mg *mocks.MockGenerator) {
// 				mg.EXPECT().Generate("http://error.com").Return("", assert.AnError)
// 			},
// 			reqContentType:  "application/json",
// 			reqBody:         `{"url":"http://error.com"}`,
// 			wantStatus:      http.StatusInternalServerError,
// 			wantContentType: "application/json",
// 			wantBody:        `{"error":"failed to generate short url"}`,
// 		},
// 		{
// 			name:            "invalid json",
// 			mockSetup:       func(ms *mocks.MockStorage, mg *mocks.MockGenerator) {},
// 			reqContentType:  "application/json",
// 			reqBody:         `{"url": "invalid-url`,
// 			wantStatus:      http.StatusBadRequest,
// 			wantContentType: "application/json",
// 			wantBody:        `{"error":"failed to decode json"}`,
// 		},
// 		{
// 			name:            "wrong content type",
// 			mockSetup:       func(ms *mocks.MockStorage, mg *mocks.MockGenerator) {},
// 			reqContentType:  "text/plain",
// 			reqBody:         `any`,
// 			wantStatus:      http.StatusBadRequest,
// 			wantContentType: "application/json",
// 			wantBody:        `{"error":"failed to shorten url"}`,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			ms := mocks.NewMockStorage(ctrl)
// 			mg := mocks.NewMockGenerator(ctrl)
// 			tt.mockSetup(ms, mg)

// 			h := New(zap.NewNop(), config, mg, ms)
// 			router := chi.NewRouter()
// 			router.Post(endpoint, h.HandleShorten)

// 			r := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(tt.reqBody))
// 			r.Header.Set("Content-Type", tt.reqContentType)
// 			w := httptest.NewRecorder()

// 			router.ServeHTTP(w, r)

// 			res := w.Result()

// 			body, _ := io.ReadAll(res.Body)
// 			defer res.Body.Close()

// 			assert.Equal(t, tt.wantStatus, res.StatusCode)
// 			assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
// 			if tt.wantBody != "" {
// 				assert.JSONEq(t, tt.wantBody, string(body))
// 			}
// 		})
// 	}
// }
