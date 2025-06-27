// Package server implements the metric collection server.
package server

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/grpcserver"
	"github.com/nekr0z/muhame/internal/router"
	"github.com/nekr0z/muhame/internal/storage"
	"github.com/nekr0z/muhame/pkg/proto"
)

// Run creates and runs the server.
func Run(ctx context.Context) error {
	fmt.Printf("running")

	cfg := newConfig()

	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	cfg.log = logger

	return run(ctx, cfg)
}

func run(ctx context.Context, cfg config) error {
	sugar := *cfg.log.Sugar()

	st, err := storage.New(&sugar, cfg.st)
	if err != nil {
		return fmt.Errorf("failed to set up storage: %w", err)
	}

	serverChan := make(chan struct{}, 1)

	var grpcServer *grpc.Server

	useGRPC := cfg.gRPCaddress.Port != 0

	httpServer := &http.Server{
		Addr:    cfg.address.String(),
		Handler: router.New(cfg.log, st, cfg.signKey, cfg.privateKey, cfg.trustedSubnet),
	}

	go func() {
		sugar.Infof("running server on %s", cfg.address.String())
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			sugar.Errorf("HTTP service error: %s", err)
		}

		sugar.Info("HTTP service stopped")
		close(serverChan)
	}()

	grpcChan := make(chan struct{}, 1)

	if useGRPC {
		go func() {
			listen, err := net.Listen("tcp", cfg.gRPCaddress.String())
			if err != nil {
				sugar.Errorf("failed to listen on %s: %w", cfg.gRPCaddress.String(), err)
			}

			grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(
				grpcserver.SignatureInterceptor(cfg.signKey),
				grpcserver.DecryptInterceptor(cfg.privateKey),
			))

			proto.RegisterMetricsServiceServer(grpcServer, grpcserver.New(st))

			sugar.Infof("running gRPC server on %s", cfg.gRPCaddress.String())
			if err := grpcServer.Serve(listen); err != nil {
				sugar.Errorf("gRPC service error: %s", err)
			}

			sugar.Info("gRPC service stopped")

			useGRPC = false

			close(grpcChan)
		}()
	} else {
		defer close(grpcChan)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigChan:
		sugar.Info("Shutting down...")
	case <-serverChan:
		sugar.Info("Server stopped, will exit")
	case <-grpcChan:
		sugar.Info("gRPC server stopped, will exit")
	case <-ctx.Done():
		sugar.Info("Context cancelled, will exit")
	}

	if useGRPC {
		grpcServer.GracefulStop()
	}

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownRelease()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		sugar.Errorf("HTTP shutdown error: %s", err)
	}

	st.Close()

	sugar.Info("Shutdown complete.")

	return nil
}

type config struct {
	log           *zap.Logger
	address       addr.NetAddress
	st            storage.Config
	signKey       string
	privateKey    *rsa.PrivateKey
	trustedSubnet string
	gRPCaddress   addr.NetAddress
}
