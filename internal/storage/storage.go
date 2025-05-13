// Package storage implements all kinds of storages.
package storage

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nekr0z/muhame/internal/metrics"
)

// ErrMetricNotFound is returned when metric is not found.
var ErrMetricNotFound = fmt.Errorf("metric not found")

// Storage provides a metric storage.
type Storage interface {
	Get(ctx context.Context, t, name string) (metrics.Metric, error)
	Update(context.Context, metrics.Named) error
	List(context.Context) ([]metrics.Named, error)
	Close()
}

// New returns a new storage.
func New(log *zap.SugaredLogger, cfg Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		log.Info("using database for storage")
		return newDB(cfg.DatabaseDSN)
	}

	if cfg.InMemory {
		return newMemStorage(), nil
	}

	return newFileStorage(context.TODO(), log, cfg), nil
}
