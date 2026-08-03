package oidcscope

import "testing"

func TestIsReservedRefreshTokenScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		scope string
		want  bool
	}{
		{name: "offline", scope: "offline", want: true},
		{name: "offline_access", scope: "offline_access", want: true},
		{name: "other scope", scope: "openid", want: false},
		{name: "empty", scope: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsReservedRefreshTokenScope(tt.scope); got != tt.want {
				t.Fatalf("IsReservedRefreshTokenScope(%q) = %v, want %v", tt.scope, got, tt.want)
			}
		})
	}
}
