package postgresql

import (
	"context"
	"database/sql"

	"github.com/chaikadn/url-shortener/internal/app/logger"
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

	if err = sqlStorage.cteateTableQuery(context.Background()); err != nil {
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

func (s *SQLStorage) Add(ctx context.Context, longURL string, shortURL string) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO urls (short_url, long_url) VALUES ($1, $2)", shortURL, longURL)
	return err
}

func (s *SQLStorage) Get(ctx context.Context, shortURL string) (string, error) {
	row := s.db.QueryRowContext(ctx, "SELECT long_url FROM urls WHERE short_url = $1", shortURL)
	var url string
	err := row.Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *SQLStorage) cteateTableQuery(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS urls ("+
		"short_url VARCHAR(100) UNIQUE NOT NULL,"+
		"long_url VARCHAR(200) NOT NULL)")
	return err
}
