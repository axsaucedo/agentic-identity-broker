package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

func resolveClientIP(r *http.Request, cfg ports.RequestContextConfig) string {
	if cfg.TrustedProxy.Enabled {
		return forwardedClientIP(r.Header.Get(cfg.TrustedProxy.ForwardedHeader))
	}

	return directClientIP(r.RemoteAddr)
}

func directClientIP(remoteAddr string) string {
	trimmed := strings.TrimSpace(remoteAddr)
	if trimmed == "" {
		return ""
	}

	host, _, err := net.SplitHostPort(trimmed)
	if err == nil {
		return normalizedIP(host)
	}

	return normalizedIP(trimmed)
}

func forwardedClientIP(headerValue string) string {
	if strings.TrimSpace(headerValue) == "" {
		return ""
	}

	parts := strings.Split(headerValue, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(parts[i])
		if candidate == "" {
			continue
		}
		return normalizedIP(candidate)
	}

	return ""
}

func normalizedIP(candidate string) string {
	if ip := net.ParseIP(candidate); ip != nil {
		return ip.String()
	}

	host, _, err := net.SplitHostPort(candidate)
	if err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
	}

	return ""
}
