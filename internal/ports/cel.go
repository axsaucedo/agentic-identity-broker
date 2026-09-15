// Package ports defines interfaces for hexagonal architecture boundaries.
package ports

import "context"

// CELCompilerPort defines the interface for compiling and validating CEL expressions.
// This is a hexagonal architecture port - domain logic depends on this interface
// for CEL expression validation and compilation.
//
// Implementation: Uses github.com/google/cel-go for expression parsing and compilation.
type CELCompilerPort interface {
	// CompileExpression compiles a CEL expression string into an executable program.
	// This validates the expression syntax at compile time (startup validation per FR-017).
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - expression: Raw CEL expression string to compile
	//
	// Returns:
	//   - CELProgram: Compiled program ready for evaluation
	//   - error: If expression is invalid (syntax error, type error, etc.)
	//
	// Errors returned:
	//   - ValidationError if expression contains syntax errors
	//   - InvalidExpressionError if expression references undefined variables/functions
	CompileExpression(ctx context.Context, expression string) (CELProgram, error)

	// ValidateExpression validates whether a CEL expression is well-formed without fully compiling it.
	// Lighter-weight than CompileExpression when you only need to check validity.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - expression: Raw CEL expression string to validate
	//
	// Returns:
	//   - error: If expression is invalid
	ValidateExpression(ctx context.Context, expression string) error
}

// CELProgram represents a compiled CEL expression ready for evaluation.
// It is immutable after compilation and can be safely evaluated multiple times
// in parallel without race conditions.
type CELProgram interface {
	// Eval evaluates the compiled CEL expression in the given context.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts (enforces evaluation timeout)
	//   - variables: Map of variable names to values available in the CEL expression
	//                 (e.g., {"claims": claims_object, "request": request_object})
	//
	// Returns:
	//   - interface{}: The result of expression evaluation (typically bool for authorization)
	//   - error: If evaluation fails or times out
	//
	// Errors returned:
	//   - TimeoutError if evaluation exceeds context deadline
	//   - EvaluationError if variable not found or type mismatch
	Eval(ctx context.Context, variables map[string]interface{}) (interface{}, error)
}

// CELAuthorizationContext represents the context provided to CEL authorization expressions.
// This is the input object passed to CEL expressions for gateway authorization decisions.
type CELAuthorizationContext struct {
	// Claims contains the validated client_assertion JWT claims.
	// Available in CEL as: claims.sub, claims.aud, claims.iss, claims.exp, claims.iat, etc.
	Claims map[string]interface{}

	// Request contains token exchange request context.
	// Available in CEL as: request.resource, request.grant_type, request.scope, etc.
	Request CELRequestContext
}

// CELRequestContext represents token exchange request parameters for CEL evaluation.
type CELRequestContext struct {
	// Resource is the target resource URI from the token exchange request.
	// The gateway is requesting a token for this resource.
	Resource string

	// GrantType is the grant_type from the token exchange request.
	// Should be "urn:ietf:params:oauth:grant-type:token-exchange" for token exchange.
	GrantType string

	// Scope is the scope parameter from the token exchange request (optional).
	// May be empty if not provided by gateway.
	Scope string

	// Principal is the user principal extracted from subject_token.
	// This is extracted using ClaimExtractionConfig.PrincipalExpression.
	Principal string

	// AgentID is the agent identifier extracted from subject_token.
	// This is extracted using ClaimExtractionConfig.AgentIDExpression.
	AgentID string
}

// CELClaimExtractionContext represents the context for claim extraction expressions.
type CELClaimExtractionContext struct {
	// SubjectToken contains the validated subject_token JWT claims.
	// Available in CEL as: subject_token.sub, subject_token.azp, etc.
	SubjectToken map[string]interface{}

	// ClientAssertion contains the validated client_assertion JWT claims.
	// May be needed for advanced claim extraction scenarios.
	ClientAssertion map[string]interface{}
}
