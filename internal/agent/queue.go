package agent

import (
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/nekr0z/muhame/internal/metrics"
)

// queue stores metrics queued for sending by agent.
type queue struct {
	sync.Mutex

	first *metric
	last  *metric
}

func (q *queue) push(m metric) {
	q.Lock()
	defer q.Unlock()

	m.next = nil

	if q.first == nil {
		q.first = &m
		q.last = q.first
		return
	}

	q.last.next = &m
	q.last = q.last.next
}

func (q *queue) pop() *metric {
	q.Lock()
	defer q.Unlock()

	if q.first == nil {
		return nil
	}

	m := q.first
	q.first = q.first.next
	m.next = nil
	return m
}

func (q *queue) sendMetrics(addr string) {
	for m := q.pop(); m != nil; m = q.pop() {
		metricType := "counter"

		if _, ok := m.m.(metrics.Gauge); ok {
			metricType = "gauge"
		}

		ep := endpoint(addr, metricType, m.name, m.m.String())

		r, err := http.NewRequest(http.MethodPost, ep, nil)
		if err != nil {
			panic(err)
		}

		r.Header.Add("Content-Type", "text/plain")

		client := &http.Client{}
		resp, err := client.Do(r)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
	}
}

type metric struct {
	name string
	m    metrics.Metric
	next *metric
}

func endpoint(addr string, metricType string, name string, value string) string {
	return strings.TrimSuffix(addr, "/") + "/" + path.Join("update", metricType, name, value)
}
