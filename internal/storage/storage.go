package storage

import (
	"context"
	"fmt"

	"github.com/nekr0z/muhame/internal/metrics"
	"go.uber.org/zap"
)

var ErrMetricNotFound = fmt.Errorf("metric not found")

type Storage interface {
	Get(ctx context.Context, t, name string) (metrics.Metric, error)
	Update(context.Context, string, metrics.Metric) error
	List(context.Context) ([]string, []metrics.Metric, error)
	Ping(context.Context) error
	Close()
}

func New(log *zap.SugaredLogger, cfg Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		log.Info("using database for storage")
		return newDB(cfg.DatabaseDSN)
	}

	return newFileStorage(context.TODO(), log, cfg), nil
}
