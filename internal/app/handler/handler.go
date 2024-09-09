package handler

import (
	"io"
	"net/http"

	"github.com/chaikadn/url-shortener/internal/app/storage"
)

var host, port = "http://localhost", ":8080"

type Handler struct {
	storage storage.Storage
}

func New(st storage.Storage) *Handler {
	return &Handler{storage: st}
}

func (h Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not alowed", http.StatusBadRequest)
		return
	}

	longURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read request: "+err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := h.storage.Add(string(longURL))
	if err != nil {
		http.Error(w, "can't shorten URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(host + port + "/" + shortURL)) // http://localhost:8080/EwHXdJfB
	if err != nil {
		http.Error(w, "can't write responce: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not alowed", http.StatusBadRequest)
		return
	}

	location, err := h.storage.Get(r.URL.Path[1:])
	if err != nil {
		http.Error(w, "can't get long URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
