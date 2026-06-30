package aib.extproc.authz

import rego.v1

# Read-only tools allowed; destructive tools denied.
# This policy is used in E2E tests for the OPA authorization feature.

readonly_tools := {"list_repositories", "list_files", "read_file"}

destructive_tools := {"delete_repository", "delete_file", "push_file", "create_issue"}

allow contains {"action": "allow", "reason": "tool is read-only"} if {
	input.type == "mcp_tool_call"
	input.mcp.tool_name in readonly_tools
}

deny contains {"action": "deny", "reason": "tool is destructive"} if {
	input.type == "mcp_tool_call"
	input.mcp.tool_name in destructive_tools
}

# Allow non-tool-call MCP methods (e.g. initialize, resources/list)
allow contains {"action": "allow", "reason": "non-tool-call method allowed"} if {
	input.type == "mcp_method"
}

# Allow MCP header-only GET requests (SSE stream setup)
allow contains {"action": "allow", "reason": "mcp header-only request allowed"} if {
	input.type == "mcp_headers_only"
	input.attributes.request.http.method == "GET"
}

result := decision if {
	count(deny) > 0
	decision := {
		"action": "deny",
		"reasons": [r | some d in deny; r := d.reason],
	}
} else := decision if {
	count(allow) > 0
	decision := {"action": "allow"}
} else := {"action": "deny", "reasons": ["default deny: no matching allow rule"]}
