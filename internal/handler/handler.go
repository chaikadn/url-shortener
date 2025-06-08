package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chaikadn/url-shortener/internal/config"
	"github.com/chaikadn/url-shortener/internal/middleware"
	"github.com/chaikadn/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

type Handler struct {
	logger           *zap.Logger
	config           *config.Config
	shortenerService service.ShortenerService
	userService      service.UserService
}

func New(log *zap.Logger, cfg *config.Config, shr service.ShortenerService, usr service.UserService) *Handler {
	return &Handler{
		logger:           log,
		config:           cfg,
		shortenerService: shr,
		userService:      usr,
	}
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "failed to decode json", http.StatusBadRequest, err)
		return
	}

	token, err := h.userService.Login(req.Username, req.Password, h.config.JWTsecret)
	if err != nil {
		h.respondError(w, "failed to login", http.StatusInternalServerError, err)
		return
	}

	h.respondJSON(w, LoginResponse{Token: token}, http.StatusOK)
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "failed to decode json", http.StatusBadRequest, err)
		return
	}

	userID, err := h.userService.Register(req.Username, req.Password)
	if err != nil {
		h.respondError(w, "failed to register", http.StatusInternalServerError, err)
		return
	}

	user, err := h.userService.GetUser(userID)
	if err != nil {
		h.respondError(w, "failed to get user", http.StatusInternalServerError, err)
		return
	}

	resp := RegisterResponse{
		ID:   user.ID,
		Name: user.Name,
		Role: user.Role,
	}
	h.respondJSON(w, resp, http.StatusCreated)
}

func (h *Handler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	originalURL, err := h.shortenerService.GetOriginalURL(key)

	if err != nil {
		h.respondError(w, "failed to redirect", http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *Handler) HandleShorten(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDkey).(int)
	if !ok {
		h.respondError(w, "authorization failed", http.StatusUnauthorized, fmt.Errorf("failed to get user id"))
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "failed to decode json", http.StatusBadRequest, err)
		return
	}

	key, err := h.shortenerService.CreateShortURL(req.URL, userID)
	if err != nil {
		h.respondError(w, "failed to shorten url", http.StatusInternalServerError, err)
		return
	}

	resp := ShortenResponse{
		Result: fmt.Sprintf("http://%s:%s/%s", h.config.Host, h.config.Port, key),
	}
	h.respondJSON(w, resp, http.StatusCreated)
}

func (h *Handler) HandleUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDkey).(int)
	if !ok {
		h.respondError(w, "authorization failed", http.StatusUnauthorized, fmt.Errorf("failed to get user id"))
		return
	}

	entries, err := h.shortenerService.GetUserURLs(userID)
	if err != nil {
		h.respondError(w, "failed to get user urls", http.StatusInternalServerError, err)
		return
	}

	if len(entries) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var urls []URLPair
	for _, entry := range entries {
		urls = append(
			urls,
			URLPair{
				OriginalURL: entry.OriginalURL,
				ShortURL:    fmt.Sprintf("http://%s:%s/%s", h.config.Host, h.config.Port, entry.Key),
			},
		)
	}

	h.respondJSON(w, UserURLsResponse{UserID: userID, URLs: urls}, http.StatusOK)
}

func (h *Handler) HandlePingStorage(w http.ResponseWriter, r *http.Request) {
	if err := h.shortenerService.PingStorage(); err != nil {
		h.respondError(w, "ping failed", http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// func (h *Handler) handleJSONReq(w http.ResponseWriter, r *http.Request, dest any)

func (h *Handler) respondJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		h.respondError(w, "failed to encode response", http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

func (h *Handler) respondError(w http.ResponseWriter, msg string, status int, err error) {
	h.logger.Error(msg, zap.Error(err))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResp := map[string]string{"error": msg}
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		h.logger.Error("failed to encode error response", zap.Error(err))
	}
}
