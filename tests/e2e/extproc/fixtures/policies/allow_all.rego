package aib.extproc.authz

import rego.v1

# Allows all requests unconditionally.
# Used in E2E tests to verify the allow path and disabled-by-default OPA behaviour.

allow contains {"action": "allow", "reason": "policy: all requests allowed"} if {
	true
}

result := decision if {
	count(allow) > 0
	decision := {"action": "allow"}
} else := {"action": "deny", "reasons": ["default deny"]}
