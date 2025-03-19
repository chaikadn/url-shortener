package memory

import (
	"context"
	"errors"

	"github.com/chaikadn/url-shortener/internal/app/logger"
)

type MemoryStorage struct {
	storage map[string]string
}

func NewStorage() *MemoryStorage {
	return &MemoryStorage{
		storage: make(map[string]string),
	}
}

func (m *MemoryStorage) Add(ctx context.Context, longURL string, shortURL string) error {
	if _, ok := m.storage[shortURL]; ok {
		return errors.New("short url already exists")
	}
	m.storage[shortURL] = longURL
	return nil
}

func (m *MemoryStorage) Get(ctx context.Context, shortURL string) (string, error) {
	if _, ok := m.storage[shortURL]; !ok {
		return "", errors.New("short url not found")
	}
	return m.storage[shortURL], nil
}

func (m *MemoryStorage) Ping(ctx context.Context) (err error) {
	logger.Log.Info("Ping memory storage")
	return nil
}

func (m *MemoryStorage) Close() (err error) {
	return nil
}
