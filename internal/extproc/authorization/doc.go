// Package authorization provides OPA-based authorization for the ExtProc service.
// For body-bearing requests, ExtProc exchanges tokens in RequestHeaders and evaluates
// the buffered request body against configurable Rego policies in RequestBody.
// For header-only requests, OPA evaluates first and token exchange runs only after allow,
// implementing Tier 1 of the three-tier authorization model.
package authorization
