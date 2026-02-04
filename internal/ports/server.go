package ports

// HealthState represents the operational state of an HTTP server.
// It is used for health status reporting and startup/shutdown coordination.
type HealthState int32

const (
	// HealthStateStarting indicates the server is initializing but not yet ready to serve.
	HealthStateStarting HealthState = iota

	// HealthStateHealthy indicates the server is running and accepting requests.
	HealthStateHealthy

	// HealthStateShuttingDown indicates the server is in graceful shutdown mode.
	HealthStateShuttingDown

	// HealthStateUnhealthy indicates the server has encountered an error or is not operational.
	HealthStateUnhealthy
)
