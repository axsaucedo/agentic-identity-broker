package middleware

import (
	"net/http"

	"github.com/go-chi/cors"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// defaultCORSMethods is the list of allowed HTTP methods used when none are configured.
var defaultCORSMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}

// defaultCORSHeaders is the list of allowed request headers used when none are configured.
var defaultCORSHeaders = []string{"Content-Type", "Authorization"}

// defaultCORSMaxAge is the default preflight cache duration in seconds (24 hours).
const defaultCORSMaxAge = 86400

// CORSMiddleware returns a middleware that sets CORS headers for API routes using
// github.com/go-chi/cors for standards-compliant handling.
//
// When cfg.AllowedOrigins is empty, no CORS headers are added (production default — secure by default).
// This is required for the SPA to make API calls when served from a different origin (development).
func CORSMiddleware(cfg ports.CORSConfig) func(next http.Handler) http.Handler {
	if len(cfg.AllowedOrigins) == 0 {
		// No origins configured: pass through without any CORS headers (production default).
		return func(next http.Handler) http.Handler { return next }
	}

	methods := cfg.AllowedMethods
	if len(methods) == 0 {
		methods = defaultCORSMethods
	}

	headers := cfg.AllowedHeaders
	if len(headers) == 0 {
		headers = defaultCORSHeaders
	}

	maxAge := cfg.MaxAge
	if maxAge == 0 {
		maxAge = defaultCORSMaxAge
	}

	return cors.Handler(cors.Options{
		AllowedOrigins: cfg.AllowedOrigins,
		AllowedMethods: methods,
		AllowedHeaders: headers,
		MaxAge:         maxAge,
	})
}
