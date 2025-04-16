package memory

import (
	"context"

	"github.com/chaikadn/url-shortener/internal/app/model"
	"go.uber.org/zap"
)

type MemoryStorage struct {
	// добавить мьютекс

	log *zap.Logger

	shortToLong map[string]string
	longToShort map[string]string
	userUrls    userURLs
}

func NewStorage(log *zap.Logger) *MemoryStorage {
	log.Info("Initialized memory storage")
	return &MemoryStorage{
		log:         log,
		shortToLong: make(map[string]string),
		longToShort: make(map[string]string),
		userUrls:    make(userURLs),
	}
}

func (m *MemoryStorage) Add(ctx context.Context, entry *model.URLEntry) error {
	if existingShort, ok := m.longToShort[entry.OriginalURL]; ok {
		m.userUrls.AddUser(entry.UserID)
		m.userUrls.AddUserURL(entry.UserID, existingShort)
		return model.ErrLongURLConflict
	}
	if _, ok := m.shortToLong[entry.ShortURL]; ok {
		return model.ErrShortURLConflict
	}

	m.shortToLong[entry.ShortURL] = entry.OriginalURL
	m.longToShort[entry.OriginalURL] = entry.ShortURL

	m.userUrls.AddUser(entry.UserID)
	m.userUrls.AddUserURL(entry.UserID, entry.ShortURL)

	return nil
}

// func (m *MemoryStorage) AddBatch(ctx context.Context, batch []*model.URLEntry) error {
// 	var err error
// 	for _, entry := range batch {
// 		err = m.Add(ctx, entry)
// 		if err != nil && !errors.Is(err, model.ErrLongURLConflict) {
// 			return err
// 		}
// 	}
// 	return err
// }

func (m *MemoryStorage) GetOriginal(ctx context.Context, shortURL string) (*model.URLEntry, error) {
	if _, ok := m.shortToLong[shortURL]; !ok {
		return nil, model.ErrNotFound
	}
	entry := &model.URLEntry{
		UserID:      "",
		ShortURL:    shortURL,
		OriginalURL: m.shortToLong[shortURL],
	}
	return entry, nil
}

func (m *MemoryStorage) GetShort(ctx context.Context, originalURL string) (*model.URLEntry, error) {
	if _, ok := m.longToShort[originalURL]; !ok {
		return nil, model.ErrNotFound
	}
	entry := &model.URLEntry{
		UserID:      "",
		ShortURL:    m.longToShort[originalURL],
		OriginalURL: originalURL,
	}
	return entry, nil
}

func (m *MemoryStorage) GetByID(ctx context.Context, userID string) ([]*model.URLEntry, error) {
	res := []*model.URLEntry{}
	if urls, ok := m.userUrls[userID]; ok {
		for shortURL := range urls {
			entry := &model.URLEntry{
				UserID:      userID,
				ShortURL:    shortURL,
				OriginalURL: m.shortToLong[shortURL],
			}
			res = append(res, entry)
		}
	}
	return res, nil
}

func (m *MemoryStorage) Ping(ctx context.Context) (err error) {
	m.log.Info("Ping memory storage")
	return nil
}

func (m *MemoryStorage) Close() (err error) {
	return nil
}
