package server

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/crypt"
	"github.com/nekr0z/muhame/internal/storage"
)

type envConfig struct {
	Address       addr.NetAddress `env:"ADDRESS"`
	StoreInterval int             `env:"STORE_INTERVAL"`
	Filename      string          `env:"FILE_STORAGE_PATH"`
	Restore       bool            `env:"RESTORE"`
	DatabaseURL   string          `env:"DATABASE_DSN"`
	Key           string          `env:"KEY"`
	CryptoKey     string          `env:"CRYPTO_KEY"`
}

func newConfig() config {
	cfg := envConfig{
		Address: addr.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
	}

	flag.Var(&cfg.Address, "a", "host:port to listen on")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "seconds between saving metrics to disk, 0 makes saving synchronous")
	flag.StringVar(&cfg.Filename, "f", "metrics.sav", "file to store metrics in")
	flag.BoolVar(&cfg.Restore, "r", true, "restore metrics from file on start")
	flag.StringVar(&cfg.DatabaseURL, "d", "", "database URL")
	flag.StringVar(&cfg.Key, "k", "", "signing key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "private key for message decryption")

	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	c := config{
		address: cfg.Address,
		st: storage.Config{
			Interval:    time.Duration(cfg.StoreInterval) * time.Second,
			Filename:    cfg.Filename,
			Restore:     cfg.Restore,
			DatabaseDSN: cfg.DatabaseURL,
		},
		signKey: cfg.Key,
	}

	c.privateKey, err = crypt.LoadPrivateKey(cfg.CryptoKey)
	if err != nil {
		c.privateKey = nil
	}

	return c
}
