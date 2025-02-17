package storage

import (
	"fmt"

	"github.com/nekr0z/muhame/internal/metrics"
)

var ErrMetricNotFound = fmt.Errorf("metric not found")

type Storage interface {
	Get(t, name string) (metrics.Metric, error)
	Update(string, metrics.Metric) error
	List() ([]string, []metrics.Metric, error)
}

type PersistentStorage interface {
	Storage
	Flush()
}
