package storage

import (
	"context"
	"fmt"

	"github.com/nekr0z/muhame/internal/metrics"
	"go.uber.org/zap"
)

var ErrMetricNotFound = fmt.Errorf("metric not found")

type Storage interface {
	Get(t, name string) (metrics.Metric, error)
	Update(string, metrics.Metric) error
	List() ([]string, []metrics.Metric, error)
	Ping(context.Context) error
	Close()
}

func New(log *zap.SugaredLogger, cfg Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		log.Info("using database for storage")
		return NewDB(cfg.DatabaseDSN)
	}

	return NewFileStorage(log, cfg), nil
}
