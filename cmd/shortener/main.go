package main

import (
	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/server"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
	"go.uber.org/zap"
)

func main() {
	cfg := config.New()

	// TODO: обработать ошибки парсинга
	cfg.ParseFlags()
	cfg.ParseEnv()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		logger.Log.Fatal("failed to initialize logger", zap.Error(err))
	}

	// TODO: обрабоать ошибку
	// defer func() {if err...}
	defer logger.Log.Sync()

	stg := memory.NewStorage()
	hnd := handler.New(stg, cfg)
	srv := server.New(hnd, cfg)

	logger.Log.Info("Starting server", zap.String("host", cfg.Host))
	if err := srv.ListenAndServe(); err != nil {
		logger.Log.Fatal("Cannot start server", zap.Error(err))
	}
}
