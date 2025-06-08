package pg

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/chaikadn/url-shortener/internal/model"
	"github.com/chaikadn/url-shortener/internal/storage"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type Storage struct {
	db  *sql.DB
	log *zap.Logger
}

func New(log *zap.Logger, dsn string) (*Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// не логировать DSN, или маскировать пароль
	log.Info("Successfully connected to Postgres", zap.String("dsn", dsn))

	s := &Storage{
		db:  db,
		log: log,
	}
	if err := s.createTables(); err != nil {
		return nil, fmt.Errorf("create tables: %w", err)
	}
	return s, nil
}

// temporary
func (s *Storage) createTables() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = s.db.Exec(
		`CREATE TABLE IF NOT EXISTS users(
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			role VARCHAR(10) NOT NULL DEFAULT 'user' CHECK(role IN ('user', 'admin')),
			password_hash TEXT NOT NULL
		)`,
	)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	s.log.Debug("Table users successfully created")

	_, err = s.db.Exec(
		`CREATE TABLE IF NOT EXISTS url_pairs(
			id SERIAL PRIMARY KEY,
			key VARCHAR(255) NOT NULL UNIQUE,
			url TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
	)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	s.log.Debug("Table url_pairs successfully created")

	_, err = s.db.Exec(
		`CREATE TABLE IF NOT EXISTS user_urls ( 
    		user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
    		url_id INT NOT NULL REFERENCES url_pairs(id) ON DELETE CASCADE, 
    	PRIMARY KEY (user_id, url_id)
		)`,
	)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	s.log.Debug("Table user_urls successfully created")

	_, err = s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_url_pairs_url ON url_pairs(url); 
		CREATE INDEX IF NOT EXISTS idx_url_pairs_key ON url_pairs(key); 
		CREATE INDEX IF NOT EXISTS idx_user_urls_user ON user_urls(user_id);`,
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	s.log.Debug("Indexes successfully created")

	return tx.Commit()
}

func (s *Storage) SaveUser(user *model.User) (int, error) {
	var userID int
	err := s.db.QueryRow(`
		INSERT INTO users(name, role, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id`,
		user.Name, user.Role, user.PasswordHash).Scan(&userID)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return -1, handlePgError(pgErr)
	}
	if err != nil {
		return -1, fmt.Errorf("storage error: %w", err)
	}

	s.log.Debug("New user added",
		zap.Int("id", userID),
		zap.String("name", user.Name),
	)
	return userID, nil
}

func (s *Storage) GetUserByID(userID int) (*model.User, error) {
	var user model.User
	err := s.db.QueryRow(`
		SELECT id, name, role, password_hash
		FROM users
		WHERE id = $1`,
		userID).Scan(&user.ID, &user.Name, &user.Role, &user.PasswordHash)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	s.log.Debug("User loaded by id",
		zap.Int("id", user.ID),
		zap.String("name", user.Name),
		zap.String("role", user.Role),
	)

	return &user, nil
}

func (s *Storage) GetUserByName(username string) (*model.User, error) {
	var user model.User
	err := s.db.QueryRow(`
		SELECT id, name, role, password_hash
		FROM users
		WHERE name = $1`,
		username).Scan(&user.ID, &user.Name, &user.Role, &user.PasswordHash)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by name: %w", err)
	}

	s.log.Debug("User loaded by name",
		zap.Int("id", user.ID),
		zap.String("name", user.Name),
		zap.String("role", user.Role),
	)

	return &user, nil
}

func (s *Storage) SaveEntry(userID int, entry *model.URLEntry) (int, error) {
	var entryID int
	err := s.db.QueryRow(`
        INSERT INTO url_pairs(url, key)
        VALUES ($1, $2)
        RETURNING id`,
		entry.OriginalURL, entry.Key).Scan(&entryID)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return -1, handlePgError(pgErr)
	}
	if err != nil {
		return -1, fmt.Errorf("storage error: %w", err)
	}

	return entryID, nil
}

