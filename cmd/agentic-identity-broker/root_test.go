package main

import "testing"

func TestRootCommandRegistersRequestContextFlags(t *testing.T) {
	for _, name := range []string{
		"request_context.trusted_proxy.enabled",
		"request_context.trusted_proxy.forwarded_header",
		"request_context.trace.response_enabled",
	} {
		if rootCmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("missing request-context CLI flag %q", name)
		}
	}
}
