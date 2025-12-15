// Package server contains domain logic for dual-server coordination.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"golang.org/x/sync/errgroup"
)

// Manager coordinates the lifecycle of dual HTTP servers (end-user and admin).
// Implements atomic startup: both servers must successfully bind and start,
// or neither server runs (fail-fast behavior).
type Manager struct {
	enduserServer   ports.ServerPort
	adminServer     ports.ServerPort
	shutdownTimeout time.Duration
	logger          *slog.Logger
}

// NewManager creates a new server manager with the provided server instances.
func NewManager(enduserServer, adminServer ports.ServerPort, shutdownTimeout time.Duration, logger *slog.Logger) *Manager {
	return &Manager{
		enduserServer:   enduserServer,
		adminServer:     adminServer,
		shutdownTimeout: shutdownTimeout,
		logger:          logger,
	}
}

// Start initiates atomic startup of both servers using errgroup.
// Both servers attempt to bind concurrently. If both succeed, they begin serving.
// If either fails, the context is cancelled and both are torn down.
//
// This is a blocking call that returns when:
// - Both servers shut down gracefully
// - Either server encounters an error
// - The context is cancelled
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("Starting dual-server manager", "mode", "atomic")

	// Use errgroup for concurrent startup with automatic context cancellation
	g, ctx := errgroup.WithContext(ctx)

	// Phase 1: Concurrent Listen (bind to ports)
	m.logger.Debug("Phase 1: Concurrent port binding")

	var enduserListener, adminListener net.Listener

	// Bind end-user server
	g.Go(func() error {
		m.logger.Debug("Binding end-user server", "name", m.enduserServer.Name())
		listener, err := m.enduserServer.Listen()
		if err != nil {
			m.logger.Error("End-user server bind failed", "error", err)
			return fmt.Errorf("enduser server bind failed: %w", err)
		}
		enduserListener = listener
		m.logger.Info("End-user server bound successfully", "name", m.enduserServer.Name())
		return nil
	})

	// Bind admin server
	g.Go(func() error {
		m.logger.Debug("Binding admin server", "name", m.adminServer.Name())
		listener, err := m.adminServer.Listen()
		if err != nil {
			m.logger.Error("Admin server bind failed", "error", err)
			return fmt.Errorf("admin server bind failed: %w", err)
		}
		adminListener = listener
		m.logger.Info("Admin server bound successfully", "name", m.adminServer.Name())
		return nil
	})

	// Wait for both binds to complete
	if err := g.Wait(); err != nil {
		m.logger.Error("Atomic startup failed during bind phase", "error", err)
		return err
	}

	m.logger.Info("Phase 1 complete: Both servers bound successfully")

	// Phase 2: Concurrent Serve (start accepting connections)
	m.logger.Debug("Phase 2: Starting HTTP servers")

	g, ctx = errgroup.WithContext(ctx)

	// Serve end-user server
	g.Go(func() error {
		m.logger.Info("Starting end-user server", "name", m.enduserServer.Name())
		if err := m.enduserServer.Serve(ctx, enduserListener); err != nil {
			m.logger.Error("End-user server failed", "error", err)
			return fmt.Errorf("enduser server failed: %w", err)
		}
		return nil
	})

	// Serve admin server
	g.Go(func() error {
		m.logger.Info("Starting admin server", "name", m.adminServer.Name())
		if err := m.adminServer.Serve(ctx, adminListener); err != nil {
			m.logger.Error("Admin server failed", "error", err)
			return fmt.Errorf("admin server failed: %w", err)
		}
		return nil
	})

	// Wait for both servers to complete (or first error)
	if err := g.Wait(); err != nil {
		m.logger.Error("Server runtime error", "error", err)
		return err
	}

	m.logger.Info("Both servers shut down cleanly")
	return nil
}

// Shutdown initiates graceful shutdown of both servers.
// Shutdowns run concurrently (parallel) with independent timeouts.
// Returns the first error encountered, or nil if both succeed.
func (m *Manager) Shutdown(parentCtx context.Context) error {
	m.logger.Info("Initiating graceful shutdown of both servers", "timeout", m.shutdownTimeout)

	// Use errgroup without context (we want both shutdowns to complete independently)
	g := new(errgroup.Group)

	// Shutdown end-user server
	g.Go(func() error {
		m.logger.Debug("Shutting down end-user server", "name", m.enduserServer.Name())
		if err := m.enduserServer.Shutdown(parentCtx, m.shutdownTimeout); err != nil {
			m.logger.Error("End-user server shutdown error", "error", err)
			return fmt.Errorf("enduser server shutdown failed: %w", err)
		}
		m.logger.Info("End-user server shut down successfully", "name", m.enduserServer.Name())
		return nil
	})

	// Shutdown admin server
	g.Go(func() error {
		m.logger.Debug("Shutting down admin server", "name", m.adminServer.Name())
		if err := m.adminServer.Shutdown(parentCtx, m.shutdownTimeout); err != nil {
			m.logger.Error("Admin server shutdown error", "error", err)
			return fmt.Errorf("admin server shutdown failed: %w", err)
		}
		m.logger.Info("Admin server shut down successfully", "name", m.adminServer.Name())
		return nil
	})

	// Wait for both shutdowns to complete
	if err := g.Wait(); err != nil {
		m.logger.Error("Shutdown completed with errors", "error", err)
		return err
	}

	m.logger.Info("All servers shut down successfully")
	return nil
}
