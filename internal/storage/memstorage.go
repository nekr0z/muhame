// Package storage implements, well, storage.
package storage

import (
	"github.com/nekr0z/muhame/internal/handlers"
	"github.com/nekr0z/muhame/internal/metrics"
)

var _ handlers.MetricsStorage = &MemStorage{}

type MemStorage struct {
	mm map[string]map[string]metrics.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		mm: make(map[string]map[string]metrics.Metric),
	}
}

func (s *MemStorage) Update(name string, m metrics.Metric) error {
	t := m.Type()

	if _, ok := s.mm[t]; !ok {
		s.mm[t] = make(map[string]metrics.Metric)
	}

	have, ok := s.mm[t][name]
	if !ok {
		s.mm[t][name] = m
		return nil
	}

	var err error
	s.mm[t][name], err = have.Update(m)

	return err
}

func (s *MemStorage) Get(t, name string) (metrics.Metric, error) {
	mm, ok := s.mm[t]
	if !ok {
		return nil, handlers.ErrMetricNotFound
	}

	m, ok := mm[name]
	if !ok {
		return nil, handlers.ErrMetricNotFound
	}

	return m, nil
}
