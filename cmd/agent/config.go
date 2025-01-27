package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/agent"
)

var flagNetAddress = addr.NetAddress{
	Host: "localhost",
	Port: 8080,
}

func configure() agent.Config {
	flag.Var(&flagNetAddress, "a", "host:port to send metrics to")
	flagReportInterval := flag.Int("r", 10, "seconds between sending consecutive reports")
	flagPollInterval := flag.Int("p", 2, "seconds between acquiring metrics")

	flag.Parse()

	if envVar, ok := os.LookupEnv("ADDRESS"); ok {
		var envAddress addr.NetAddress
		err := envAddress.Set(envVar)
		if err == nil {
			flagNetAddress = envAddress
		}
	}

	if env, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		envReportInterval, err := strconv.Atoi(env)
		if err != nil {
			flagReportInterval = &envReportInterval
		}
	}

	if env, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		envPollInterval, err := strconv.Atoi(env)
		if err != nil {
			flagPollInterval = &envPollInterval
		}
	}

	return agent.Config{
		Address:        flagNetAddress,
		ReportInterval: time.Duration(*flagReportInterval) * time.Second,
		PollInterval:   time.Duration(*flagPollInterval) * time.Second,
	}
}
