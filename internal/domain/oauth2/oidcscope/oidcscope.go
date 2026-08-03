package oidcscope

// ReservedRefreshTokenScopes are AS-level scopes that request refresh-token issuance.
// They are permitted regardless of an agent's allowed scopes.
var ReservedRefreshTokenScopes = []string{"offline", "offline_access"}

// IsReservedRefreshTokenScope reports whether s is an AS-level refresh-token scope.
func IsReservedRefreshTokenScope(s string) bool {
	for _, scope := range ReservedRefreshTokenScopes {
		if scope == s {
			return true
		}
	}
	return false
}
