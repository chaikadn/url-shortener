package storage

import (
	"context"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/storage/file"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
	"github.com/chaikadn/url-shortener/internal/app/storage/postgresql"
)

type Storage interface {
	Add(ctx context.Context, longURL string, shortURL string) (err error)
	Get(ctx context.Context, shortURL string) (longURL string, err error)
	// Delete
	Ping(ctx context.Context) (err error)
	Close() (err error)
}

func Initialize(cfg *config.Config) (Storage, error) {
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
