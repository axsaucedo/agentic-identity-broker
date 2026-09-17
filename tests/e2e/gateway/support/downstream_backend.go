package support

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const testcontainersHost = "host.testcontainers.internal"

// DownstreamBackend is a container-reachable streamable MCP backend that
// records every Authorization header it receives.
type DownstreamBackend struct {
	listener net.Listener
	server   *http.Server
	url      string

	mu                   sync.RWMutex
	requestCount         int
	authorizationHeaders []string
	closeErr             error
	closeOnce            sync.Once
}

// StartDownstreamBackend starts a streamable MCP backend on an address that
// agentgateway containers can reach through Testcontainers host access.
func StartDownstreamBackend() (*DownstreamBackend, error) {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, fmt.Errorf("listen on downstream MCP backend port: %w", err)
	}

	backend := &DownstreamBackend{
		listener: listener,
		url: "http://" + net.JoinHostPort(
			testcontainersHost,
			strconv.Itoa(listener.Addr().(*net.TCPAddr).Port),
		) + "/mcp",
	}
	mcpBackend := mcpserver.NewMCPServer(
		"agentgateway-downstream-backend", "1.0",
		mcpserver.WithToolCapabilities(false),
	)
	mcpBackend.AddTool(
		mcp.NewTool("whoami", mcp.WithDescription("Returns the Authorization header received by the backend.")),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			authorization, _ := ctx.Value(downstreamAuthorizationHeaderKey{}).(string)
			return mcp.NewToolResultText(authorization), nil
		},
	)
	backend.server = &http.Server{Handler: mcpserver.NewStreamableHTTPServer(
		mcpBackend,
		mcpserver.WithEndpointPath("/mcp"),
		// agentgateway connects over the local host-access tunnel but preserves
		// host.testcontainers.internal as the request host.
		mcpserver.WithDisableLocalhostProtection(true),
		mcpserver.WithHTTPContextFunc(func(ctx context.Context, request *http.Request) context.Context {
			authorization := request.Header.Get("Authorization")
			backend.record(authorization)
			return context.WithValue(ctx, downstreamAuthorizationHeaderKey{}, authorization)
		}),
	)}

	go backend.serve()
	return backend, nil
}

// URL returns the container-reachable MCP endpoint.
func (b *DownstreamBackend) URL() string {
	if b == nil {
		return ""
	}
	return b.url
}

// AuthorizationHeaders returns a copy of headers observed in request order.
func (b *DownstreamBackend) AuthorizationHeaders() []string {
	if b == nil {
		return nil
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return append([]string(nil), b.authorizationHeaders...)
}

// RequestCount returns the number of HTTP requests that reached the backend.
func (b *DownstreamBackend) RequestCount() int {
	if b == nil {
		return 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.requestCount
}

// Reset clears all request observations.
func (b *DownstreamBackend) Reset() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.requestCount = 0
	b.authorizationHeaders = nil
}

// Close stops the backend. It is safe to call more than once.
func (b *DownstreamBackend) Close() error {
	if b == nil {
		return nil
	}
	b.closeOnce.Do(func() {
		if b.server == nil {
			return
		}
		if err := b.server.Close(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			b.setCloseError(fmt.Errorf("close downstream MCP backend: %w", err))
		}
	})

	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.closeErr
}

func (b *DownstreamBackend) serve() {
	if err := b.server.Serve(b.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		b.setCloseError(fmt.Errorf("serve downstream MCP backend: %w", err))
	}
}

func (b *DownstreamBackend) record(authorization string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.requestCount++
	b.authorizationHeaders = append(b.authorizationHeaders, authorization)
}

func (b *DownstreamBackend) setCloseError(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closeErr == nil {
		b.closeErr = err
	}
}

// ExtProcStandIn is a container-reachable TCP listener that counts connection
// attempts without starting an ExtProc service.
type ExtProcStandIn struct {
	listener net.Listener

	mu          sync.RWMutex
	connections int
	closeErr    error
	closeOnce   sync.Once
}

// StartExtProcStandIn starts a TCP listener suitable for detecting unexpected
// ExtProc connections from an agentgateway container.
func StartExtProcStandIn() (*ExtProcStandIn, error) {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, fmt.Errorf("listen on ExtProc stand-in port: %w", err)
	}

	standIn := &ExtProcStandIn{listener: listener}
	go standIn.acceptConnections()
	return standIn, nil
}

// ConnectionCount returns the number of connections accepted by the stand-in.
func (s *ExtProcStandIn) ConnectionCount() int {
	if s == nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connections
}

// Close stops the stand-in. It is safe to call more than once.
func (s *ExtProcStandIn) Close() error {
	if s == nil {
		return nil
	}
	s.closeOnce.Do(func() {
		if s.listener == nil {
			return
		}
		if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			s.setCloseError(fmt.Errorf("close ExtProc stand-in: %w", err))
		}
	})

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closeErr
}

func (s *ExtProcStandIn) acceptConnections() {
	for {
		connection, err := s.listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				s.setCloseError(fmt.Errorf("accept ExtProc stand-in connection: %w", err))
			}
			return
		}

		s.mu.Lock()
		s.connections++
		s.mu.Unlock()
		_ = connection.Close()
	}
}

func (s *ExtProcStandIn) setCloseError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closeErr == nil {
		s.closeErr = err
	}
}

func (s *ExtProcStandIn) port() int {
	if s == nil || s.listener == nil {
		return 0
	}
	return s.listener.Addr().(*net.TCPAddr).Port
}

type downstreamAuthorizationHeaderKey struct{}
