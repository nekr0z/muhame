package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nekr0z/muhame/internal/agent"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parseFlags()

	log.Printf("running and sending metrics to %s", flagNetAddress.String())

	if err := agent.Run(
		ctx,
		fmt.Sprintf("http://%s", flagNetAddress.String()),
		time.Second*time.Duration(flagReportInterval),
		time.Second*time.Duration(flagPollInterval)); err != nil {
		log.Fatal(err)
	}
}
