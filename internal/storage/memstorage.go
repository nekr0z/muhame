// Package storage implements, well, storage.
package storage

import (
	"context"
	"errors"

	"github.com/nekr0z/muhame/internal/metrics"
)

var ErrNotADatabase = errors.New("not a database")

var _ Storage = &memStorage{}

type memStorage struct {
	mm map[string]map[string]metrics.Metric
}

func newMemStorage() *memStorage {
	return &memStorage{
		mm: make(map[string]map[string]metrics.Metric),
	}
}

func (s *memStorage) Update(_ context.Context, name string, m metrics.Metric) error {
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

func (s *memStorage) List(_ context.Context) ([]string, []metrics.Metric, error) {
	var names []string
	var mms []metrics.Metric

	for _, mm := range s.mm {
		for name, m := range mm {
			names = append(names, name)
			mms = append(mms, m)
		}
	}

	return names, mms, nil
}

func (s *memStorage) Ping(context.Context) error {
	return ErrNotADatabase
}

func (s *memStorage) Close() {
}
