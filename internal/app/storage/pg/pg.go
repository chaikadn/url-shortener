package pg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chaikadn/url-shortener/internal/app/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type PGStorage struct {
	log *zap.Logger
	db  *sql.DB
}

// TODO:
// для привязки сокращенных url к uuid пользователя создать новую таблицу users,
// в которой связываются пользователи и уникальные короткие url

func NewStorage(log *zap.Logger, dsn string) (*PGStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	log.Info("Connected to postgress", zap.String("dsn", dsn))

	storage := &PGStorage{
		log: log,
		db:  db,
	}
	// сделать миграцию вместо этого
	if err = storage.createTableQuery(context.Background()); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *PGStorage) Ping(ctx context.Context) error {
	s.log.Info("Ping postgresql storage")
	return s.db.PingContext(ctx)
}

func (s *PGStorage) Close() error {
	return s.db.Close()
}

// протестировать
func (s *PGStorage) Add(ctx context.Context, entry *model.URLEntry) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO urls (short_url, original_url)
		VALUES ($1, $2)`,
		entry.ShortURL, entry.OriginalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			switch pgErr.ConstraintName {
			case "urls_short_url_key":
				return model.ErrShortURLConflict
			case "urls_original_url_key":
				return model.ErrLongURLConflict
			}
		}
		return err
	}
	return nil
}

// func (s *PGStorage) AddBatch(ctx context.Context, batch []*model.URLEntry) error {
// 	// TODO: сделать транзакцию со всей пачкой, а не циклом по одной
// 	var err error
// 	for _, entry := range batch {
// 		err = s.Add(ctx, entry)
// 		if err != nil && !errors.Is(err, model.ErrLongURLConflict) {
// 			return err
// 		}
// 	}
// 	return err
// }

func (s *PGStorage) GetOriginal(ctx context.Context, shortURL string) (*model.URLEntry, error) {
	entry := model.URLEntry{ShortURL: shortURL}
	row := s.db.QueryRowContext(ctx, `SELECT original_url FROM urls WHERE short_url = $1`, shortURL)
	err := row.Scan(&entry.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &entry, nil
}

func (s *PGStorage) GetShort(ctx context.Context, originalURL string) (*model.URLEntry, error) {
	entry := model.URLEntry{OriginalURL: originalURL}
	row := s.db.QueryRowContext(ctx, `SELECT short_url FROM urls WHERE original_url = $1`, originalURL)
	err := row.Scan(&entry.ShortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &entry, nil
}

// TODO:
func (s *PGStorage) GetByID(ctx context.Context, userID string) ([]*model.URLEntry, error) {
	return []*model.URLEntry{}, nil
}

func (s *PGStorage) createTableQuery(ctx context.Context) error {
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
