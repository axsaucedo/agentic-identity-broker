// Package server — circuit_breaker.go wraps github.com/sony/gobreaker/v2 to
// protect outbound HTTP calls to the identity broker. It prevents a
// thundering herd when the broker recovers after an outage by fast-failing
// requests while the circuit is open.
package server

import (
	"errors"
	"log/slog"

	"github.com/sony/gobreaker/v2"

	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
)

// ErrCircuitOpen is returned by Exchange when the circuit breaker is open.
// Callers should surface this as a 503 Service Unavailable to signal a transient
// infrastructure failure rather than a client error.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// newGobreakerCB creates a gobreaker CircuitBreaker configured from cfg.
// The returned breaker uses consecutive-failure counting: it opens after
// cfg.CircuitBreaker.MaxFailures consecutive failures and allows a single
// probe after cfg.CircuitBreaker.ResetTimeout.
func newGobreakerCB(cfg *extprocconfig.Config, logger *slog.Logger) *gobreaker.CircuitBreaker[string] {
	maxFail := uint32(cfg.CircuitBreaker.MaxFailures) //nolint:gosec // validated positive by config.Validate
	timeout := cfg.CircuitBreaker.ResetTimeout

	st := gobreaker.Settings{
		Name:        "token-exchange",
		MaxRequests: 1, // only one probe allowed in half-open state
		Timeout:     timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= maxFail
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Info("circuit breaker state changed",
				"circuit", name,
				"from", from.String(),
				"to", to.String())
		},
	}

	return gobreaker.NewCircuitBreaker[string](st)
}
