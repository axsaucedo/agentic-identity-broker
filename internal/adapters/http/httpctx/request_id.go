// Package httpctx holds neutral HTTP request-scoped context values shared across the HTTP
// adapter packages (middleware, handlers). It is a leaf utility — it implements no port and
// depends only on the standard library — so importing it does not couple two adapters together.
package httpctx

import "context"

// requestIDKey is the unexported context key for the request correlation ID.
type requestIDKey struct{}

// WithRequestID returns a context carrying the request correlation ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext returns the request correlation ID, or the empty string when unset.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
