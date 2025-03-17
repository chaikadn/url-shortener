package main

import (
	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/server"
	"github.com/chaikadn/url-shortener/internal/app/storage"
	"go.uber.org/zap"
)

func main() {
	cfg := config.New()
	if err := cfg.Load(); err != nil {
		logger.Log.Fatal("failed to initialize config", zap.Error(err))
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		logger.Log.Fatal("failed to initialize logger", zap.Error(err))
	}
	defer logger.Log.Sync()

	// с текущей реализацией всегда будет возвращен nil, исправить
	stg, err := storage.Initialize(cfg)
	if err != nil {
		logger.Log.Fatal("failed to initialize storage", zap.Error(err))
	}
	defer stg.Close()

	hnd, err := handler.New(stg, cfg)
	if err != nil {
		logger.Log.Fatal("failed to initialize handler", zap.Error(err))
	}

	srv := server.New(hnd, cfg)

	logger.Log.Info("Starting server", zap.String("host", cfg.Host))
	if err := srv.ListenAndServe(); err != nil {
		logger.Log.Fatal("failed to start server", zap.Error(err))
	}
}
