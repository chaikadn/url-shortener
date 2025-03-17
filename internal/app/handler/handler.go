package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/storage"
	"github.com/chaikadn/url-shortener/internal/app/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	storage storage.Storage
	config  *config.Config
}

func New(stg storage.Storage, cfg *config.Config) (*Handler, error) {
	return &Handler{
		storage: stg,
		config:  cfg,
	}, nil
}

func (h *Handler) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", h.shortenFromText)
	r.Post("/api/shorten", h.shortenFromJSON)
	r.Get("/{short-url}", h.getURL)
	r.Get("/ping", h.pingStorage)
	return r
}

func (h *Handler) getURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "short-url")

	originalURL, err := h.storage.Get(r.Context(), shortURL)
	if err != nil {
		logger.Log.Error("failed to get url", zap.Error(err))
		http.Error(w, "failed to get url", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) shortenFromText(w http.ResponseWriter, r *http.Request) {
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("failed to read request", zap.Error(err))
		http.Error(w, "failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	shortURL, err := h.shortenAndSave(r.Context(), string(originalURL))
	if err != nil {
		logger.Log.Error("failed to shorten url", zap.Error(err))
		http.Error(w, "failed to shorten url", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(h.config.BaseURL + "/" + shortURL))
	if err != nil {
		logger.Log.Error("failed to make response", zap.Error(err))
		http.Error(w, "failed to make response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) shortenFromJSON(w http.ResponseWriter, r *http.Request) {
	req := request{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Error("failed to decode request", zap.Error(err))
		http.Error(w, "failed to decode request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	shortURL, err := h.shortenAndSave(r.Context(), req.URL)
	if err != nil {
		logger.Log.Error("failed to shorten url", zap.Error(err))
		http.Error(w, "failed to shorten url", http.StatusBadRequest)
		return
	}

	resp := response{
		Result: h.config.BaseURL + "/" + shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		logger.Log.Error("failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) shortenAndSave(ctx context.Context, originalURL string) (string, error) {
	if !util.IsValidURL(originalURL) {
		return "", errors.New("url is invalid or empty")
	}

	shortURL := util.RandStr(8)
	if err := h.storage.Add(ctx, originalURL, shortURL); err != nil {
		return "", err
	}

	return shortURL, nil
}

func (h *Handler) pingStorage(w http.ResponseWriter, r *http.Request) {
	err := h.storage.Ping(r.Context())
	if err != nil {
		logger.Log.Error("failed to connect to storage", zap.Error(err))
		http.Error(w, "failed to connect to storage", http.StatusInternalServerError)
		return
	}
}
