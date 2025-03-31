package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	rt := chi.NewRouter()

	rt.Post("/", h.handleShortenText)
	rt.Get("/ping", h.handlePing)
	rt.Get("/{short-url}", h.handleRedirect)

	rt.Route("/api", func(r chi.Router) {
		// TODO: добавить middleware для игнорирования trailing slashes и сгруппировать маршруты
		r.Post("/shorten", h.handleShortenJSON)
		r.Post("/shorten/batch", h.handleShortenBatch)
	})

	return rt
}

func (h *Handler) handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "short-url")
	entry, err := h.storage.GetOriginal(r.Context(), shortURL)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "url not found", err)
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get url", err)
		return
	}
	http.Redirect(w, r, entry.OriginalURL, http.StatusTemporaryRedirect)
}

func (h *Handler) handleShortenText(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to read body", err)
		return
	}
	defer r.Body.Close()

	shortURL, err := h.shorten(r.Context(), string(body))
	switch {
	case err == nil:
		h.respondText(w, http.StatusCreated, []byte(shortURL)) // 201
	case errors.Is(err, storage.ErrLongURLConflict):
		h.respondText(w, http.StatusConflict, []byte(shortURL)) // 406
	case errors.Is(err, storage.ErrShortURLConflict):
		h.respondError(w, http.StatusInternalServerError, "failed to get short url", err)
	default:
		h.respondError(w, http.StatusBadRequest, "failed to shorten url", err)
	}
}

func (h *Handler) handleShortenJSON(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to decode json body", err)
		return
	}
	defer r.Body.Close()

	shortURL, err := h.shorten(r.Context(), req.URL)
	switch {
	case err == nil:
		h.respondJSON(w, http.StatusCreated, response{Result: shortURL})
	case errors.Is(err, storage.ErrLongURLConflict):
		h.respondJSON(w, http.StatusConflict, response{Result: shortURL})
	case errors.Is(err, storage.ErrShortURLConflict):
		h.respondError(w, http.StatusInternalServerError, "failed to get short url", err)
	default:
		h.respondError(w, http.StatusBadRequest, "failed to shorten url", err)
	}
}

func (h *Handler) handleShortenBatch(w http.ResponseWriter, r *http.Request) {
	var reqBatch []requestBatch
	if err := json.NewDecoder(r.Body).Decode(&reqBatch); err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to decode json body", err)
		return
	}
	defer r.Body.Close()

	if len(reqBatch) == 0 {
		h.respondError(w, http.StatusBadRequest, "batch cannot be empty", storage.ErrEmptyBatch)
		return
	}

	var respBatch []responseBatch
	var status = http.StatusCreated
	for _, entry := range reqBatch {
		shortURL, err := h.shorten(r.Context(), entry.OriginalURL)
		switch {
		case err == nil:
			respBatch = append(respBatch, responseBatch{CorrelationID: entry.CorrelationID, ShortURL: shortURL})
		case errors.Is(err, storage.ErrLongURLConflict):
			status = http.StatusConflict
			respBatch = append(respBatch, responseBatch{CorrelationID: entry.CorrelationID, ShortURL: shortURL})
		default:
			h.respondError(w, http.StatusBadRequest, "failed to shorten batch", err)
			return
		}
	}
	h.respondJSON(w, status, respBatch)
}

func (h *Handler) handlePing(w http.ResponseWriter, r *http.Request) {
	err := h.storage.Ping(r.Context())
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to connect to storage", err)
		return
	}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Log.Error("failed to encode json", zap.Error(err))
	}
}

func (h *Handler) respondText(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	_, err := w.Write(data)
	if err != nil {
		logger.Log.Error("failed to write body", zap.Error(err))
	}
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string, err error) {
	logger.Log.Error(message, zap.Error(err))
	http.Error(w, message, status)
}

func (h *Handler) shorten(ctx context.Context, originalURL string) (string, error) {
	// использовать валидатор (например, go-playground/validator)
	if !util.IsValidURL(originalURL) {
		return "", fmt.Errorf("invalid url '%s'", originalURL)
	}

	// использовать хеширование с обрезкой до 8 символов
	shortKey := util.RandStr(8)
	entry := storage.URLEntry{OriginalURL: originalURL, ShortURL: shortKey}

	switch err := h.storage.Add(ctx, &entry); {
	case err == nil:
		return fmt.Sprintf("%s/%s", h.config.BaseURL, shortKey), nil
	case errors.Is(err, storage.ErrLongURLConflict):
		existingEntry, getErr := h.storage.GetShort(ctx, originalURL)
		if getErr != nil {
			return "", fmt.Errorf("failed to resolve url conflict: %w", getErr)
		}
		return fmt.Sprintf("%s/%s", h.config.BaseURL, existingEntry.ShortURL), storage.ErrLongURLConflict
	default:
		return "", err
	}
}
