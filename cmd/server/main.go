package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/router"
	"github.com/nekr0z/muhame/internal/storage"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	if err := run(flagNetAddress); err != nil {
		log.Fatal(err)
	}
}

func run(addr addr.NetAddress) error {
	st := storage.NewMemStorage()

	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	sugar := *logger.Sugar()

	r := router.New(logger, st)

	sugar.Infof("running server on %s", addr.String())
	return http.ListenAndServe(addr.String(), r)
}
