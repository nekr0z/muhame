package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nekr0z/muhame/internal/server"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func init() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func main() {
	err := server.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
