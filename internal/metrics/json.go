package metrics

import (
	"encoding/json"
	"fmt"
)

// JSONMetric is a metric serialized into JSON format.
type JSONMetric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// Metric returns the Metric value.
func (j JSONMetric) Metric() (Metric, error) {
	switch j.MType {
	case Gauge(0).Type():
		if j.Value == nil {
			return nil, fmt.Errorf("gauge metric has no value")
		}
		return Gauge(*j.Value), nil
	case Counter(0).Type():
		if j.Delta == nil {
			return nil, fmt.Errorf("counter metric has no delta")
		}
		return Counter(*j.Delta), nil
	default:
		return nil, fmt.Errorf("unknown metric type %s", j.MType)
	}
}

// Named returns the Named metric.
func (j JSONMetric) Named() (Named, error) {
	n := Named{Name: j.ID}

	var err error

	n.Metric, err = j.Metric()

	return n, err
}

// ToJSON converts the metric to JSON format.
func ToJSON(m Metric, name string) []byte {
	var jm JSONMetric
	jm.ID = name

	switch m := m.(type) {
	case Gauge:
		jm.MType = Gauge(0).Type()
		v := float64(m)
		jm.Value = &v
	case Counter:
		jm.MType = Counter(0).Type()
		d := int64(m)
		jm.Delta = &d
	default:
		return nil
	}

	b, err := json.Marshal(jm)
	if err != nil {
		panic(err)
	}

	return b
}

// FromJSON unmarshals the JSON format to Metric.
func FromJSON(b []byte) (Named, error) {
	var jm JSONMetric
	var n Named

	err := json.Unmarshal(b, &jm)
	if err != nil {
		return n, err
	}

	n.Metric, err = jm.Metric()
	if err != nil {
		return n, err
	}

	n.Name = jm.ID

	return n, nil
}
