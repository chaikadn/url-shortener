package server

import (
	"net/http"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/go-chi/chi/v5"
)

func New(hnd *handler.Handler, cfg *config.Config) *http.Server {
	r := chi.NewRouter()
	r.Use(logger.WithLogging)
	r.Use(handler.WithGzip)
	r.Mount("/", hnd.Route())

	return &http.Server{
		Addr:    cfg.Host,
		Handler: r,
	}
}
