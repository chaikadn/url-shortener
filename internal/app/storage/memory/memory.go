package memory

import (
	"context"

	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/storage"
)

// для быстрого двунаправленного поиска использовать две мапы
type MemoryStorage struct {
	storage map[string]*storage.URLEntry
	nextID  int
}

func NewStorage() *MemoryStorage {
	return &MemoryStorage{
		storage: make(map[string]*storage.URLEntry),
		nextID:  0,
	}
}

func (m *MemoryStorage) Add(ctx context.Context, entry *storage.URLEntry) error {
	if _, err := m.GetShort(ctx, entry.OriginalURL); err == nil {
		return storage.ErrLongURLConflict
	}
	if _, ok := m.storage[entry.ShortURL]; ok {
		return storage.ErrShortURLConflict
	}

	entry.ID = m.nextID
	m.storage[entry.ShortURL] = entry
	m.nextID++
	return nil
}

func (m *MemoryStorage) AddBatch(ctx context.Context, batch []*storage.URLEntry) error {
	for _, entry := range batch {
		if err := m.Add(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemoryStorage) GetOriginal(ctx context.Context, shortURL string) (*storage.URLEntry, error) {
	if _, ok := m.storage[shortURL]; !ok {
		return nil, storage.ErrNotFound
	}
	return m.storage[shortURL], nil
}

// для быстрого двунаправленного поиска можно создать 2 мапы
func (m *MemoryStorage) GetShort(ctx context.Context, originalURL string) (*storage.URLEntry, error) {
	for _, entry := range m.storage {
		if entry.OriginalURL == originalURL {
			return entry, nil
		}
	}
	return nil, storage.ErrNotFound
}

func (m *MemoryStorage) Ping(ctx context.Context) (err error) {
	logger.Log.Info("Ping memory storage")
	return nil
}

func (m *MemoryStorage) Close() (err error) {
	return nil
}

func (m *MemoryStorage) GetNextID() int {
	return m.nextID
}
