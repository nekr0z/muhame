package agent

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
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
		sendMetric(*m, addr)
	}
}

func sendMetric(m queuedMetric, addr string) {
	bb := metrics.ToJSON(m.val, m.name)

	b := compress(bb)

	req, err := http.NewRequest(http.MethodPost, endpoint(addr), &b)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, _ := http.DefaultClient.Do(req)
	// Error is ignored since increment #7 test expects us to just happily go
	// on, even if the response is breaking HTTP session.

	if resp == nil {
		return
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

type queuedMetric struct {
	name string
	val  metrics.Metric
	next *queuedMetric
}

func endpoint(addr string) string {
	return strings.TrimSuffix(addr, "/") + "/update/"
}

func compress(b []byte) bytes.Buffer {
	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		panic(err)
	}

	_, _ = w.Write(b)
	_ = w.Close()

	return buf
}
