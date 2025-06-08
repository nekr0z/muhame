// Package agent implements the metric-sending agent.
package agent

import (
	"context"
	"crypto/rsa"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/nekr0z/muhame/internal/addr"
	confighelper "github.com/nekr0z/muhame/internal/config"
	"github.com/nekr0z/muhame/internal/crypt"
	"github.com/nekr0z/muhame/internal/httpclient"
)

type envConfig struct {
	Address        addr.NetAddress `env:"ADDRESS" json:"address"`
	ReportInterval int             `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval   int             `env:"POLL_INTERVAL" json:"poll_interval"`
	Key            string          `env:"KEY" json:"key"`
	RateLimit      int             `env:"RATE_LIMIT" json:"rate_limit"`
	CryptoKey      string          `env:"CRYPTO_KEY" json:"crypto_key"`
}

// Agent is the metric-sending agent.
type Agent struct {
	address        addr.NetAddress
	reportInterval time.Duration
	pollInterval   time.Duration
	signKey        string
	workers        int

	pubKey *rsa.PublicKey

	q      *queue
	workCh chan struct{}
	wg     *sync.WaitGroup
}

// New creates a new agent.
func New() Agent {
	cfg := envConfig{
		Address: addr.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		ReportInterval: 10,
		PollInterval:   2,
		RateLimit:      1,
	}

	confighelper.ConfigFromFile(&cfg)

	flags := flag.NewFlagSet("muhame-agent", flag.ExitOnError)

	flags.Func("c", "config file", func(s string) error {
		return nil
	})
	flags.Func("config", "config file", func(s string) error {
		return nil
	})
	flags.Var(&cfg.Address, "a", "host:port to send metrics to")
	flags.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "seconds between sending consecutive reports")
	flags.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "seconds between acquiring metrics")
	flags.StringVar(&cfg.Key, "k", cfg.Key, "signing key")
	flags.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "simultaneous requests")
	flags.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "public key for message encryption")

	flags.Parse(os.Args[1:])

	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	a := Agent{
		address:        cfg.Address,
		reportInterval: time.Duration(cfg.ReportInterval) * time.Second,
		pollInterval:   time.Duration(cfg.PollInterval) * time.Second,
		signKey:        cfg.Key,
		workers:        cfg.RateLimit,
		q:              &queue{},
		workCh:         make(chan struct{}),
		wg:             &sync.WaitGroup{},
	}

	a.pubKey, err = crypt.LoadPublicKey(cfg.CryptoKey)
	if err != nil {
		a.pubKey = nil
	}

	return a
}

// Run starts the agent to collect all metrics and send them to the server.
func (a Agent) Run() {
	log.Printf("running and sending metrics to %s", a.address.String())
	if a.signKey != "" {
		log.Printf("using key \"%s\" to sign messages", a.signKey)
	}

	ctx, cancel := context.WithCancel(context.Background())

	a.wg.Add(a.workers)
	for range a.workers {
		go a.worker(ctx)
	}

	a.wg.Add(3)
	go a.collectBasic(ctx)
	go a.collectAux(ctx)
	go a.send(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Print("Shutting down...")
	cancel()

	a.wg.Wait()
	log.Print("Done.")
}

func (a Agent) worker(ctx context.Context) {
	defer a.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.workCh:
			a.q.sendMetrics(httpclient.New().WithKey(a.signKey).WithCrypto(a.pubKey), a.address.StringWithProto())
		}
	}
}

func (a Agent) collectBasic(ctx context.Context) {
	defer a.wg.Done()
	var counter int64
	for {
		select {
		case <-ctx.Done():
			return
		default:
			counter++
			collectBasicMetrics(a.q, counter)
			time.Sleep(a.pollInterval)
		}
	}
}

func (a Agent) collectAux(ctx context.Context) {
	defer a.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			collectAuxMetrics(a.q)
			time.Sleep(a.pollInterval)
		}
	}
}

func (a Agent) send(ctx context.Context) {
	defer a.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			a.workCh <- struct{}{}
			time.Sleep(a.reportInterval)
		}
	}
}
