package main

import (
	"log"

	"github.com/nekr0z/muhame/internal/server"
)

func main() {
	err := server.Run()
	if err != nil {
		log.Fatal(err)
	}
}
