// Package server implements the metric collection server.
package server

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/router"
	"github.com/nekr0z/muhame/internal/storage"
)

// Run creates and runs the server.
func Run() error {
	cfg := newConfig()
	return run(cfg)
}

func run(cfg config) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	sugar := *logger.Sugar()

	st, err := storage.New(&sugar, cfg.st)
	if err != nil {
		return fmt.Errorf("failed to set up storage: %w", err)
	}

	server := &http.Server{
		Addr:    cfg.address.String(),
		Handler: router.New(logger, st, cfg.signKey, cfg.privateKey),
	}

	serverChan := make(chan struct{}, 1)

	go func() {
		sugar.Infof("running server on %s", cfg.address.String())
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			sugar.Errorf("HTTP service error: %s", err)
		}
		sugar.Info("HTTP service stopped")
		serverChan <- struct{}{}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigChan:
		sugar.Info("Shutting down...")
	case <-serverChan:
		sugar.Info("Server stopped, will exit")
	}

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		sugar.Errorf("HTTP shutdown error: %s", err)
	}

	st.Close()

	sugar.Info("Shutdown complete.")

	return nil
}

type config struct {
	address    addr.NetAddress
	st         storage.Config
	signKey    string
	privateKey *rsa.PrivateKey
}
