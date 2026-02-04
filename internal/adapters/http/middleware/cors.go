package middleware

import (
	"net/http"
)

// CORSMiddleware returns a middleware that sets CORS headers for API routes.
// Allows cross-origin requests to consent management APIs.
// This is required for the SPA to make API calls when served from a different origin.
func CORSMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			// Allow all origins for development (should be configured per environment in production)
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Custom-Principal")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Handle preflight OPTIONS request
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}
