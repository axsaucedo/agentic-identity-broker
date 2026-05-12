package oauth2server

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
	"github.com/ory/fosite"
)

// TokenClaimsEvaluator compiles and evaluates CEL expressions for custom JWT claims.
type TokenClaimsEvaluator struct {
	program cel.Program
}

// NewTokenClaimsEvaluator compiles a CEL expression at startup.
// Returns error if the expression is invalid (fail-fast per FR-013b).
// Returns nil evaluator if expression is empty (no custom claims).
func NewTokenClaimsEvaluator(expression string) (*TokenClaimsEvaluator, error) {
	if expression == "" {
		return nil, nil
	}

	env, err := cel.NewEnv(
		cel.Variable("agent", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("principal", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("request", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("scope", cel.ListType(cel.StringType)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("invalid token_claims_expression: %w", issues.Err())
	}

	// Type verification: we accept any map output type.
	// The actual return value is validated at runtime in Evaluate().
	outTypeStr := ast.OutputType().String()
	if !strings.HasPrefix(outTypeStr, "map") && outTypeStr != "dyn" {
		return nil, fmt.Errorf("token_claims_expression must return a map, got %s", outTypeStr)
	}

	program, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL program: %w", err)
	}

	return &TokenClaimsEvaluator{program: program}, nil
}

// Evaluate runs the CEL expression with the given fosite requester context.
// Returns a map of custom claims to add to the JWT.
// Returns error if evaluation fails (fail-closed per FR-013e).
func (e *TokenClaimsEvaluator) Evaluate(_ context.Context, requester fosite.Requester) (map[string]interface{}, error) {
	if e == nil || e.program == nil {
		return nil, nil
	}

	// Build agent context: id=agent UUID, client_id=upstream OAuth2 client ID,
	// display_name from the agent entity if available.
	agentCtx := map[string]interface{}{
		"id":           requester.GetClient().GetID(),
		"client_id":    "",
		"display_name": "",
	}
	if h, ok := requester.GetClient().(agentHolder); ok {
		if agent := h.getAgent(); agent != nil {
			if agent.ClientID != nil {
				agentCtx["client_id"] = agent.ClientID.String()
			}
			agentCtx["display_name"] = agent.DisplayName
		}
	}

	// Build principal context: id=subject (principal string).
	// email and display_name are not available from the session alone and are left empty;
	// they may be populated in a future extension when profile attributes are persisted.
	subject := requester.GetSession().GetSubject()
	principalCtx := map[string]interface{}{
		"id":           subject,
		"email":        "",
		"display_name": "",
	}

	// Build request context: grant_type from the access request (if available), scopes from the requester.
	grantType := ""
	if ar, ok := requester.(*fosite.AccessRequest); ok && len(ar.GetGrantTypes()) > 0 {
		grantType = ar.GetGrantTypes()[0]
	}
	requestCtx := map[string]interface{}{
		"grant_type": grantType,
		"scopes":     requester.GetGrantedScopes(),
	}

	activation := map[string]interface{}{
		"agent":     agentCtx,
		"principal": principalCtx,
		"request":   requestCtx,
		"scope":     requester.GetGrantedScopes(),
	}

	out, _, err := e.program.Eval(activation)
	if err != nil {
		return nil, fmt.Errorf("CEL evaluation failed: %w", err)
	}

	return convertCELOutput(out)
}

func convertCELOutput(out ref.Val) (map[string]interface{}, error) {
	if out == nil || out.Value() == nil {
		return nil, nil
	}

	// Try direct conversion from CEL value
	if mapVal, ok := out.Value().(map[string]interface{}); ok {
		return mapVal, nil
	}
	if mapVal, ok := out.Value().(map[ref.Val]ref.Val); ok {
		result := make(map[string]interface{}, len(mapVal))
		for k, v := range mapVal {
			result[fmt.Sprint(k.Value())] = v.Value()
		}
		return result, nil
	}

	return nil, fmt.Errorf("token_claims_expression must return map, got %T", out.Value())
}
