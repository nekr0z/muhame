package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/nekr0z/muhame/internal/addr"
)

var flagNetAddress = addr.NetAddress{
	Host: "localhost",
	Port: 8080,
}

var (
	flagPollInterval   int
	flagReportInterval int
)

func parseFlags() {
	flag.Var(&flagNetAddress, "a", "host:port to send metrics to")
	flag.IntVar(&flagReportInterval, "r", 10, "seconds between sending consecutive reports")
	flag.IntVar(&flagPollInterval, "p", 2, "seconds between acquiring metrics")

	flag.Parse()

	if env, ok := os.LookupEnv("ADDRESS"); ok {
		var envAddress addr.NetAddress
		err := envAddress.Set(env)
		if err == nil {
			flagNetAddress = envAddress
		}
	}

	if env, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		envReportInterval, err := strconv.Atoi(env)
		if err != nil {
			flagReportInterval = envReportInterval
		}
	}

	if env, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		envPollInterval, err := strconv.Atoi(env)
		if err != nil {
			flagPollInterval = envPollInterval
		}
	}
}
