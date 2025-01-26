package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/nekr0z/muhame/internal/agent"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()

	log.Printf("running and sending metrics to %s", flagNetAddress.String())

	if err := agent.Run(
		ctx,
		fmt.Sprintf("http://%s", flagNetAddress.String()),
		time.Second*time.Duration(reportInterval),
		time.Second*time.Duration(pollInterval)); err != nil {
		log.Fatal(err)
	}
}
