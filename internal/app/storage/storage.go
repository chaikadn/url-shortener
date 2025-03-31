package storage

import (
	"context"
	"errors"
)

type URLEntry struct {
	ID          int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var (
	ErrLongURLConflict  = errors.New("long url conflict")
	ErrShortURLConflict = errors.New("short url conflict")
	ErrNotFound         = errors.New("url not found")
	ErrEmptyBatch       = errors.New("empty batch")
)

type Storage interface {
	Add(ctx context.Context, entry *URLEntry) (err error)
	AddBatch(ctx context.Context, batch []*URLEntry) (err error)
	GetOriginal(ctx context.Context, shortURL string) (entry *URLEntry, err error)
	GetShort(ctx context.Context, originalURL string) (entry *URLEntry, err error)
	// Delete(ctx context.Context, shortURL string) (err error)
	Ping(ctx context.Context) (err error)
	Close() (err error)
}
