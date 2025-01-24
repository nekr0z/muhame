package main

import (
	"context"
	"log"

	"github.com/nekr0z/muhame/internal/agent"
)

func main() {
	if err := agent.Run(context.TODO(), "http://localhost:8080"); err != nil {
		log.Fatal(err)
	}
}
