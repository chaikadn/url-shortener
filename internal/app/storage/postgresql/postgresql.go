package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/storage"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLStorage struct {
	db *sql.DB
}

func NewStorage(dsn string) (*SQLStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	sqlStorage := &SQLStorage{
		db: db,
	}

	// нужно ли?
	if err := sqlStorage.Ping(context.Background()); err != nil {
		return nil, err
	}

	if err = sqlStorage.createTableQuery(context.Background()); err != nil {
		return nil, err
	}
	return sqlStorage, nil
}

func (s *SQLStorage) Ping(ctx context.Context) error {
	logger.Log.Info("Ping postgresql storage")
	return s.db.PingContext(ctx)
}

func (s *SQLStorage) Close() error {
	return s.db.Close()
}

// протестировать
func (s *SQLStorage) Add(ctx context.Context, entry *storage.URLEntry) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO urls (short_url, original_url)
		VALUES ($1, $2)`,
		entry.ShortURL, entry.OriginalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			switch pgErr.ConstraintName {
			case "urls_short_url_key":
				return storage.ErrShortURLConflict
			case "urls_original_url_key":
				return storage.ErrLongURLConflict
			}
		}
		return err
	}
	return nil
}

func (s *SQLStorage) AddBatch(ctx context.Context, batch []*storage.URLEntry) error {
	// Транзакция
	for _, entry := range batch {
		err := s.Add(ctx, entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLStorage) GetOriginal(ctx context.Context, shortURL string) (*storage.URLEntry, error) {
	entry := storage.URLEntry{ShortURL: shortURL}
	row := s.db.QueryRowContext(ctx, `SELECT original_url FROM urls WHERE short_url = $1`, shortURL)
	err := row.Scan(&entry.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return &entry, nil
}

func (s *SQLStorage) GetShort(ctx context.Context, originalURL string) (*storage.URLEntry, error) {
	entry := storage.URLEntry{OriginalURL: originalURL}
	row := s.db.QueryRowContext(ctx, `SELECT short_url FROM urls WHERE original_url = $1`, originalURL)
	err := row.Scan(&entry.ShortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return &entry, nil
}

func (s *SQLStorage) createTableQuery(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS urls (
            short_url VARCHAR(128) UNIQUE NOT NULL,
            original_url VARCHAR(256) UNIQUE NOT NULL
        )`)
	if err != nil {
		return err
	}

	// _, err = s.db.ExecContext(ctx, `
	//     CREATE INDEX IF NOT EXISTS idx_urls_long_url
	//     ON urls (long_url)`)
	// if err != nil {
	// 	return err
	// }
	return nil
}
