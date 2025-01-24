package agent

import (
	"io"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/nekr0z/muhame/internal/metrics"
)

// queue stores metrics queued for sending by agent.
type queue struct {
	sync.Mutex

	first *queuedMetric
	last  *queuedMetric
}

func (q *queue) push(m queuedMetric) {
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

func (q *queue) pop() *queuedMetric {
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

		if _, ok := m.val.(metrics.Gauge); ok {
			metricType = "gauge"
		}

		ep := endpoint(addr, metricType, m.name, m.val.String())

		resp, err := http.Post(ep, "text/plain", nil)
		if err != nil {
			panic(err)
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

type queuedMetric struct {
	name string
	val  metricValue
	next *queuedMetric
}

type metricValue interface {
	String() string
}

func endpoint(addr string, metricType string, name string, value string) string {
	return strings.TrimSuffix(addr, "/") + "/" + path.Join("update", metricType, name, value)
}
