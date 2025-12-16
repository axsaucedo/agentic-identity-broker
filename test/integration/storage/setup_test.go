//go:build integration
// +build integration

package storage

import (
	"os"
	"os/exec"
)

func init() {
	// Configure testcontainers to work with Podman on macOS
	// Podman on macOS requires the machine to be running
	// testcontainers will automatically use Podman if Docker is not found

	// Set DOCKER_HOST to use Podman socket if it exists and DOCKER_HOST is not already set
	if os.Getenv("DOCKER_HOST") == "" {
		// Try common Podman socket locations
		podmanSockets := []string{
			// macOS Podman machine socket
			os.ExpandEnv("$HOME/.local/share/containers/podman/machine/podman.sock"),
			// Linux rootless Podman socket
			os.ExpandEnv("$XDG_RUNTIME_DIR/podman/podman.sock"),
			"/run/podman/podman.sock",
		}

		for _, socket := range podmanSockets {
			// Check if socket exists
			if _, err := os.Stat(socket); err == nil {
				os.Setenv("DOCKER_HOST", "unix://"+socket)
				break
			}
		}
	}

	// Enable Podman by checking for podman command and setting appropriate env
	if _, err := exec.LookPath("podman"); err == nil {
		// Podman is available, testcontainers should use it
		// On macOS, podman-machine needs to be running
		// The user should start it with: podman machine start
		//
		// On Linux with rootless Podman, the socket should be available automatically
		if os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE") == "" {
			// Don't set this - let testcontainers auto-detect
		}
	}
}
