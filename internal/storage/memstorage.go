// Package storage implements, well, storage.
package storage

import "github.com/nekr0z/muhame/internal/metrics"

type MemStorage struct {
	data map[string]metrics.Metric
}
