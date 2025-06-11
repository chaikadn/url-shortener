package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chaikadn/url-shortener/internal/config"
	"github.com/chaikadn/url-shortener/internal/middleware"
	"github.com/chaikadn/url-shortener/internal/model"
	"github.com/chaikadn/url-shortener/internal/service"
	"github.com/chaikadn/url-shortener/internal/storage"
	"github.com/chaikadn/url-shortener/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandler_HandlePingStorage(t *testing.T) {
	const endpoint = "/ping"
	tests := []struct {
		name            string
		mockSetup       func(*mocks.MockShortenerService)
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "successful ping",
			mockSetup: func(mss *mocks.MockShortenerService) {
				mss.EXPECT().PingStorage().Return(nil)
			},
			wantStatus:      http.StatusOK,
			wantContentType: "",
			wantBody:        "",
		},
		{
			name: "storage error",
			mockSetup: func(mss *mocks.MockShortenerService) {
				mss.EXPECT().PingStorage().Return(assert.AnError)
			},
			wantStatus:      http.StatusInternalServerError,
			wantContentType: "application/json",
			wantBody:        `{"error":"ping failed"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mss := mocks.NewMockShortenerService(ctrl)
			tt.mockSetup(mss)

			hnd := New(zap.NewNop(), &config.Config{}, mss, nil)
			router := chi.NewRouter()
			router.Get(endpoint, hnd.HandlePingStorage)

			r := httptest.NewRequest(http.MethodGet, endpoint, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			res := w.Result()
			body, _ := io.ReadAll(res.Body)
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestHandler_HandleLogin(t *testing.T) {
	const endpoint = "/api/login"
	var config = &config.Config{
		Host:      "localhost",
		Port:      "8080",
		JWTsecret: "some-secret",
	}

	tests := []struct {
		name       string
		mockSetup  func(*mocks.MockUserService)
		reqBody    string
		wantStatus int
		wantBody   string
	}{
		{
			name: "succsessful login",
			mockSetup: func(mus *mocks.MockUserService) {
				mus.EXPECT().Login("user1", "password1", "some-secret").Return("some-token", nil)
			},
			reqBody:    `{"username":"user1", "password": "password1"}`,
			wantStatus: http.StatusOK,
			wantBody:   `{"token":"some-token"}`,
		},
		{
			name: "wrong username or password",
			mockSetup: func(mus *mocks.MockUserService) {
				mus.EXPECT().Login("user2", "password2", "some-secret").Return("", service.ErrUnauthorized)
			},
			reqBody:    `{"username":"user2", "password": "password2"}`,
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"wrong username or password"}`,
		},
		{
			name: "storage error",
			mockSetup: func(mus *mocks.MockUserService) {
				mus.EXPECT().Login("user3", "password3", "some-secret").Return("", assert.AnError)
			},
			reqBody:    `{"username":"user3", "password": "password3"}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"failed to login"}`,
		},
		{
			name:       "decode error",
			mockSetup:  func(*mocks.MockUserService) {},
			reqBody:    `{bad json"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"failed to decode json"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mus := mocks.NewMockUserService(ctrl)
			tt.mockSetup(mus)

			hnd := New(zap.NewNop(), config, nil, mus)
			router := chi.NewRouter()
			router.Post(endpoint, hnd.HandleLogin)

			r := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(tt.reqBody))
			// r.Header.Set("Content-Type", tt.reqContentType)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			res := w.Result()

			body, _ := io.ReadAll(res.Body)
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, string(body))
			}

		})
	}
}

func TestHandler_HandleRedirect(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(*mocks.MockShortenerService)
		url             string
		wantStatus      int
		wantContentType string
		wantLocation    string
		wantBody        string
	}{
		{
			name: "successful redirect",
			mockSetup: func(mss *mocks.MockShortenerService) {
				mss.EXPECT().GetOriginalURL("test123").Return("http://redirect.com", nil)
			},
			url:             "http://localhost:8080/test123",
			wantStatus:      http.StatusTemporaryRedirect,
			wantContentType: "",
			wantLocation:    "http://redirect.com",
			wantBody:        "",
		},
		{
			name: "url not found",
			mockSetup: func(mss *mocks.MockShortenerService) {
				mss.EXPECT().GetOriginalURL("missing").Return("", storage.ErrURLNotFound)
			},
			url:             "http://localhost:8080/missing",
			wantStatus:      http.StatusNotFound,
			wantContentType: "application/json",
			wantLocation:    "",
			wantBody:        `{"error":"url not found"}`,
		},
		{
			name: "failed to redirect",
			mockSetup: func(mss *mocks.MockShortenerService) {
				mss.EXPECT().GetOriginalURL("fail").Return("", assert.AnError)
			},
			url:             "http://localhost:8080/fail",
			wantStatus:      http.StatusInternalServerError,
			wantContentType: "application/json",
			wantLocation:    "",
			wantBody:        `{"error":"failed to redirect"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mss := mocks.NewMockShortenerService(ctrl)
			tt.mockSetup(mss)

			hnd := New(zap.NewNop(), &config.Config{}, mss, nil)
			router := chi.NewRouter()
			router.Get("/{key}", hnd.HandleRedirect)

			r := httptest.NewRequest(http.MethodGet, tt.url, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			res := w.Result()

			body, _ := io.ReadAll(res.Body)
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			if tt.wantLocation != "" {
				assert.Equal(t, tt.wantLocation, res.Header.Get("Location"))
			}
			if tt.wantContentType != "" {
				assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
			}
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestHandler_HandleRegister(t *testing.T) {
	const endpoint = "/api/register"

	tests := []struct {
		name       string
		mockSetup  func(*mocks.MockUserService)
		reqBody    string
		wantStatus int
		wantBody   string
	}{
		{
			name: "succsessful registration",
			mockSetup: func(mus *mocks.MockUserService) {
				mus.EXPECT().Register("user1", "password1").Return(1, nil)
				mus.EXPECT().GetUser(1).Return(&model.User{ID: 1, Name: "user1", Role: "user"}, nil)
			},
			reqBody:    `{"username":"user1", "password": "password1"}`,
			wantStatus: http.StatusCreated,
			wantBody:   `{"id": 1, "name": "user1", "role": "user"}`,
		},
		{
			name: "user already exists",
			mockSetup: func(mus *mocks.MockUserService) {
				mus.EXPECT().Register("user2", "password2").Return(-1, storage.ErrUserAlreadyExists)
			},
			reqBody:    `{"username":"user2", "password": "password2"}`,
			wantStatus: http.StatusConflict,
			wantBody:   `{"error":"user already exists"}`,
		},
		{
			name: "user service error #1",
			mockSetup: func(mus *mocks.MockUserService) {
				mus.EXPECT().Register("user3", "password3").Return(-1, assert.AnError)
			},
			reqBody:    `{"username":"user3", "password": "password3"}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"failed to register"}`,
		},
		{
			name: "user service error #2",
			mockSetup: func(mus *mocks.MockUserService) {
				mus.EXPECT().Register("user1", "password1").Return(1, nil)
				mus.EXPECT().GetUser(1).Return(&model.User{}, assert.AnError)
			},
			reqBody:    `{"username":"user1", "password": "password1"}`,
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"failed to get user data"}`,
		},
		{
			name:       "decode error",
			mockSetup:  func(*mocks.MockUserService) {},
			reqBody:    `{bad json"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"failed to decode json"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mus := mocks.NewMockUserService(ctrl)
			tt.mockSetup(mus)

			hnd := New(zap.NewNop(), &config.Config{}, nil, mus)
			router := chi.NewRouter()
			router.Post(endpoint, hnd.HandleRegister)

			r := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(tt.reqBody))
			// r.Header.Set("Content-Type", tt.reqContentType)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			res := w.Result()

			body, _ := io.ReadAll(res.Body)
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, string(body))
			}

		})
	}
}

