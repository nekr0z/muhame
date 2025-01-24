// Package agent implements the metric-sending agent.
package agent

import (
	"context"
	"time"
)

// Run starts the agent to collect all metrics and send them to the server.
func Run(ctx context.Context, addr string) error {
	q := &queue{}

	go collect(ctx, q)
	go send(ctx, q, addr)

	<-ctx.Done()
	return ctx.Err()
}

func collect(ctx context.Context, q *queue) {
	var counter int64
	for {
		select {
		case <-ctx.Done():
			return
		default:
			counter++
			collectMetrics(q, counter)
			time.Sleep(2 * time.Second)
		}
	}
}

func send(ctx context.Context, q *queue, addr string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			q.sendMetrics(addr)
			time.Sleep(10 * time.Second)
		}
	}
}
