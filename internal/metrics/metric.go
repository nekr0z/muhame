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

// Parse returns Metric of correct type t and value v.
func Parse(t, v string) (Metric, error) {
	switch t {
	case Gauge(0).Type():
		res, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse gauge value: %w", err)
		}
		return Gauge(res), nil
	case Counter(0).Type():
		res, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse counter value: %w", err)
		}
		return Counter(res), nil
	default:
		return nil, fmt.Errorf("unknown metric type %s", t)
	}
}

// Gauge represents a gauge metric.
type Gauge float64

var _ Metric = Gauge(0)

func (g Gauge) String() string {
	return fmt.Sprintf("%g", g)
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
