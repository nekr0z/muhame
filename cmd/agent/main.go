package main

import (
	"context"
	"log"

	"github.com/nekr0z/muhame/internal/agent"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := agent.Run(ctx, "http://localhost:8080"); err != nil {
		log.Fatal(err)
	}
}
