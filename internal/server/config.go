package server

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/nekr0z/muhame/internal/addr"
	confighelper "github.com/nekr0z/muhame/internal/config"
	"github.com/nekr0z/muhame/internal/crypt"
	"github.com/nekr0z/muhame/internal/storage"
)

type envConfig struct {
	Address       addr.NetAddress `env:"ADDRESS" json:"address"`
	StoreInterval int             `env:"STORE_INTERVAL" json:"store_interval"`
	Filename      string          `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore       bool            `env:"RESTORE" json:"restore"`
	DatabaseURL   string          `env:"DATABASE_DSN" json:"database_dsn"`
	Key           string          `env:"KEY" json:"key"`
	CryptoKey     string          `env:"CRYPTO_KEY" json:"crypto_key"`
	TrustedSubnet string          `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	GRPC          addr.NetAddress `env:"GRPC_ADDRESS" json:"grpc_address"`
}

func newConfig() config {
	cfg := envConfig{
		Address: addr.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		Restore:       true,
		StoreInterval: 300,
		Filename:      "metrics.sav",
	}

	confighelper.ConfigFromFile(&cfg)

	flags := flag.NewFlagSet("muhame-server", flag.ExitOnError)

	flags.Func("c", "config file", func(s string) error {
		return nil
	})
	flags.Func("config", "config file", func(s string) error {
		return nil
	})
	flags.Var(&cfg.Address, "a", "host:port to listen on")
	flags.IntVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "seconds between saving metrics to disk, 0 makes saving synchronous")
	flags.StringVar(&cfg.Filename, "f", cfg.Filename, "file to store metrics in")
	flags.BoolVar(&cfg.Restore, "r", cfg.Restore, "restore metrics from file on start")
	flags.StringVar(&cfg.DatabaseURL, "d", cfg.DatabaseURL, "database URL")
	flags.StringVar(&cfg.Key, "k", cfg.Key, "signing key")
	flags.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "private key for message decryption")
	flags.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "trusted subnet")
	flags.Var(&cfg.GRPC, "g", "host:port to use for gRPC")

	flags.Parse(os.Args[1:])

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
		signKey:       cfg.Key,
		trustedSubnet: cfg.TrustedSubnet,
		gRPCaddress:   cfg.GRPC,
	}

	c.privateKey, err = crypt.LoadPrivateKey(cfg.CryptoKey)
	if err != nil {
		c.privateKey = nil
	}

	return c
}
