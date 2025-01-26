package handlers

import (
	"fmt"

	"github.com/nekr0z/muhame/internal/metrics"
)

type MetricsStorage interface {
	Update(string, metrics.Metric) error
	Get(string, string) (metrics.Metric, error)
}

var ErrMetricNotFound = fmt.Errorf("metric not found")
