package storage

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nekr0z/muhame/internal/metrics"
)

var ErrMetricNotFound = fmt.Errorf("metric not found")

type Storage interface {
	Get(ctx context.Context, t, name string) (metrics.Metric, error)
	Update(context.Context, metrics.Named) error
	List(context.Context) ([]metrics.Named, error)
	Close()
}

func New(log *zap.SugaredLogger, cfg Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		log.Info("using database for storage")
		return newDB(cfg.DatabaseDSN)
	}

	return newFileStorage(context.TODO(), log, cfg), nil
}
