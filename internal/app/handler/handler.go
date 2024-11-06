package handler

import (
	"encoding/json"
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
	r.Post("/api/shorten", h.shorten)
	r.Get("/{id}", h.getURL)
	return r
}

func (h *Handler) postURL(w http.ResponseWriter, r *http.Request) {
	longURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !util.IsValidURL(string(longURL)) {
		http.Error(w, "URL is invalid or empty", http.StatusBadRequest)
		return
	}

	shortURL, err := h.storage.Add(string(longURL))
	if err != nil {
		http.Error(w, "Unable to shorten URL: ", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(h.config.BaseURL + "/" + shortURL)) // http://localhost:8080/EwHXdJfB
	if err != nil {
		http.Error(w, "Unable to make response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	location, err := h.storage.Get(id)

	if err != nil {
		http.Error(w, "Unable to get URL", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	req := request{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}
	if !util.IsValidURL(req.URL) {
		http.Error(w, "URL is invalid or empty", http.StatusBadRequest)
		return
	}
	res, err := h.storage.Add(string(req.URL))
	if err != nil {
		http.Error(w, "cannot shorten URL", http.StatusInternalServerError)
		return
	}
	resp := Response{
		Result: h.config.BaseURL + "/" + res,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
