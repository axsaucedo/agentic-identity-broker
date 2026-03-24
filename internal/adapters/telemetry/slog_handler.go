package telemetry

import (
	"context"
	"errors"
	"log/slog"
)

// multiHandler fans out log records to multiple slog.Handler implementations.
type multiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a slog.Handler that writes to all provided handlers.
// Each record is cloned before being passed to subsequent handlers to prevent
// mutation of shared state across handlers.
func NewMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

// Enabled returns true if any of the underlying handlers is enabled for the given level.
func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle fans the record out to all handlers that are enabled for the record's level.
// Records are cloned before each handler call to prevent cross-handler mutation.
func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r.Clone()); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

// WithAttrs returns a new multiHandler with the given attributes added to all underlying handlers.
func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

// WithGroup returns a new multiHandler with the given group name applied to all underlying handlers.
func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}
