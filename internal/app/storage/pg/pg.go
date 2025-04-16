package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chaikadn/url-shortener/internal/app/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// TODO: добавить транзакции

type PGStorage struct {
	log *zap.Logger
	db  *sql.DB
}

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
	if err = storage.createTablesQuery(context.Background()); err != nil {
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

func (s *PGStorage) Add(ctx context.Context, entry *model.URLEntry) error {
	if err := s.addNewUser(ctx, entry.UserID); err != nil {
		return err
	}

	pairID, err := s.addNewURLpair(ctx, entry.ShortURL, entry.OriginalURL)

	if err == nil || errors.Is(err, model.ErrLongURLConflict) {
		if addErr := s.addURLpairToUser(ctx, entry.UserID, pairID); addErr != nil {
			return addErr
		}
	}
	return err
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

func (s *PGStorage) GetByID(ctx context.Context, userID string) ([]*model.URLEntry, error) {
	var entries []*model.URLEntry

	rows, err := s.db.QueryContext(ctx, `
		SELECT urls.short_url, urls.original_url
		FROM user_urls
		JOIN urls ON user_urls.url_pair_id = urls.id
		WHERE user_urls.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		entry := model.URLEntry{}
		if err := rows.Scan(&entry.ShortURL, &entry.OriginalURL); err != nil {
			return nil, err
		}
		entry.UserID = userID
		entries = append(entries, &entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// добавляем пользователя если новый
func (s *PGStorage) addNewUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users(user_id) VALUES($1) 
        ON CONFLICT(user_id) DO NOTHING`,
		userID)
	return err
}

// добавляем новую пару url, возвращаем ее id, или id старой пары и ErrLongURLConflict, или -1 и err
func (s *PGStorage) addNewURLpair(ctx context.Context, shortURL, originalURL string) (pairID int, err error) {
	err = s.db.QueryRowContext(ctx, `
        INSERT INTO urls (short_url, original_url)
        VALUES ($1, $2)
        RETURNING id`,
		shortURL, originalURL).Scan(&pairID)
	if err == nil {
		return pairID, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		switch pgErr.ConstraintName {
		case "urls_short_url_key":
			return -1, model.ErrShortURLConflict
		case "urls_original_url_key":
			row := s.db.QueryRowContext(ctx,
				`SELECT id FROM urls WHERE original_url = $1`,
				originalURL)
			if scanErr := row.Scan(&pairID); scanErr != nil {
				return -1, fmt.Errorf("failed to get existing url id: %w", scanErr)
			}
			return pairID, model.ErrLongURLConflict
		default:
			return -1, fmt.Errorf("unexpected unique constraint violation: %w", err)
		}
	}

	return -1, err
}

// связываем пару url с пользователем, если не связана
func (s *PGStorage) addURLpairToUser(ctx context.Context, userID string, pairID int) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_urls(user_id, url_pair_id) VALUES($1, $2)
		ON CONFLICT(user_id, url_pair_id) DO NOTHING`,
		userID, pairID)
	return err
}

func (s *PGStorage) createTablesQuery(ctx context.Context) error {
	// в транзакцию (индексты вроде как не нужны т.к. primary key неявно их создает)
	_, err := s.db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
            short_url VARCHAR(255) NOT NULL UNIQUE,
            original_url VARCHAR(255) NOT NULL UNIQUE
        )`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR(255) PRIMARY KEY
		)`)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_urls (
			user_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
			url_pair_id INT NOT NULL REFERENCES urls(id),
			PRIMARY KEY (user_id, url_pair_id)
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
