package main

import (
	"flag"

	"github.com/nekr0z/muhame/internal/addr"
)

var flagNetAddress = addr.NetAddress{
	Host: "localhost",
	Port: 8080,
}

func init() {
	flag.Var(&flagNetAddress, "a", "host:port to listen on")
}
