package server

import (
	"net/http"

	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/go-chi/chi/v5"
)

func New(handler handler.Handler) *http.Server {
	r := chi.NewRouter()

	r.Mount("/", handler.Route())

	return &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
}
