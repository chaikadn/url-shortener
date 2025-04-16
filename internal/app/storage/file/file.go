package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/chaikadn/url-shortener/internal/app/model"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
	"go.uber.org/zap"
)

type FileStorage struct {
	log *zap.Logger

	file    *os.File
	decoder *json.Decoder
	encoder *json.Encoder
	memory  *memory.MemoryStorage
}

func NewStorage(log *zap.Logger, filename string, memoryStorage *memory.MemoryStorage) (*FileStorage, error) {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0775); err != nil {
		return nil, err
	}

	fl, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	f := &FileStorage{
		log:     log,
		file:    fl,
		decoder: json.NewDecoder(fl),
		encoder: json.NewEncoder(fl),
		memory:  memoryStorage,
	}
	log.Info("Initialized file storage", zap.String("file", filename))

	if err := f.loadEntries(); err != nil {
		return nil, err
	}

	return f, nil
}

// ПРОВЕРИТЬ
func (f *FileStorage) loadEntries() error {
	if _, err := f.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	count := 0
	for {
		var entry model.URLEntry
		err := f.decoder.Decode(&entry)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to decode %s: %w", f.file.Name(), err)
		}
		err = f.memory.Add(context.Background(), &entry)
		if errors.Is(err, model.ErrLongURLConflict) || errors.Is(err, model.ErrShortURLConflict) {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to load entries from %s: %w", f.file.Name(), err)
		}
		count++
	}
	f.log.Info("Entries loaded", zap.String("file", f.file.Name()), zap.Int("count", count))
	return nil
}

func (f *FileStorage) Add(ctx context.Context, entry *model.URLEntry) error {
	switch err := f.memory.Add(ctx, entry); {
	case err == nil:
		return f.encoder.Encode(entry)
	case errors.Is(err, model.ErrLongURLConflict):
		if e, getErr := f.GetShort(ctx, entry.OriginalURL); getErr == nil {
			e.UserID = entry.UserID
			encErr := f.encoder.Encode(e)
			return errors.Join(err, encErr)
		} else {
			return errors.Join(err, getErr)
		}
	default:
		return err
	}
}

// func (f *FileStorage) AddBatch(ctx context.Context, batch []*model.URLEntry) error {
// 	// TODO: закодировать целиком а, не циклом по одной
// 	var err error
// 	for _, entry := range batch {
// 		err = f.Add(ctx, entry)
// 		if err != nil && !errors.Is(err, model.ErrLongURLConflict) {
// 			return err
// 		}
// 	}
// 	return err
// }

func (f *FileStorage) GetOriginal(ctx context.Context, shortURL string) (*model.URLEntry, error) {
	entry, err := f.memory.GetOriginal(ctx, shortURL)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (f *FileStorage) GetShort(ctx context.Context, originalURL string) (*model.URLEntry, error) {
	entry, err := f.memory.GetShort(ctx, originalURL)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (f *FileStorage) GetByID(ctx context.Context, userID string) ([]*model.URLEntry, error) {
	res, err := f.memory.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (f *FileStorage) Ping(ctx context.Context) error {
	f.log.Info("Ping file storage")
	return nil
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
