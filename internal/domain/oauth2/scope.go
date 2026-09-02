package oauth2

import (
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/oidcscope"
)

func SplitScope(scope string) []string {
	if scope == "" {
		return []string{}
	}
	return strings.Split(scope, " ")
}

func IsScopeAllowed(allowedScopes []string, scope string) bool {
	if oidcscope.IsReservedRefreshTokenScope(scope) || len(allowedScopes) == 0 {
		return true
	}
	for _, allowedScope := range allowedScopes {
		if allowedScope == scope {
			return true
		}
	}
	return false
}
