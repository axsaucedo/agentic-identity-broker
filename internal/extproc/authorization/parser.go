package authorization

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// MCPMessage represents a parsed JSON-RPC 2.0 message from the MCP protocol.
// Params is kept as a raw map to support arbitrary method parameter schemas.
type MCPMessage struct {
	JSONRPC string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	ID      any            `json:"id"`
	Params  map[string]any `json:"params"`
}

// ParseMCPMessage parses a single JSON-RPC 2.0 message from body bytes.
// Returns an error for empty body, malformed JSON, non-object JSON values,
// missing/invalid jsonrpc version, or empty method.
// Batch messages (JSON arrays) are handled by ParseMCPBatch.
func ParseMCPMessage(body []byte) (*MCPMessage, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("mcp parser: empty body")
	}

	var msg MCPMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("mcp parser: invalid JSON-RPC: %w", err)
	}

	// Validate required JSON-RPC 2.0 fields to prevent misclassification of
	// arbitrary JSON as MCP messages (e.g. when protocol="mcp" is set by the proxy
	// but the body is not valid JSON-RPC 2.0).
	if msg.JSONRPC != "2.0" {
		return nil, fmt.Errorf("mcp parser: missing or invalid jsonrpc field (expected \"2.0\", got %q)", msg.JSONRPC)
	}
	if msg.Method == "" {
		return nil, fmt.Errorf("mcp parser: missing or empty method field")
	}

	return &msg, nil
}

// ParseMCPBatch detects and parses JSON-RPC 2.0 batch requests (FR-023).
// If body begins with '[', it is parsed as a batch and each element is parsed
// individually. Returns an error if any element is malformed.
// Returns nil, nil if body is not a JSON array (caller should use ParseMCPMessage).
func ParseMCPBatch(body []byte) ([]*MCPMessage, error) {
	trimmed := bytes.TrimLeft(body, " \t\r\n")
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, nil // not a batch
	}

	var rawMessages []json.RawMessage
	if err := json.Unmarshal(body, &rawMessages); err != nil {
		return nil, fmt.Errorf("mcp parser: invalid batch JSON: %w", err)
	}

	messages := make([]*MCPMessage, 0, len(rawMessages))
	for i, raw := range rawMessages {
		msg, err := ParseMCPMessage(raw)
		if err != nil {
			return nil, fmt.Errorf("mcp parser: batch element %d: %w", i, err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
