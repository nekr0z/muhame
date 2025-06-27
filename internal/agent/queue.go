package agent

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/nekr0z/muhame/internal/httpclient"
	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/pkg/proto"
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

func (q *queue) sendMetricsHTTP(c httpclient.Client, addr string) {
	mm := q.popAll()

	if len(mm) == 0 {
		return
	}

	sendAllHTTP(c, addr, mm)
}

func (q *queue) sendMetricsGRPC(ctx context.Context, c proto.MetricsServiceClient) {
	mm := q.popAll()

	if len(mm) == 0 {
		return
	}

	sendAllGRPC(ctx, c, mm)
}

func (q *queue) popAll() []queuedMetric {
	mm := make([]queuedMetric, 0)

	for m := q.pop(); m != nil; m = q.pop() {
		mm = append(mm, *m)
	}

	return mm
}

func sendAllHTTP(c httpclient.Client, addr string, mm []queuedMetric) {
	if sendBulkHTTP(c, addr, mm) == nil {
		return
	}

	for _, m := range mm {
		sendMetricHTTP(c, m, addr)
	}
}

func sendAllGRPC(ctx context.Context, c proto.MetricsServiceClient, mm []queuedMetric) {
	pm := make([]*proto.Metric, len(mm))

	for i, m := range mm {
		pm[i] = queuedMetricToProto(m)
	}

	if _, err := c.BulkUpdate(ctx, &proto.BulkRequest{
		Payload: &proto.BulkRequest_Metrics{
			Metrics: &proto.Metrics{
				Metrics: pm,
			},
		},
	}); err == nil {
		return
	}

	for _, met := range pm {
		// error is ignored to match HTTP behavior
		_, _ = c.Update(ctx, &proto.MetricRequest{
			Payload: &proto.MetricRequest_Metric{
				Metric: met,
			},
		})
	}
}

func sendBulkHTTP(c httpclient.Client, addr string, mm []queuedMetric) error {
	b := zipBulk(mm)

	code, err := c.Send(b.Bytes(), endpointBulk(addr))
	if err != nil {
		return err
	}

	if code != http.StatusOK {
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
	defer func() {
		err = w.Close()
		if err != nil {
			panic(err)
		}
	}()

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

func sendMetricHTTP(c httpclient.Client, m queuedMetric, addr string) {
	bb := metrics.ToJSON(m.val, m.name)

	b := compress(bb)

	_, _ = c.Send(b.Bytes(), endpointSingle(addr))
	// Error is ignored since increment #7 test expects us to just happily go
	// on, even if the response is breaking HTTP session.
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

func queuedMetricToProto(m queuedMetric) *proto.Metric {
	switch v := m.val.(type) {
	case metrics.Counter:
		return &proto.Metric{
			Name: m.name,
			Value: &proto.Metric_Counter{
				Counter: &proto.Counter{
					Delta: int64(v),
				},
			},
		}
	case metrics.Gauge:
		return &proto.Metric{
			Name: m.name,
			Value: &proto.Metric_Gauge{
				Gauge: &proto.Gauge{
					Value: float64(v),
				},
			},
		}
	default:
		return nil
	}
}
