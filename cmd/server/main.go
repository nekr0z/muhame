package main

import (
	"log"
	"net/http"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/router"
	"github.com/nekr0z/muhame/internal/storage"
)

func main() {
	parseFlags()

	if err := run(flagNetAddress); err != nil {
		log.Fatal(err)
	}
}

func run(addr addr.NetAddress) error {
	st := storage.NewMemStorage()

	r := router.New(st)

	log.Printf("running server on %s", addr.String())
	return http.ListenAndServe(addr.String(), r)
}
