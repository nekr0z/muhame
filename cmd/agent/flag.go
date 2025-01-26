package main

import (
	"flag"

	"github.com/nekr0z/muhame/internal/addr"
)

var flagNetAddress = addr.NetAddress{
	Host: "localhost",
	Port: 8080,
}

var (
	pollInterval   int
	reportInterval int
)

func init() {
	flag.Var(&flagNetAddress, "a", "host:port to send metrics to")
	flag.IntVar(&reportInterval, "r", 10, "seconds between sending consecutive reports")
	flag.IntVar(&pollInterval, "p", 2, "seconds between acquiring metrics")
}
