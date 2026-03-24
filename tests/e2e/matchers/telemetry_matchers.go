// Package matchers provides custom Gomega matchers for telemetry-specific assertions.
package matchers

import (
	"fmt"

	"github.com/onsi/gomega/types"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ContainSpanWithName asserts that a []sdktrace.ReadOnlySpan contains a span with the given name.
func ContainSpanWithName(name string) types.GomegaMatcher {
	return &spanWithNameMatcher{name: name}
}

type spanWithNameMatcher struct {
	name string
}

func (m *spanWithNameMatcher) Match(actual interface{}) (bool, error) {
	spans, ok := actual.([]sdktrace.ReadOnlySpan)
	if !ok {
		return false, fmt.Errorf("ContainSpanWithName expects []sdktrace.ReadOnlySpan, got %T", actual)
	}
	for _, span := range spans {
		if span.Name() == m.name {
			return true, nil
		}
	}
	return false, nil
}

func (m *spanWithNameMatcher) FailureMessage(actual interface{}) string {
	spans, _ := actual.([]sdktrace.ReadOnlySpan)
	names := spanNames(spans)
	return fmt.Sprintf("Expected spans to contain a span named %q\nActual span names: %v", m.name, names)
}

func (m *spanWithNameMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected spans NOT to contain a span named %q", m.name)
}

// ContainSpanWithAttribute asserts that a []sdktrace.ReadOnlySpan contains a span with
// the given name that has the specified attribute key and value.
func ContainSpanWithAttribute(spanName string, key attribute.Key, value string) types.GomegaMatcher {
	return &spanWithAttributeMatcher{spanName: spanName, key: key, value: value}
}

type spanWithAttributeMatcher struct {
	spanName string
	key      attribute.Key
	value    string
}

func (m *spanWithAttributeMatcher) Match(actual interface{}) (bool, error) {
	spans, ok := actual.([]sdktrace.ReadOnlySpan)
	if !ok {
		return false, fmt.Errorf("ContainSpanWithAttribute expects []sdktrace.ReadOnlySpan, got %T", actual)
	}
	for _, span := range spans {
		if span.Name() != m.spanName {
			continue
		}
		for _, kv := range span.Attributes() {
			if kv.Key == m.key && kv.Value.AsString() == m.value {
				return true, nil
			}
		}
	}
	return false, nil
}

func (m *spanWithAttributeMatcher) FailureMessage(actual interface{}) string {
	spans, _ := actual.([]sdktrace.ReadOnlySpan)
	return fmt.Sprintf(
		"Expected span %q to have attribute %q=%q\nActual span names: %v",
		m.spanName, m.key, m.value, spanNames(spans),
	)
}

func (m *spanWithAttributeMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected span %q NOT to have attribute %q=%q",
		m.spanName, m.key, m.value,
	)
}

// HaveHTTPSpan asserts that a []sdktrace.ReadOnlySpan contains an HTTP span for the given
// method, route, and status code.
// The span name format matches otelchi: "METHOD /route/pattern".
// The http.status_code attribute (semconv v1.20.0, as used by otelchi v0.12.2) is checked against statusCode.
func HaveHTTPSpan(method, route string, statusCode int) types.GomegaMatcher {
	return &httpSpanMatcher{method: method, route: route, statusCode: statusCode}
}

type httpSpanMatcher struct {
	method     string
	route      string
	statusCode int
}

func (m *httpSpanMatcher) Match(actual interface{}) (bool, error) {
	spans, ok := actual.([]sdktrace.ReadOnlySpan)
	if !ok {
		return false, fmt.Errorf("HaveHTTPSpan expects []sdktrace.ReadOnlySpan, got %T", actual)
	}

	expectedName := m.method + " " + m.route

	// otelchi v0.12.2 uses semconv/v1.20.0 which sets "http.status_code" (not "http.response.status_code").
	statusCodeKey := attribute.Key("http.status_code")
	for _, span := range spans {
		if span.Name() != expectedName {
			continue
		}
		for _, kv := range span.Attributes() {
			if kv.Key == statusCodeKey && kv.Value.AsInt64() == int64(m.statusCode) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (m *httpSpanMatcher) FailureMessage(actual interface{}) string {
	spans, _ := actual.([]sdktrace.ReadOnlySpan)
	return fmt.Sprintf(
		"Expected spans to contain HTTP span %q %q with status %d\nActual span names: %v",
		m.method, m.route, m.statusCode, spanNames(spans),
	)
}

func (m *httpSpanMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected spans NOT to contain HTTP span %q %q with status %d",
		m.method, m.route, m.statusCode,
	)
}

// spanNames returns the names of all spans for diagnostic output.
func spanNames(spans []sdktrace.ReadOnlySpan) []string {
	names := make([]string, len(spans))
	for i, s := range spans {
		names[i] = s.Name()
	}
	return names
}
