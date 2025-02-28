package storage

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nekr0z/muhame/internal/metrics"
)

type DB struct {
	*sql.DB
}

func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) Close() {
	_ = db.DB.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

func (db *DB) Get(t, name string) (metrics.Metric, error) {
	return nil, nil
}

func (db *DB) Update(name string, metric metrics.Metric) error {
	return nil
}

func (db *DB) List() ([]string, []metrics.Metric, error) {
	return nil, nil, nil
}
