package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/router"
	"github.com/nekr0z/muhame/internal/storage"
	"go.uber.org/zap"
)

func ConfugureAndRun() error {
	cfg := configure()
	return run(cfg)
}

func run(cfg config) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	sugar := *logger.Sugar()

	st := storage.NewFileStorage(&sugar, cfg.st)

	server := &http.Server{
		Addr:    cfg.address.String(),
		Handler: router.New(logger, st),
	}

	serverChan := make(chan struct{}, 1)

	go func() {
		sugar.Infof("running server on %s", cfg.address.String())
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			sugar.Fatalf("HTTP service error: %s", err)
		}
		sugar.Info("HTTP service stopped")
		serverChan <- struct{}{}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		sugar.Info("Shutting down...")
	case <-serverChan:
		sugar.Info("Server stopped, will exit")
	}

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		sugar.Fatalf("HTTP shutdown error: %s", err)
	}

	st.Stop()

	sugar.Info("Shutdown complete.")

	return nil
}

type config struct {
	address addr.NetAddress
	st      storage.Config
}
