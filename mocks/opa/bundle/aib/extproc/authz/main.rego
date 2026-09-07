package aib.extproc.authz

import rego.v1

# Allow MCP lifecycle methods (safe, no tool execution)
allow contains {"reason": "safe MCP lifecycle method"} if {
	input.type == "mcp_method"
	input.mcp.method in {"initialize", "ping", "notifications/initialized"}
}

# Allow read-only MCP tool calls
allow contains {"reason": "read-only tool"} if {
	input.type == "mcp_tool_call"
	input.mcp.tool_name in {"list_repositories", "get_file_contents", "search_code", "whoami"}
}

# Deny destructive tool calls
deny contains {"reason": "destructive operations are not permitted"} if {
	input.type == "mcp_tool_call"
	input.mcp.tool_name in {"delete_repository", "force_push"}
}

# Allow all non-MCP traffic (passthrough)
deny contains {"reason": "non-MCP passthrough"} if {
	input.type == "unknown"
}

result := {"action": "deny", "reasons": _deny_reasons} if {
	count(deny) > 0
	_deny_reasons := [entry.reason | some entry in deny]
}

# Allow if no deny and at least one allow
result := {"action": "allow"} if {
	count(deny) == 0
	count(allow) > 0
}