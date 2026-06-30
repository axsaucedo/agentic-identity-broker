package aib.extproc.authz

import rego.v1

# Read-only tools allowed; destructive tools denied.
# This policy is used in E2E tests for the OPA authorization feature.

readonly_tools := {"list_repositories", "list_files", "read_file"}

destructive_tools := {"delete_repository", "delete_file", "push_file", "create_issue"}

required_permission_tool := "permissioned_read"
required_service_id := "22222222-2222-2222-2222-222222222222"

permission_set_context_available if {
	input.context.granted_permission_sets_available
}

permission_set_grants_service(service_id) if {
	permission_set_context_available
	some permission_set_id
	service_id in input.context.granted_permission_sets[permission_set_id]
}

allow contains {"action": "allow", "reason": "tool is read-only"} if {
	input.type == "mcp_tool_call"
	input.mcp.tool_name in readonly_tools
}

deny contains {"action": "deny", "reason": "tool is destructive"} if {
	input.type == "mcp_tool_call"
	input.mcp.tool_name in destructive_tools
}

deny contains {"action": "deny", "reason": "permission set context unavailable"} if {
	input.type == "mcp_tool_call"
	input.mcp.tool_name == required_permission_tool
	not permission_set_context_available
}

allow contains {"action": "allow", "reason": "permission set grants service access"} if {
	input.type == "mcp_tool_call"
	input.mcp.tool_name == required_permission_tool
	permission_set_grants_service(required_service_id)
}

deny contains {"action": "deny", "reason": "required permission set missing"} if {
	input.type == "mcp_tool_call"
	input.mcp.tool_name == required_permission_tool
	permission_set_context_available
	not permission_set_grants_service(required_service_id)
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
