package oauth2

import "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"

// ModeStrategy determines whether a classified agent is permitted in the active OAuth server mode.
// Implementations are selected at startup by the builder — no runtime mode checks in handlers.
type ModeStrategy interface {
	AcceptsClientMode(mode storage.ClientMode) bool
	Name() string
}

type proxyModeStrategy struct{}

// NewProxyModeStrategy returns a ModeStrategy that accepts only ProxyClient agents.
func NewProxyModeStrategy() ModeStrategy { return &proxyModeStrategy{} }

func (s *proxyModeStrategy) AcceptsClientMode(mode storage.ClientMode) bool {
	return mode == storage.ProxyClient
}

func (s *proxyModeStrategy) Name() string { return "proxy" }

type localModeStrategy struct{}

// NewLocalModeStrategy returns a ModeStrategy that accepts CIMDClient and LocalClient agents.
func NewLocalModeStrategy() ModeStrategy { return &localModeStrategy{} }

func (s *localModeStrategy) AcceptsClientMode(mode storage.ClientMode) bool {
	return mode == storage.CIMDClient || mode == storage.LocalClient
}

func (s *localModeStrategy) Name() string { return "local" }

type hybridModeStrategy struct{}

// NewHybridModeStrategy returns a ModeStrategy that accepts all unambiguous client modes.
func NewHybridModeStrategy() ModeStrategy { return &hybridModeStrategy{} }

func (s *hybridModeStrategy) AcceptsClientMode(mode storage.ClientMode) bool {
	return mode != storage.AmbiguousClient
}

func (s *hybridModeStrategy) Name() string { return "hybrid" }
