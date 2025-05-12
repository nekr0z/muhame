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
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/retry"
)

const (
	countersTable = "counters"
	gaugesTable   = "gauges"
)

var (
	gaugeInsert   = fmt.Sprintf("INSERT INTO %s(name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value", gaugesTable)
	counterInsert = fmt.Sprintf("INSERT INTO %s(name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = %s.value + EXCLUDED.value", countersTable, countersTable)
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

// Close implements the Storage interface.
func (db *db) Close() {
	_ = db.DB.Close()
}

// Ping allows to check if the connection is alive.
func (db *db) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

// Get implements the Storage interface.
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

// Update implements the Storage interface.
func (db *db) Update(ctx context.Context, metric metrics.Named) error {
	switch v := metric.Metric.(type) {
	case metrics.Gauge:
		return db.saveGauge(ctx, metric.Name, v)
	case metrics.Counter:
		return db.updateCounter(ctx, metric.Name, v)
	default:
		return fmt.Errorf("unknown metric type")
	}
}

// List returns all metrics.
func (db *db) List(ctx context.Context) ([]metrics.Named, error) {
	values := make([]metrics.Named, 0)

	values, err1 := db.appendCounters(ctx, values)
	values, err2 := db.appendGauges(ctx, values)

	return values, errors.Join(err1, err2)
}

// BulkUpdate updates multiple metrics in a single transaction.
func (db *db) BulkUpdate(ctx context.Context, mm []metrics.Named) error {
	tx, err := retry.OnError(func() (*sql.Tx, error) {
		return db.BeginTx(ctx, nil)
	}, isConnectionException)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmtGauge, err := retry.OnError(func() (*sql.Stmt, error) {
		return tx.PrepareContext(ctx, gaugeInsert)
	}, isConnectionException)
	if err != nil {
		return err
	}
	defer stmtGauge.Close()

	stmtCounter, err := retry.OnError(func() (*sql.Stmt, error) {
		return tx.PrepareContext(ctx, counterInsert)
	}, isConnectionException)
	if err != nil {
		return err
	}
	defer stmtCounter.Close()

	for _, m := range mm {
		switch v := m.Metric.(type) {
		case metrics.Counter:
			_, err = stmtCounter.ExecContext(ctx, m.Name, v)
		case metrics.Gauge:
			_, err = stmtGauge.ExecContext(ctx, m.Name, v)
		default:
			err = fmt.Errorf("unknown metric type")
		}
		if err != nil {
			return err
		}
	}

	return retry.Error(func() error {
		return tx.Commit()
	}, isConnectionException)
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
	_, err := retry.OnError(func() (sql.Result, error) {
		return db.ExecContext(ctx, gaugeInsert, name, gauge)
	}, isConnectionException)
	return err
}

func (db *db) updateCounter(ctx context.Context, name string, counter metrics.Counter) error {
	_, err := retry.OnError(func() (sql.Result, error) {
		return db.ExecContext(ctx, counterInsert, name, counter)
	}, isConnectionException)
	return err
}

func (db *db) appendCounters(ctx context.Context, values []metrics.Named) ([]metrics.Named, error) {
	q := fmt.Sprintf("SELECT name, value FROM %s", countersTable)

	rows, err := retry.OnError(func() (*sql.Rows, error) {
		return db.QueryContext(ctx, q)
	}, isConnectionException)
	if err != nil {
		return values, err
	}
	defer rows.Close()

	var (
		n string
		c metrics.Counter
	)

	for rows.Next() {
		err := rows.Scan(&n, &c)
		if err != nil {
			return values, err
		}

		values = append(values, metrics.Named{
			Name:   n,
			Metric: c,
		})
	}

	return values, rows.Err()
}

func (db *db) appendGauges(ctx context.Context, values []metrics.Named) ([]metrics.Named, error) {
	q := fmt.Sprintf("SELECT name, value FROM %s", gaugesTable)

	rows, err := retry.OnError(func() (*sql.Rows, error) {
		return db.QueryContext(ctx, q)
	}, isConnectionException)
	if err != nil {
		return values, err
	}
	defer rows.Close()

	var (
		n string
		g metrics.Gauge
	)

	for rows.Next() {
		err := rows.Scan(&n, &g)
		if err != nil {
			return values, err
		}

		values = append(values, metrics.Named{
			Name:   n,
			Metric: g,
		})
	}

	return values, rows.Err()
}

func scanMetric[M metrics.Counter | metrics.Gauge](m *M, r *sql.Row) error {
	err := r.Scan(m)

	if errors.Is(err, sql.ErrNoRows) {
		err = ErrMetricNotFound
	}

	return err
}

func isConnectionException(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgerrcode.IsConnectionException(pgErr.Code)
}
