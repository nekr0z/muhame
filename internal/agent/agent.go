// Package agent implements the metric-sending agent.
package agent

import (
	"context"
	"time"
)

// Run starts the agent to collect all metrics and send them to the server.
func Run(ctx context.Context, addr string, reportInterval, pollInterval time.Duration) error {
	q := &queue{}

	go collect(ctx, q, pollInterval)
	go send(ctx, q, addr, reportInterval)

	<-ctx.Done()
	return ctx.Err()
}

func collect(ctx context.Context, q *queue, interval time.Duration) {
	var counter int64
	for {
		select {
		case <-ctx.Done():
			return
		default:
			counter++
			collectMetrics(q, counter)
			time.Sleep(interval)
		}
	}
}

func send(ctx context.Context, q *queue, addr string, interval time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			q.sendMetrics(addr)
			time.Sleep(interval)
		}
	}
}
