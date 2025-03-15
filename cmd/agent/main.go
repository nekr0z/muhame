package main

import (
	"context"
	"log"

	"github.com/nekr0z/muhame/internal/agent"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := agent.New()

	addr := a.Address()
	log.Printf("running and sending metrics to %s", addr.String())

	if err := a.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
