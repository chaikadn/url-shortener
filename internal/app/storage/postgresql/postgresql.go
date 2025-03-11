package postgresql

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type SqlStorage struct {
	db *sql.DB
}

func NewStorage(dsn string) (*SqlStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &SqlStorage{
		db: db,
	}, nil
}

func (p *SqlStorage) Ping() error {
	return p.db.Ping()
}

func (p *SqlStorage) Close() error {
	return p.db.Close()
}
