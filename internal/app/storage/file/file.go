package file

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/storage/memory"
)

type urlEntry struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

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
		var entry urlEntry
		err := f.decoder.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := f.memory.Add(context.Background(), entry.OriginalURL, entry.ShortURL); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorage) Add(ctx context.Context, longURL string, shortURL string) error {
	entry := urlEntry{
		ShortURL:    shortURL,
		OriginalURL: longURL,
	}
	if err := f.encoder.Encode(entry); err != nil {
		return err
	}
	return f.memory.Add(ctx, longURL, shortURL)
}

func (f *FileStorage) Get(ctx context.Context, shortURL string) (string, error) {
	return f.memory.Get(ctx, shortURL)
}

func (f *FileStorage) Ping(ctx context.Context) error {
	logger.Log.Info("Ping file storage")
	return nil
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
