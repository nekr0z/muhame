package main

import (
	"context"
	"log"

	"github.com/nekr0z/muhame/internal/agent"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := configure()

	log.Printf("running and sending metrics to %s", flagNetAddress.String())

	if err := agent.Run(ctx, config); err != nil {
		log.Fatal(err)
	}
}
