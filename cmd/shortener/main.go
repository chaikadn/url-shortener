package main

import (
	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/server"
	"github.com/chaikadn/url-shortener/internal/app/storage"
	"github.com/chaikadn/url-shortener/internal/app/storage/file"
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

	stg, err := initStorage(cfg)
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

// ВРЕМЕННО, решить где это должно быть
func initStorage(cfg *config.Config) (storage.Storage, error) {
	if cfg.DatabaseDSN != "" {
		sqlStorage, err := postgresql.NewStorage(cfg.DatabaseDSN)

		// TODO: возможно лучше сделать, чтобы если будет ошибка, отправить в лог и перейти к следующей очереди
		if err != nil {
			return nil, err
		}
		return sqlStorage, nil
	}

	memoryStorage := memory.NewStorage()

	if cfg.FileStoragePath != "" {
		fileStorage, err := file.NewStorage(cfg.FileStoragePath, memoryStorage)

		// TODO: возможно лучше сделать, чтобы если будет ошибка, отправить в лог и перейти к следующей очереди
		if err != nil {
			return nil, err
		}
		return fileStorage, nil
	}

	// fallback
	return memoryStorage, nil
}
