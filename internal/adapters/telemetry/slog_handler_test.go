package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubHandler is a minimal slog.Handler used in tests.
type stubHandler struct {
	enabled      bool
	handleErr    error
	handleCalled bool
	attrs        []slog.Attr
	group        string
}

func (s *stubHandler) Enabled(_ context.Context, _ slog.Level) bool { return s.enabled }
func (s *stubHandler) Handle(_ context.Context, _ slog.Record) error {
	s.handleCalled = true
	return s.handleErr
}
func (s *stubHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &stubHandler{enabled: s.enabled, attrs: attrs}
}
func (s *stubHandler) WithGroup(name string) slog.Handler {
	return &stubHandler{enabled: s.enabled, group: name}
}

func TestMultiHandler_Handle_FanOutAllHandlers(t *testing.T) {
	t.Run("calls all enabled handlers even when first returns an error", func(t *testing.T) {
		firstErr := errors.New("first handler error")
		first := &stubHandler{enabled: true, handleErr: firstErr}
		second := &stubHandler{enabled: true}

		h := NewMultiHandler(first, second)
		err := h.Handle(context.Background(), slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0))

		require.Error(t, err)
		assert.True(t, first.handleCalled, "first handler must be called")
		assert.True(t, second.handleCalled, "second handler must be called despite first returning an error")
		assert.ErrorIs(t, err, firstErr)
	})

	t.Run("joins multiple errors from multiple failing handlers", func(t *testing.T) {
		firstErr := errors.New("first handler error")
		secondErr := errors.New("second handler error")
		first := &stubHandler{enabled: true, handleErr: firstErr}
		second := &stubHandler{enabled: true, handleErr: secondErr}

		h := NewMultiHandler(first, second)
		err := h.Handle(context.Background(), slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0))

		require.Error(t, err)
		assert.ErrorIs(t, err, firstErr)
		assert.ErrorIs(t, err, secondErr)
	})

	t.Run("returns nil when all handlers succeed", func(t *testing.T) {
		first := &stubHandler{enabled: true}
		second := &stubHandler{enabled: true}

		h := NewMultiHandler(first, second)
		err := h.Handle(context.Background(), slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0))

		require.NoError(t, err)
		assert.True(t, first.handleCalled)
		assert.True(t, second.handleCalled)
	})

	t.Run("skips disabled handlers", func(t *testing.T) {
		enabled := &stubHandler{enabled: true}
		disabled := &stubHandler{enabled: false}

		h := NewMultiHandler(disabled, enabled)
		err := h.Handle(context.Background(), slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0))

		require.NoError(t, err)
		assert.False(t, disabled.handleCalled, "disabled handler must not be called")
		assert.True(t, enabled.handleCalled)
	})
}

func TestMultiHandler_Enabled(t *testing.T) {
	t.Run("returns true if any handler is enabled", func(t *testing.T) {
		h := NewMultiHandler(&stubHandler{enabled: false}, &stubHandler{enabled: true})
		assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))
	})

	t.Run("returns false when no handlers are enabled", func(t *testing.T) {
		h := NewMultiHandler(&stubHandler{enabled: false}, &stubHandler{enabled: false})
		assert.False(t, h.Enabled(context.Background(), slog.LevelInfo))
	})
}

func TestMultiHandler_WithAttrs(t *testing.T) {
	t.Run("propagates attrs to all handlers", func(t *testing.T) {
		first := &stubHandler{enabled: true}
		second := &stubHandler{enabled: true}
		h := NewMultiHandler(first, second)

		attrs := []slog.Attr{slog.String("key", "value")}
		derived := h.WithAttrs(attrs)

		err := derived.Handle(context.Background(), slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0))
		require.NoError(t, err)
	})
}

func TestMultiHandler_WithGroup(t *testing.T) {
	t.Run("propagates group to all handlers", func(t *testing.T) {
		first := &stubHandler{enabled: true}
		second := &stubHandler{enabled: true}
		h := NewMultiHandler(first, second)

		derived := h.WithGroup("mygroup")
		err := derived.Handle(context.Background(), slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0))
		require.NoError(t, err)
	})
}
