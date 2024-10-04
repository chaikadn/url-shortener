package handler

import (
	"io"
	"net/http"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/storage"
	"github.com/chaikadn/url-shortener/internal/app/util"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	storage storage.Storage
	config  *config.Config
	// logger interface
}

func New(st storage.Storage, cn *config.Config) *Handler {
	return &Handler{
		storage: st,
		config:  cn,
	}
}

func (h *Handler) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", h.postURL)
	r.Get("/{id}", h.getURL)
	return r
}

func (h *Handler) postURL(w http.ResponseWriter, r *http.Request) {
	longURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request", http.StatusBadRequest)
		// log.Printf("can't read request: %v", err.Error())
		return
	}
	defer r.Body.Close()

	if !util.IsValidURL(string(longURL)) {
		http.Error(w, "URL is invalid or empty", http.StatusBadRequest)
		// log.Printf(...)
		return
	}

	shortURL, err := h.storage.Add(string(longURL))
	if err != nil {
		http.Error(w, "Unable to shorten URL: ", http.StatusInternalServerError)
		// log.Printf(...)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(h.config.BaseURL + "/" + shortURL)) // http://localhost:8080/EwHXdJfB
	if err != nil {
		http.Error(w, "Unable to make responce", http.StatusInternalServerError)
		// log.Printf(...)
		return
	}
}

func (h *Handler) getURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	location, err := h.storage.Get(id)

	if err != nil {
		http.Error(w, "Unable to get URL", http.StatusBadRequest)
		// log.Printf("Error getting URL for id %s: %v", id, err)
		return
	}

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
