package server

import (
	"net/http"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func New(cfg *config.Config, log *zap.Logger, hnd *handler.Handler) *http.Server {
	r := chi.NewRouter()
	r.Use(
		logger.WithLogging(log),
		handler.WithGzip,
		handler.WithAuth(log, cfg.JWTSecret),
	)
	r.Mount("/", hnd.Route())

	return &http.Server{
		Addr:    cfg.Host,
		Handler: r,
	}
}
