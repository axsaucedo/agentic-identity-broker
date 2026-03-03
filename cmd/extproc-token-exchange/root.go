// Package main is the entry point for the ExtProc Token Exchange service.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"

	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
	extprocserver "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/server"
)

var rootCmd = &cobra.Command{
	Use:   "extproc-token-exchange",
	Short: "Envoy ExtProc Token Exchange Service",
	Long: `ExtProc Token Exchange Service implements the Envoy External Processor (ExtProc)
gRPC interface for transparent OAuth2 token exchange. It intercepts incoming
requests, exchanges Bearer tokens via RFC 8693, and replaces the Authorization
header before the request reaches the upstream service.`,
	RunE: run,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// run is the main execution function for the ExtProc Token Exchange service.
func run(cmd *cobra.Command, _ []string) error {
	// 1. Load config — CLI flags (highest) → EXTPROC_* env vars → YAML file → defaults.
	cfg, err := extprocconfig.LoadWithCommand(cmd)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// 2. Initialize logger.
	logger := initLogger(cfg)
	logger.Info("ExtProc Token Exchange Service starting",
		"grpc_bind", cfg.GRPC.Bind,
		"grpc_port", cfg.GRPC.Port,
		"token_endpoint", cfg.OAuth2.TokenEndpoint,
		"issuer", cfg.OAuth2.Issuer,
		"client_id", cfg.OAuth2.ClientID,
		"client_secret", "[REDACTED]")

	// 3. Setup signal handling for graceful shutdown.
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 4. Create TokenExchanger (performs startup client_credentials grant — fails fast on error).
	exchanger, err := extprocserver.NewTokenExchanger(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize token exchanger: %w", err)
	}
	defer exchanger.Shutdown()

	// 5. Create ExtProc server.
	svc := extprocserver.NewServer(cfg, exchanger, logger)

	// 6. Create gRPC server with max_concurrent_streams.
	grpcOpts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(uint32(cfg.GRPC.MaxConcurrentStreams)),
		grpc.KeepaliveParams(keepalive.ServerParameters{}),
	}
	grpcSrv := grpc.NewServer(grpcOpts...)
	extprocv3.RegisterExternalProcessorServer(grpcSrv, svc)

	// 7. Listen on configured bind:port.
	addr := fmt.Sprintf("%s:%d", cfg.GRPC.Bind, cfg.GRPC.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	logger.Info("gRPC server listening", "addr", addr)

	// 8. Serve in background goroutine.
	serveErr := make(chan error, 1)
	go func() {
		if err := grpcSrv.Serve(listener); err != nil {
			serveErr <- err
		}
	}()

	// 9. Wait for shutdown signal or serve error.
	select {
	case <-sigCtx.Done():
		logger.Info("Shutdown signal received, stopping gRPC server")
	case err := <-serveErr:
		return fmt.Errorf("gRPC server error: %w", err)
	}

	// 10. Graceful shutdown.
	grpcSrv.GracefulStop()
	logger.Info("ExtProc Token Exchange Service stopped")
	return nil
}

// initLogger creates a structured slog.Logger from the service configuration.
func initLogger(cfg *extprocconfig.Config) *slog.Logger {
	level := slog.LevelInfo
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	var handler slog.Handler
	if cfg.Log.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	return slog.New(handler)
}

func init() {
	extprocconfig.RegisterFlags(rootCmd)
}
