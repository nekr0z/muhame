package main

import (
	"log"

	"github.com/nekr0z/muhame/internal/server"
)

func main() {
	err := server.ConfugureAndRun()
	if err != nil {
		log.Fatal(err)
	}
}
