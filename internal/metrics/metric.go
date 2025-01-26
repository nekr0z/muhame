// Package metrics represents metrics.
package metrics

import (
	"fmt"
	"strconv"
)

type Metric interface {
	String() string
	Update(Metric) (Metric, error)
	Type() string
}

// Gauge represents a gauge metric.
type Gauge float64

var _ Metric = Gauge(0)

func ParseGauge(s string) (Gauge, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse gauge value: %w", err)
	}
	return Gauge(v), nil
}

func (g Gauge) String() string {
	return fmt.Sprintf("%f", g)
}

func (g Gauge) Update(m Metric) (Metric, error) {
	n, ok := m.(Gauge)
	if !ok {
		return g, fmt.Errorf("cannot update gauge with non-gauge metric")
	}
	return n, nil
}

func (g Gauge) Type() string {
	return "gauge"
}

// Counter represents a counter metric.
type Counter int64

var _ Metric = Counter(0)

func ParseCounter(s string) (Counter, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse counter value: %w", err)
	}
	return Counter(v), nil
}

func (c Counter) String() string {
	return fmt.Sprintf("%d", c)
}

func (c Counter) Update(m Metric) (Metric, error) {
	inc, ok := m.(Counter)
	if !ok {
		return c, fmt.Errorf("cannot update counter with non-counter metric")
	}
	return c + inc, nil
}

func (c Counter) Type() string {
	return "counter"
}
