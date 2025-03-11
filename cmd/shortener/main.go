package main

import (
	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/server"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
	"github.com/chaikadn/url-shortener/internal/app/storage/postgresql"
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

	memStg := memory.NewStorage()

	sqlStg, err := postgresql.NewStorage(cfg.DatabaseDSN)
	if err != nil {
		logger.Log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer sqlStg.Close()

	hnd, err := handler.New(memStg, sqlStg, cfg)
	if err != nil {
		logger.Log.Fatal("failed to initialize handler", zap.Error(err))
	}

	srv := server.New(hnd, cfg)

	logger.Log.Info("Starting server", zap.String("host", cfg.Host))
	if err := srv.ListenAndServe(); err != nil {
		logger.Log.Fatal("failed to start server", zap.Error(err))
	}
}
