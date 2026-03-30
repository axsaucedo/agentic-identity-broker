// Package oauth2server implements the broker's local OAuth2 token minting mode.
// It uses ory/fosite as a headless domain library — only the handler layer is used,
// not the HTTP orchestration layer. Fosite types are contained within this package
// and never leak into ports, adapters/http, or app.
package oauth2server
