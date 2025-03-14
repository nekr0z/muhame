package agent

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
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

func (q *queue) sendMetrics(c *http.Client, addr string) {
	mm := make([]queuedMetric, 0)

	for m := q.pop(); m != nil; m = q.pop() {
		mm = append(mm, *m)
	}

	if len(mm) == 0 {
		return
	}

	sendAll(c, addr, mm)
}

func sendAll(c *http.Client, addr string, mm []queuedMetric) {
	if sendBulk(c, addr, mm) == nil {
		return
	}

	for _, m := range mm {
		sendMetric(c, m, addr)
	}
}

func sendBulk(c *http.Client, addr string, mm []queuedMetric) error {
	b := zipBulk(mm)

	req, err := http.NewRequest(http.MethodPost, endpointBulk(addr), b)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	if resp == nil {
		return fmt.Errorf("nil response")
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bulk not accepted")
	}

	return nil
}

func zipBulk(mm []queuedMetric) *bytes.Buffer {
	var b bytes.Buffer

	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	bw := bufio.NewWriter(w)

	_, err = bw.WriteRune('[')
	if err != nil {
		panic(err)
	}

	for i, m := range mm {
		if i != 0 {
			_, err = bw.WriteRune(',')
			if err != nil {
				panic(err)
			}
		}

		_, err = bw.Write(metrics.ToJSON(m.val, m.name))
		if err != nil {
			panic(err)
		}
	}

	_, err = bw.WriteRune(']')
	if err != nil {
		panic(err)
	}

	err = bw.Flush()
	if err != nil {
		panic(err)
	}

	return &b
}

func sendMetric(c *http.Client, m queuedMetric, addr string) {
	bb := metrics.ToJSON(m.val, m.name)

	b := compress(bb)

	req, err := http.NewRequest(http.MethodPost, endpointSingle(addr), &b)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, _ := c.Do(req)
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

func endpointSingle(addr string) string {
	return strings.TrimSuffix(addr, "/") + "/update/"
}

func endpointBulk(addr string) string {
	return strings.TrimSuffix(addr, "/") + "/updates/"
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
