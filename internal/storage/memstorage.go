// Package storage implements, well, storage.
package storage

import (
	"context"

	"github.com/nekr0z/muhame/internal/metrics"
)

var _ Storage = &memStorage{}

type memStorage struct {
	mm map[string]map[string]metrics.Metric
}

func newMemStorage() *memStorage {
	return &memStorage{
		mm: make(map[string]map[string]metrics.Metric),
	}
}

// Update implements the Storage interface.
func (s *memStorage) Update(_ context.Context, m metrics.Named) error {
	t := m.Type()

	if _, ok := s.mm[t]; !ok {
		s.mm[t] = make(map[string]metrics.Metric)
	}

	have, ok := s.mm[t][m.Name]
	if !ok {
		s.mm[t][m.Name] = m.Metric
		return nil
	}

	var err error
	s.mm[t][m.Name], err = have.Update(m.Metric)

	return err
}

// Get implements the Storage interface.
func (s *memStorage) Get(_ context.Context, t, name string) (metrics.Metric, error) {
	mm, ok := s.mm[t]
	if !ok {
		return nil, ErrMetricNotFound
	}

	m, ok := mm[name]
	if !ok {
		return nil, ErrMetricNotFound
	}

	return m, nil
}

// List implements the Storage interface.
func (s *memStorage) List(_ context.Context) ([]metrics.Named, error) {
	var mms []metrics.Named

	for _, mm := range s.mm {
		for name, m := range mm {
			named := metrics.Named{Name: name, Metric: m}
			mms = append(mms, named)
		}
	}

	return mms, nil
}

// Close implements the Storage interface.
func (s *memStorage) Close() {
}
