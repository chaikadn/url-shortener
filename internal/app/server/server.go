package server

import (
	"net/http"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/go-chi/chi/v5"
)

func New(handler *handler.Handler, config *config.Config) *http.Server {
	r := chi.NewRouter()
	r.Use(logger.WithLogging)
	r.Mount("/", handler.Route())

	return &http.Server{
		Addr:    config.Host,
		Handler: r,
	}
}
