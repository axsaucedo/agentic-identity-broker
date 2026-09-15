package authorization

// OPADecision represents the structured result of OPA policy evaluation.
type OPADecision struct {
	Action  string   `json:"action"`            // "allow", "deny", "approval_required", "ciba_required"
	Reasons []string `json:"reasons,omitempty"` // Human-readable denial reasons
}

// ParseDecision extracts an OPADecision from a raw OPA evaluation result map.
// The result is expected to be a map[string]any produced by a Rego query where
// the decision rule evaluates to an object with "action" and optional "reasons" fields.
//
// Behaviour:
//   - nil or empty map → deny (undefined / missing result)
//   - action "approval_required" or "ciba_required" → deny (not yet supported)
//   - action "allow" → allow with no reasons
//   - action "deny" → deny with reasons extracted from the "reasons" field
//   - any other action value → deny
func ParseDecision(result map[string]any) *OPADecision {
	if result == nil {
		return &OPADecision{Action: "deny", Reasons: []string{"undefined result"}}
	}

	actionRaw, ok := result["action"]
	if !ok {
		return &OPADecision{Action: "deny", Reasons: []string{"undefined result"}}
	}

	action, ok := actionRaw.(string)
	if !ok {
		return &OPADecision{Action: "deny", Reasons: []string{"invalid action type"}}
	}

	switch action {
	case "allow":
		return &OPADecision{Action: "allow"}

	case "deny":
		var reasons []string
		if reasonsRaw, exists := result["reasons"]; exists {
			if reasonsList, ok := reasonsRaw.([]any); ok {
				for _, r := range reasonsList {
					if s, ok := r.(string); ok {
						reasons = append(reasons, s)
					}
				}
			}
		}
		return &OPADecision{Action: "deny", Reasons: reasons}

	case "approval_required", "ciba_required":
		return &OPADecision{
			Action:  "deny",
			Reasons: []string{action + " is not yet supported"},
		}

	default:
		return &OPADecision{
			Action:  "deny",
			Reasons: []string{"unknown action: " + action},
		}
	}
}
