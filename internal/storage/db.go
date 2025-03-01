package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nekr0z/muhame/internal/metrics"
)

const (
	countersTable = "counters"
	gaugesTable   = "gauges"
)

//go:embed migrations/*.sql
var fs embed.FS

type db struct {
	*sql.DB
}

func newDB(dsn string) (*db, error) {
	src, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return nil, err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &db{database}, nil
}

func (db *db) Close() {
	_ = db.DB.Close()
}

func (db *db) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

func (db *db) Get(ctx context.Context, t, name string) (metrics.Metric, error) {
	switch t {
	case metrics.Counter(0).Type():
		return db.getCounter(ctx, name)
	case metrics.Gauge(0).Type():
		return db.getGauge(ctx, name)
	default:
		return nil, fmt.Errorf("unknown type %s", t)
	}
}

func (db *db) Update(ctx context.Context, name string, metric metrics.Metric) error {
	switch v := metric.(type) {
	case metrics.Gauge:
		return db.saveGauge(ctx, name, v)
	case metrics.Counter:
		return db.updateCounter(ctx, name, v)
	default:
		return fmt.Errorf("unknown metric type")
	}
}

func (db *db) List(ctx context.Context) ([]string, []metrics.Metric, error) {
	names := make([]string, 0)
	values := make([]metrics.Metric, 0)

	names, values, err1 := db.appendCounters(ctx, names, values)
	names, values, err2 := db.appendGauges(ctx, names, values)

	return names, values, errors.Join(err1, err2)
}

func (db *db) getCounter(ctx context.Context, name string) (metrics.Counter, error) {
	var c metrics.Counter

	q := fmt.Sprintf("SELECT value FROM %s WHERE name = $1", countersTable)
	r := db.QueryRowContext(ctx, q, name)

	err := scanMetric(&c, r)
	return c, err
}

func (db *db) getGauge(ctx context.Context, name string) (metrics.Gauge, error) {
	var g metrics.Gauge

	q := fmt.Sprintf("SELECT value FROM %s WHERE name = $1", gaugesTable)
	r := db.QueryRowContext(ctx, q, name)

	err := scanMetric(&g, r)
	return g, err
}

func (db *db) saveGauge(ctx context.Context, name string, gauge metrics.Gauge) error {
	q := fmt.Sprintf("INSERT INTO %s(name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value", gaugesTable)
	_, err := db.ExecContext(ctx, q, name, gauge)
	return err
}

func (db *db) updateCounter(ctx context.Context, name string, counter metrics.Counter) error {
	q := fmt.Sprintf("INSERT INTO %s(name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = %s.value + EXCLUDED.value", countersTable, countersTable)
	_, err := db.ExecContext(ctx, q, name, counter)
	return err
}

func (db *db) appendCounters(ctx context.Context, names []string, values []metrics.Metric) ([]string, []metrics.Metric, error) {
	q := fmt.Sprintf("SELECT name, value FROM %s", countersTable)
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return names, values, err
	}
	defer rows.Close()

	var (
		n string
		c metrics.Counter
	)

	for rows.Next() {
		err := rows.Scan(&n, &c)
		if err != nil {
			return names, values, err
		}

		names = append(names, n)
		values = append(values, c)
	}

	return names, values, rows.Err()
}

func (db *db) appendGauges(ctx context.Context, names []string, values []metrics.Metric) ([]string, []metrics.Metric, error) {
	q := fmt.Sprintf("SELECT name, value FROM %s", gaugesTable)
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return names, values, err
	}
	defer rows.Close()

	var (
		n string
		g metrics.Gauge
	)

	for rows.Next() {
		err := rows.Scan(&n, &g)
		if err != nil {
			return names, values, err
		}

		names = append(names, n)
		values = append(values, g)
	}

	return names, values, rows.Err()
}

func scanMetric[M metrics.Counter | metrics.Gauge](m *M, r *sql.Row) error {
	err := r.Scan(m)

	if errors.Is(err, sql.ErrNoRows) {
		err = ErrMetricNotFound
	}

	return err
}
