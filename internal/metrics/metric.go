// Package metrics represents metrics.
package metrics

import "fmt"

// Metric represents a metric.
type Metric interface {
	String() string
}

// Gauge represents a gauge metric.
type Gauge float64

func (g Gauge) String() string {
	return fmt.Sprintf("%f", g)
}

// Counter represents a counter metric.
type Counter int64

func (c Counter) String() string {
	return fmt.Sprintf("%d", c)
}