func TestHandler_HandleShorten(t *testing.T) {
	const endpoint = "/api/shorten"

	var config = &config.Config{
		Host:      "localhost",
		Port:      "8080",
		JWTsecret: "some-secret",
	}

	tests := []struct {
		name            string
		mockSetup       func(*mocks.MockShortenerService)
		userID          any
		reqBody         string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name: "successful save",
			mockSetup: func(mss *mocks.MockShortenerService) {
				mss.EXPECT().CreateShortURL("http://example.com", 1).Return("key123", nil)
			},
			userID:          1,
			reqBody:         `{"url":"http://example.com"}`,
			wantStatus:      http.StatusCreated,
			wantContentType: "application/json",
			wantBody:        `{"result":"http://localhost:8080/key123"}`,
		},
		{
			name: "service error",
			mockSetup: func(mss *mocks.MockShortenerService) {
				mss.EXPECT().CreateShortURL("http://error.com", 1).Return("", assert.AnError)
			},
			userID:          1,
			reqBody:         `{"url":"http://error.com"}`,
			wantStatus:      http.StatusInternalServerError,
			wantContentType: "application/json",
			wantBody:        `{"error":"failed to shorten url"}`,
		},
		{
			name:            "authorization error",
			mockSetup:       func(mss *mocks.MockShortenerService) {},
			userID:          "wrong-id",
			reqBody:         `any`,
			wantStatus:      http.StatusUnauthorized,
			wantContentType: "application/json",
			wantBody:        `{"error":"authorization failed"}`,
		},
		{
			name:            "invalid json",
			mockSetup:       func(mss *mocks.MockShortenerService) {},
			userID:          1,
			reqBody:         `{"url": "invalid-url`,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "application/json",
			wantBody:        `{"error":"failed to decode json"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mss := mocks.NewMockShortenerService(ctrl)
			tt.mockSetup(mss)

			hnd := New(zap.NewNop(), config, mss, nil)
			router := chi.NewRouter()
			router.Post(endpoint, hnd.HandleShorten)

			r := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			ctx := context.WithValue(r.Context(), middleware.UserIDkey, tt.userID)
			router.ServeHTTP(w, r.WithContext(ctx))

			res := w.Result()

			body, _ := io.ReadAll(res.Body)
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestHandler_HandleUserURLs(t *testing.T) {
	const endpoint = "/api/urls"

	var config = &config.Config{
		Host:      "localhost",
		Port:      "8080",
		JWTsecret: "some-secret",
	}

	tests := []struct {
		name       string
		mockSetup  func(*mocks.MockShortenerService)
		userID     any
		wantStatus int
		wantBody   string
	}{
		{
			name: "successful non empty",
			mockSetup: func(mss *mocks.MockShortenerService) {
				mss.EXPECT().
					GetUserURLs(1).
					Return([]*model.URLEntry{
						{
							ID:          1,
							Key:         "key1",
							OriginalURL: "http://example1.com",
						},
						{
							ID:          2,
							Key:         "key2",
							OriginalURL: "http://example2.com",
						},
					}, nil)
			},
			userID:     1,
			wantStatus: http.StatusOK,
			wantBody: `
				{
					"user_id": 1,
					"urls":
					[
						{"original": "http://example1.com", "short": "http://localhost:8080/key1"},
						{"original": "http://example2.com", "short": "http://localhost:8080/key2"}
					]
				}`,
		},
		{
			name: "successful empty",
			mockSetup: func(mss *mocks.MockShortenerService) {
				mss.EXPECT().GetUserURLs(1).Return([]*model.URLEntry{}, nil)
			},
			userID:     1,
			wantStatus: http.StatusNoContent,
		},
		{
			name: "service error",
			mockSetup: func(mss *mocks.MockShortenerService) {
				mss.EXPECT().GetUserURLs(1).Return([]*model.URLEntry{}, assert.AnError)
			},
			userID:     1,
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"failed to get user urls"}`,
		},
		{
			name:       "authorization error",
			mockSetup:  func(mss *mocks.MockShortenerService) {},
			userID:     "wrong-id",
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"authorization failed"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mss := mocks.NewMockShortenerService(ctrl)
			tt.mockSetup(mss)

			hnd := New(zap.NewNop(), config, mss, nil)
			router := chi.NewRouter()
			router.Post(endpoint, hnd.HandleUserURLs)

			r := httptest.NewRequest(http.MethodPost, endpoint, nil)
			w := httptest.NewRecorder()

			ctx := context.WithValue(r.Context(), middleware.UserIDkey, tt.userID)
			router.ServeHTTP(w, r.WithContext(ctx))

			res := w.Result()

			body, _ := io.ReadAll(res.Body)
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, string(body))
			}
		})
	}
}
