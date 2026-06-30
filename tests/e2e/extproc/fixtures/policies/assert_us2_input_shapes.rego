package aib.extproc.authz

import rego.v1

allow contains {"action": "allow"} if {
  input.type == "mcp_tool_call"
  input.mcp.tool_name == "create_issue"
  input.mcp.arguments.title == "Bug"
  input.mcp.arguments.repo == "acme/app"
}

allow contains {"action": "allow"} if {
  input.type == "mcp_method"
  input.mcp.method == "initialize"
  input.mcp.params.protocolVersion == "2025-03-26"
  input.mcp.params.clientInfo.name == "test-agent"
  input.mcp.params.clientInfo.version == "1.0"
}

allow contains {"action": "allow"} if {
  input.type == "unknown"
  input.attributes.request.http.body == "{\"action\":\"run\",\"tool\":\"bash\",\"command\":\"ls -la\"}"
}

result := decision if {
  count(allow) > 0
  decision := {"action": "allow"}
} else := {"action": "deny", "reasons": ["unexpected OPA input shape"]}
