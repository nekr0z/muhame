// Package agent implements the metric-sending agent.
package agent

import (
	"context"
	"time"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/httpclient"
)

// Run starts the agent to collect all metrics and send them to the server.
func Run(ctx context.Context, config Config) error {
	q := &queue{}

	go collect(ctx, q, config.PollInterval)
	go send(ctx, q, config.Address, config.ReportInterval)

	<-ctx.Done()
	return ctx.Err()
}

// Config configures the agent.
type Config struct {
	Address        addr.NetAddress
	ReportInterval time.Duration
	PollInterval   time.Duration
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

func send(ctx context.Context, q *queue, address addr.NetAddress, interval time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			q.sendMetrics(httpclient.New(), address.StringWithProto())
			time.Sleep(interval)
		}
	}
}
