// Package agent implements the metric-sending agent.
package agent

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/httpclient"
)

type envConfig struct {
	Address        addr.NetAddress `env:"ADDRESS"`
	ReportInterval int             `env:"REPORT_INTERVAL"`
	PollInterval   int             `env:"POLL_INTERVAL"`
	Key            string          `env:"KEY"`
}

func New() Agent {
	cfg := envConfig{
		Address: addr.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
	}

	flag.Var(&cfg.Address, "a", "host:port to send metrics to")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "seconds between sending consecutive reports")
	flag.IntVar(&cfg.PollInterval, "p", 2, "seconds between acquiring metrics")
	flag.StringVar(&cfg.Key, "k", "", "signing key")

	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	return Agent{
		address:        cfg.Address,
		reportInterval: time.Duration(cfg.ReportInterval) * time.Second,
		pollInterval:   time.Duration(cfg.PollInterval) * time.Second,
		signKey:        cfg.Key,
	}
}

// Run starts the agent to collect all metrics and send them to the server.
func (a Agent) Run(ctx context.Context) error {
	if a.signKey != "" {
		log.Printf("using key \"%s\" to sign messages", a.signKey)
	}

	q := &queue{}

	go collect(ctx, q, a.pollInterval)
	go send(ctx, q, a.address, a.reportInterval, a.signKey)

	<-ctx.Done()
	return ctx.Err()
}

type Agent struct {
	address        addr.NetAddress
	reportInterval time.Duration
	pollInterval   time.Duration
	signKey        string
}

func (a Agent) Address() addr.NetAddress {
	return a.address
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

func send(ctx context.Context, q *queue, address addr.NetAddress, interval time.Duration, key string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			q.sendMetrics(httpclient.New().WithKey(key), address.StringWithProto())
			time.Sleep(interval)
		}
	}
}
