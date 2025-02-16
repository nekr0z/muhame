package server

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/nekr0z/muhame/internal/addr"
)

func configure() config {
	cfg := config{
		address: addr.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
	}

	flag.Var(&cfg.address, "a", "host:port to listen on")
	flagStoreInterval := flag.Int("i", 300, "seconds between saving metrics to disk, 0 makes saving synchronous")
	flag.StringVar(&cfg.st.Filename, "f", "metrics.sav", "file to store metrics in")
	flag.BoolVar(&cfg.st.Restore, "r", true, "restore metrics from file on start")

	flag.Parse()

	if env, ok := os.LookupEnv("ADDRESS"); ok {
		var envAddress addr.NetAddress
		err := envAddress.Set(env)
		if err == nil {
			cfg.address = envAddress
		}
	}

	if env, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		envStoreInterval, err := strconv.Atoi(env)
		if err != nil {
			flagStoreInterval = &envStoreInterval
		}
	}
	cfg.st.Interval = time.Duration(*flagStoreInterval) * time.Second

	if env, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.st.Filename = env
	}

	if env, ok := os.LookupEnv("RESTORE"); ok {
		envRestore, err := strconv.ParseBool(env)
		if err != nil {
			cfg.st.Restore = envRestore
		}
	}

	return cfg
}
