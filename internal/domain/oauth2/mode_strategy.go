package oauth2

import (
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/servermode"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// ModeStrategy determines whether a classified agent is permitted in the active OAuth server mode.
// Implementations are selected at startup by the builder — no runtime mode checks in handlers.
type ModeStrategy interface {
	AcceptsClientType(ct storage.ClientType) bool
	Mode() servermode.Mode
}

type proxyModeStrategy struct{}

// NewProxyModeStrategy returns a ModeStrategy that accepts only ProxyClient agents.
func NewProxyModeStrategy() ModeStrategy { return &proxyModeStrategy{} }

func (s *proxyModeStrategy) AcceptsClientType(ct storage.ClientType) bool {
	return ct == storage.ProxyClient
}

func (s *proxyModeStrategy) Mode() servermode.Mode { return servermode.Proxy }

type localModeStrategy struct{}

// NewLocalModeStrategy returns a ModeStrategy that accepts CIMDClient and LocalClient agents.
func NewLocalModeStrategy() ModeStrategy { return &localModeStrategy{} }

func (s *localModeStrategy) AcceptsClientType(ct storage.ClientType) bool {
	return ct == storage.CIMDClient || ct == storage.LocalClient
}

func (s *localModeStrategy) Mode() servermode.Mode { return servermode.Local }

type hybridModeStrategy struct{}

// NewHybridModeStrategy returns a ModeStrategy that accepts all unambiguous client types.
func NewHybridModeStrategy() ModeStrategy { return &hybridModeStrategy{} }

func (s *hybridModeStrategy) AcceptsClientType(ct storage.ClientType) bool {
	return ct == storage.ProxyClient || ct == storage.CIMDClient || ct == storage.LocalClient
}

func (s *hybridModeStrategy) Mode() servermode.Mode { return servermode.Hybrid }
