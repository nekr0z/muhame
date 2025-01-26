package main

import (
	"flag"
	"os"

	"github.com/nekr0z/muhame/internal/addr"
)

var flagNetAddress = addr.NetAddress{
	Host: "localhost",
	Port: 8080,
}

func parseFlags() {
	flag.Var(&flagNetAddress, "a", "host:port to listen on")

	flag.Parse()

	if env, ok := os.LookupEnv("ADDRESS"); ok {
		var envAddress addr.NetAddress
		err := envAddress.Set(env)
		if err == nil {
			flagNetAddress = envAddress
		}
	}
}
