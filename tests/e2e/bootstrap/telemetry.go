// Package bootstrap provides test infrastructure for E2E testing.
package bootstrap

import (
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// NewInMemoryTracerProvider creates an in-memory TracerProvider for E2E testing.
// Returns both the provider and the SpanRecorder for asserting on emitted spans.
// The provider uses AlwaysSample to capture all spans during tests.
func NewInMemoryTracerProvider() (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(recorder),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	return tp, recorder
}
