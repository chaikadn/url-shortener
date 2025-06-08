package service

import (
	"errors"
	"fmt"

	"github.com/chaikadn/url-shortener/internal/generator"
	"github.com/chaikadn/url-shortener/internal/model"
	"github.com/chaikadn/url-shortener/internal/storage"
)

type ShortenerService interface {
	CreateShortURL(originalURL string, userID int) (key string, err error)
	GetOriginalURL(key string) (originalURL string, err error)
	GetUserURLs(userID int) ([]*model.URLEntry, error)
	PingStorage() error
}

type shortenerService struct {
	db  storage.URLStorage
	gen generator.Generator
}

func NewShortenerService(db storage.URLStorage, gen generator.Generator) ShortenerService {
	return &shortenerService{db: db, gen: gen}
}

// REFACTOR:
func (s *shortenerService) CreateShortURL(originalURL string, userID int) (string, error) {
	key, err := s.gen.Generate(originalURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}

	entryID, err := s.db.SaveEntry(userID, &model.URLEntry{Key: key, OriginalURL: originalURL})

	// TODO: retry
	if errors.Is(err, storage.ErrKeyAlreadyExists) {
		return "", storage.ErrKeyAlreadyExists
	}

	if errors.Is(err, storage.ErrURLAlreadyExists) {
		existingEntry, err := s.db.GetEntryByURL(originalURL)
		if err != nil {
			return "", fmt.Errorf("failed to load existing entry: %w", err)
		}
		entryID = existingEntry.ID
		key = existingEntry.Key
	}

	if err != nil && !errors.Is(err, storage.ErrURLAlreadyExists) {
		return "", fmt.Errorf("failed to save entry: %w", err)
	}

	err = s.db.LinkUserWithEntry(userID, entryID)
	if err != nil {
		return "", err
	}

	return key, nil
}

func (s *shortenerService) GetOriginalURL(key string) (string, error) {
	entry, err := s.db.GetEntryByKey(key)
	if err != nil {
		return "", fmt.Errorf("failed to get original url: %w", err)
	}
	return entry.OriginalURL, nil
}

func (s *shortenerService) GetUserURLs(userID int) ([]*model.URLEntry, error) {
	return s.db.GetUserURLs(userID)
}

func (s *shortenerService) PingStorage() error {
	return s.db.Ping()
}
