package storage

import (
	"context"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/model"
	"github.com/chaikadn/url-shortener/internal/app/storage/file"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
	"github.com/chaikadn/url-shortener/internal/app/storage/pg"
	"go.uber.org/zap"
)

type Storage interface {
	Add(ctx context.Context, entry *model.URLEntry) (err error)
	GetByID(ctx context.Context, userID string) (entries []*model.URLEntry, err error)
	GetOriginal(ctx context.Context, shortURL string) (entry *model.URLEntry, err error)
	GetShort(ctx context.Context, originalURL string) (entry *model.URLEntry, err error)
	// AddBatch(ctx context.Context, batch []*model.URLEntry) (err error)
	// Delete(ctx context.Context, shortURL string) (err error)
	Ping(ctx context.Context) (err error)
	Close() (err error)
}

func New(cfg *config.Config, log *zap.Logger) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		return pg.NewStorage(log, cfg.DatabaseDSN)
	}
	if cfg.FileStoragePath != "" {
		return file.NewStorage(log, cfg.FileStoragePath, memory.NewStorage(log))
	}
	return memory.NewStorage(log), nil
}
