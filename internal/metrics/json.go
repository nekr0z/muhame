package metrics

import (
	"encoding/json"
	"fmt"
)

type JSONMetric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

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
