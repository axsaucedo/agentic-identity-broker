package telemetry

import "context"

// NoopShutdown is a no-op shutdown function returned when telemetry is disabled.
// It always returns nil and performs no cleanup operations.
func NoopShutdown(_ context.Context) error {
	return nil
}
