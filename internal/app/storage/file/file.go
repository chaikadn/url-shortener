package file

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/storage"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
)

type FileStorage struct {
	file    *os.File
	decoder *json.Decoder
	encoder *json.Encoder
	memory  *memory.MemoryStorage
}

func NewStorage(filename string, memoryStorage *memory.MemoryStorage) (*FileStorage, error) {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0775); err != nil {
		return nil, err
	}

	fl, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	f := &FileStorage{
		file:    fl,
		decoder: json.NewDecoder(fl),
		encoder: json.NewEncoder(fl),
		memory:  memoryStorage,
	}

	if err := f.loadEntries(); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *FileStorage) loadEntries() error {
	if _, err := f.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	for {
		var entry storage.URLEntry
		err := f.decoder.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := f.memory.Add(context.Background(), &entry); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorage) Add(ctx context.Context, entry *storage.URLEntry) error {
	// не гарантируется синхронизация file и memory, eсли f.encoder.Encode(entry) выдаст ошибку

	if err := f.memory.Add(ctx, entry); err != nil {
		return err
	}
	if err := f.encoder.Encode(entry); err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) AddBatch(ctx context.Context, batch []*storage.URLEntry) error {
	for _, entry := range batch {
		if err := f.Add(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorage) GetOriginal(ctx context.Context, shortURL string) (*storage.URLEntry, error) {
	entry, err := f.memory.GetOriginal(ctx, shortURL)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (f *FileStorage) GetShort(ctx context.Context, originalURL string) (*storage.URLEntry, error) {
	entry, err := f.memory.GetShort(ctx, originalURL)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (f *FileStorage) Ping(ctx context.Context) error {
	logger.Log.Info("Ping file storage")
	return nil
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
