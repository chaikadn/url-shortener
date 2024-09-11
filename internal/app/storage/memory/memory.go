package memory

import (
	"errors"

	"github.com/chaikadn/url-shortener/internal/app/util"
)

type MemoryStorage struct {
	Storage map[string]string
}

func NewStorage() *MemoryStorage {
	return &MemoryStorage{
		Storage: make(map[string]string, 0), // нужен ли 0?
	}
}

func (m *MemoryStorage) Add(longURL string) (string, error) {
	shortURL := util.RandStr(8)
	if _, ok := m.Storage[shortURL]; ok {
		return "", errors.New("short URL already exists")
	}
	m.Storage[shortURL] = longURL

	return shortURL, nil
}

func (m *MemoryStorage) Get(shortURL string) (string, error) {
	if _, ok := m.Storage[shortURL]; !ok {
		return "", errors.New("short URL not found")
	}

	return m.Storage[shortURL], nil
}
