package memory

import (
	"errors"

	"github.com/chaikadn/url-shortener/internal/app/storage"
	"github.com/chaikadn/url-shortener/internal/app/util"
)

var _ storage.Storage = &memoryStorage{}

type memoryStorage struct {
	Storage map[string]string
}

func NewStorage() storage.Storage {
	return &memoryStorage{
		Storage: make(map[string]string),
	}
}

func (m *memoryStorage) Add(longURL string) (string, error) {
	shortURL := util.RandStr(8)

	if _, ok := m.Storage[shortURL]; ok {
		return "", errors.New("short URL already exists")
	}
	m.Storage[shortURL] = longURL

	return shortURL, nil
}

func (m *memoryStorage) Get(shortURL string) (string, error) {

	if _, ok := m.Storage[shortURL]; !ok {
		return "", errors.New("short URL not found")
	}

	return m.Storage[shortURL], nil
}
