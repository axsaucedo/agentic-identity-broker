package aib.extproc.authz

import rego.v1

# Denies all requests unconditionally.
# Used in E2E tests to verify deny-by-default behaviour.

deny contains {"action": "deny", "reason": "policy: all tools denied"} if {
	true
}

result := decision if {
	count(deny) > 0
	decision := {
		"action": "deny",
		"reasons": [r | some d in deny; r := d.reason],
	}
} else := {"action": "deny", "reasons": ["default deny"]}
