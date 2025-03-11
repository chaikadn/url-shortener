package postgresql

import (
	"database/sql"

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

	return &SQLStorage{
		db: db,
	}, nil
}

func (p *SQLStorage) Ping() error {
	return p.db.Ping()
}

func (p *SQLStorage) Close() error {
	return p.db.Close()
}
