package memory

import (
	"errors"

	"github.com/chaikadn/url-shortener/internal/app/model"
)

type MemoryStorage struct {
	storage map[string]*model.URLEntry
	nextID  int
}

func NewStorage() *MemoryStorage {
	return &MemoryStorage{
		storage: make(map[string]*model.URLEntry),
		nextID:  1,
	}
}

func (m *MemoryStorage) Add(data *model.URLEntry) error {
	if _, ok := m.storage[data.ShortURL]; ok {
		return errors.New("short URL already exists")
	}
	m.storage[data.ShortURL] = data
	m.nextID++
	return nil
}

func (m *MemoryStorage) Get(shortURL string) (*model.URLEntry, error) {
	if _, ok := m.storage[shortURL]; !ok {
		return nil, errors.New("short URL not found")
	}
	return m.storage[shortURL], nil
}

func (m *MemoryStorage) GetNextID() int {
	return m.nextID
}