func (s *Storage) GetEntryByID(entryID int) (*model.URLEntry, error) {
	var entry model.URLEntry
	err := s.db.QueryRow(`
		SELECT id, key, url, created_at FROM url_pairs
		WHERE id = $1`,
		entryID).Scan(&entry.ID, &entry.Key, &entry.OriginalURL, &entry.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrURLnotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entry by id: %w", err)
	}

	s.log.Debug("Entry loaded by id",
		zap.Int("id", entry.ID),
		zap.String("key", entry.Key),
		zap.String("url", entry.OriginalURL),
	)
	return &entry, nil
}

func (s *Storage) GetEntryByKey(key string) (*model.URLEntry, error) {
	var entry model.URLEntry
	err := s.db.QueryRow(`
		SELECT id, key, url, created_at FROM url_pairs
		WHERE key = $1`,
		key).Scan(&entry.ID, &entry.Key, &entry.OriginalURL, &entry.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrURLnotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entry by key: %w", err)
	}

	s.log.Debug("Entry loaded by key",
		zap.Int("id", entry.ID),
		zap.String("key", entry.Key),
		zap.String("url", entry.OriginalURL),
	)
	return &entry, nil
}

func (s *Storage) GetEntryByURL(originalURL string) (*model.URLEntry, error) {
	var entry model.URLEntry
	err := s.db.QueryRow(`
		SELECT id, key, url, created_at FROM url_pairs
		WHERE url = $1`,
		originalURL).Scan(&entry.ID, &entry.Key, &entry.OriginalURL, &entry.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrURLnotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entry by url: %w", err)
	}

	s.log.Debug("Entry loaded by url",
		zap.Int("id", entry.ID),
		zap.String("key", entry.Key),
		zap.String("url", entry.OriginalURL),
	)
	return &entry, nil
}

func (s *Storage) LinkUserWithEntry(userID int, entryID int) error {
	_, err := s.db.Exec(`
        INSERT INTO user_urls(user_id, url_id)
        VALUES ($1, $2)
		ON CONFLICT (user_id, url_id) DO NOTHING`,
		userID, entryID)

	if err != nil {
		return fmt.Errorf("failed to link url %v with user %v: %w", entryID, userID, err)
	}
	return nil
}

func (s *Storage) GetUserURLs(userID int) (entries []*model.URLEntry, err error) {
	var res []*model.URLEntry
	rows, err := s.db.Query(`
		SELECT up.id, up.key, up.url, up.created_at
        FROM users u
        JOIN user_urls uu ON u.id = uu.user_id
        JOIN url_pairs up ON uu.url_id = up.id
        WHERE u.id = $1`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user urls: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entry model.URLEntry
		if err = rows.Scan(
			&entry.ID,
			&entry.Key,
			&entry.OriginalURL,
			&entry.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user url: %w", err)
		}
		res = append(res, &entry)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to get user urls: %w", err)
	}

	s.log.Debug("User URLs loaded from storage",
		zap.Int("user_id", userID),
		zap.Int("urls_count", len(res)),
	)

	return res, nil
}

func (s *Storage) Ping() error {
	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	s.log.Debug("Succsessful pg storage ping")
	return nil
}

func (s *Storage) Close() error {
	s.log.Debug("Closing pg storage")
	return s.db.Close()
}

func handlePgError(pgErr *pgconn.PgError) error {
	switch pgErr.Code {
	case pgerrcode.UniqueViolation:
		switch pgErr.ConstraintName {
		case "users_name_key":
			return storage.ErrUserAlreadyExists
		case "url_pairs_key_key":
			return storage.ErrKeyAlreadyExists
		case "url_pairs_url_key":
			return storage.ErrURLAlreadyExists
		default:
			return fmt.Errorf("unknown constraint name: %w", pgErr)
		}
	default:
		return fmt.Errorf("unknown err code: %w", pgErr)
	}
}
