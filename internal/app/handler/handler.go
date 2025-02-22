package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/model"
	"github.com/chaikadn/url-shortener/internal/app/storage/file"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
	"github.com/chaikadn/url-shortener/internal/app/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// TODO: настроить логировангие через существующий middleware

type Handler struct {
	memoryStorage *memory.MemoryStorage
	config        *config.Config
}

func New(memSt *memory.MemoryStorage, cfg *config.Config) (*Handler, error) {
	if cfg.FileStoragePath != "" {
		dec, err := file.NewJSONDecoder(cfg.FileStoragePath)
		if err != nil {
			return nil, errors.New("failed to open file: " + err.Error())
		}
		defer dec.Close()

		for {
			entry := model.URLEntry{}
			err := dec.ReadTo(&entry)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, errors.New("failed decode file: " + err.Error())
			}
			err = memSt.Add(&entry)
			if err != nil {
				return nil, errors.New("failed to save URL: " + err.Error())
			}
		}
	}
	return &Handler{
		memoryStorage: memSt,
		config:        cfg,
	}, nil
}

func (h *Handler) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", h.shortenFromText)
	r.Post("/api/shorten", h.shortenFromJSON)
	r.Get("/{short-url}", h.getURL)
	return r
}

func (h *Handler) shortenFromText(w http.ResponseWriter, r *http.Request) {
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error("failed to read request", zap.Error(err))
		http.Error(w, "failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	shortURL, err := h.shortenAndSave(string(originalURL))
	if err != nil {
		logger.Log.Error("failed to shorten URL", zap.Error(err))
		http.Error(w, "failed to shorten URL", http.StatusBadRequest)
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

func (h *Handler) getURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "short-url")

	data, err := h.memoryStorage.Get(shortURL)
	if err != nil {
		logger.Log.Error("failed to get URL", zap.Error(err))
		http.Error(w, "failed to get URL", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", data.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
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

	shortURL, err := h.shortenAndSave(req.URL)
	if err != nil {
		logger.Log.Error("failed to shorten URL", zap.Error(err))
		http.Error(w, "failed to shorten URL", http.StatusBadRequest)
		return
	}

	resp := Response{
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

func (h *Handler) shortenAndSave(originalURL string) (string, error) {
	if !util.IsValidURL(originalURL) {
		return "", errors.New("URL is invalid or empty")
	}

	urlEntry := model.URLEntry{
		ID:          h.memoryStorage.GetNextID(),
		ShortURL:    util.RandStr(8),
		OriginalURL: originalURL,
	}

	if err := h.memoryStorage.Add(&urlEntry); err != nil {
		// http.Error(w, "failed to shorten URL", http.StatusInternalServerError)
		return "", errors.New("failed to save URL: " + err.Error())
	}
	if h.config.FileStoragePath != "" {
		enc, err := file.NewJSONEncoder(h.config.FileStoragePath)
		if err != nil {
			// http.Error(w, "failed to write file", http.StatusInternalServerError)
			return "", errors.New("failed to open file: " + err.Error())
		}
		defer enc.Close()
		err = enc.WriteFrom(&urlEntry)
		if err != nil {
			return "", errors.New("failed to encode file: " + err.Error())
		}
	}
	return urlEntry.ShortURL, nil
}
