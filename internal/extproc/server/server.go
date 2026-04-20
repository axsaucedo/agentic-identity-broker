// Package server implements the Envoy ExtProc gRPC server for token exchange.
// The Server type handles the streaming Process RPC, extracts Bearer tokens,
// performs RFC 8693 token exchange, and replaces the Authorization header.
package server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
)

// Exchanger performs RFC 8693 token exchange with in-memory caching.
// Implementations must be safe for concurrent use.
type Exchanger interface {
	// Exchange exchanges subjectToken for a downstream token scoped to resourceURI.
	// Returns the exchanged access token or an error.
	Exchange(subjectToken, resourceURI string) (string, error)
	// Shutdown releases any resources held by the exchanger (e.g., background goroutines).
	Shutdown()
}

// Server implements the Envoy ExternalProcessorServer gRPC interface.
// It intercepts request headers, performs token exchange, and replaces
// the Authorization header before the request reaches the upstream.
type Server struct {
	extprocv3.UnimplementedExternalProcessorServer
	cfg       *extprocconfig.Config
	exchanger Exchanger
	logger    *slog.Logger
}

// NewServer creates a new ExtProc Server.
// cfg provides the service configuration; exchanger performs token exchange;
// logger is used for structured logging.
func NewServer(cfg *extprocconfig.Config, exchanger Exchanger, logger *slog.Logger) *Server {
	return &Server{
		cfg:       cfg,
		exchanger: exchanger,
		logger:    logger,
	}
}

// Process implements the streaming ExtProc gRPC RPC.
// It processes each message from Envoy in sequence:
//   - RequestHeaders: extracts Bearer token and :path, performs token exchange
//   - All other phases: pass through unchanged
func (s *Server) Process(stream extprocv3.ExternalProcessor_ProcessServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			if status.Code(err) == codes.Canceled {
				return nil
			}
			s.logger.Debug("stream recv error", "error", err)
			return err
		}

		var resp *extprocv3.ProcessingResponse

		switch msg := req.Request.(type) {
		case *extprocv3.ProcessingRequest_RequestHeaders:
			resp = s.processRequestHeaders(msg.RequestHeaders)

		case *extprocv3.ProcessingRequest_RequestBody:
			// FR-009 equivalent: echo body bytes back unchanged.
			// An empty BodyResponse{} clears the body in agentgateway; we must
			// explicitly set the BodyMutation to return the original bytes.
			resp = echoRequestBody(msg.RequestBody)

		case *extprocv3.ProcessingRequest_ResponseHeaders:
			resp = passThroughResponseHeaders()

		case *extprocv3.ProcessingRequest_ResponseBody:
			// Same pattern: echo response body back unchanged.
			resp = echoResponseBody(msg.ResponseBody)

		case *extprocv3.ProcessingRequest_RequestTrailers:
			resp = passThroughRequestTrailers()

		case *extprocv3.ProcessingRequest_ResponseTrailers:
			resp = passThroughResponseTrailers()

		default:
			// Unknown phase — pass through unchanged
			resp = passThrough()
		}

		if err := stream.Send(resp); err != nil {
			s.logger.Info("stream send error", "error", err)
			return err
		}

		// After an ImmediateResponse the stream is complete.
		if _, isImmediate := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse); isImmediate {
			return nil
		}
	}
}

// processRequestHeaders implements the primary ExtProc processing phase.
// It handles:
//   - No Bearer token → pass through (FR-009)
//   - Empty/invalid :path → 503 ImmediateResponse (FR-013)
//   - Successful exchange → replace Authorization header (FR-007)
//   - Re-auth required (broker error_uri) → URLElicitationRequiredError ImmediateResponse (HTTP 200, code -32042)
//   - Exchange failure → 500 ImmediateResponse (FR-008, FR-010)
func (s *Server) processRequestHeaders(headers *extprocv3.HttpHeaders) *extprocv3.ProcessingResponse {
	bearerToken := extractBearerToken(headers)
	if bearerToken == "" {
		// FR-009: No Bearer token — pass through without modification.
		s.logger.Debug("no Bearer token — passing through")
		return passThrough()
	}

	path := extractHeader(headers, ":path")
	scheme := extractHeader(headers, ":scheme")
	authority := extractHeader(headers, ":authority")

	// FR-013: Build an absolute resource URI for the RFC 8693 token exchange.
	// agentgateway sends a relative :path (e.g., /mcp) with :scheme and :authority as separate
	// pseudo-headers. Construct the absolute URI: {scheme}://{authority}{path}.
	// Falls back to :path directly if it is already an absolute URI.
	resourceURI := buildResourceURI(scheme, authority, path)

	if err := validateResourceURI(resourceURI); err != nil {
		s.logger.Warn("extproc: invalid resource URI — rejecting with 503",
			"path", path,
			"scheme", scheme,
			"authority", authority,
			"resource_uri", resourceURI,
			"error", err)
		return immediateResponse(httpv3.StatusCode_ServiceUnavailable,
			`{"error":"invalid_resource","error_description":"request URI is empty or invalid"}`)
	}

	exchangedToken, err := s.exchanger.Exchange(bearerToken, resourceURI)
	if err != nil {
		if errors.Is(err, ErrAssertionExpired) {
			s.logger.Error("token exchange failed: client assertion expired — background refresh may have failed",
				"resource", resourceURI)
			return immediateResponse(httpv3.StatusCode_ServiceUnavailable,
				`{"error":"service_unavailable","error_description":"client assertion expired"}`)
		}
		var brokerErr *BrokerExchangeError
		if errors.As(err, &brokerErr) && brokerErr.ErrorURI != "" {
			// Re-authentication required: return URLElicitationRequiredError immediately from
			// the headers phase. id is null because the request body has not been read yet;
			// this is correct per JSON-RPC 2.0 §5 ("if the id cannot be determined, use null").
			// Note: agentgateway 0.12.0 commits the request to the backend as soon as a
			// RequestHeaders response is received, so ImmediateResponse must be returned here
			// (not from a body phase) to prevent the request from being forwarded.
			s.logger.Info("token exchange requires re-authentication — returning URLElicitationRequiredError",
				"resource", resourceURI,
				"code", brokerErr.Code,
				"error_uri", brokerErr.ErrorURI)
			return s.urlElicitationResponse(brokerErr)
		}
		if errors.Is(err, ErrCircuitOpen) {
			s.logger.Warn("token exchange rejected: circuit breaker is open",
				"resource", resourceURI)
			return immediateResponse(httpv3.StatusCode_ServiceUnavailable,
				`{"error":"service_unavailable","error_description":"circuit breaker is open"}`)
		}
		s.logger.Error("token exchange failed", "resource", resourceURI, "error", err)
		return immediateResponse(httpv3.StatusCode_InternalServerError,
			`{"error":"token_exchange_failed","error_description":"token exchange request failed"}`)
	}
	s.logger.Debug("token exchanged successfully", "resource", resourceURI)
	return replaceAuthorizationHeader("Bearer " + exchangedToken)
}

// extractBearerToken extracts the raw token value from a "Bearer <token>" Authorization header.
// Returns empty string if the header is absent or uses a different scheme.
func extractBearerToken(headers *extprocv3.HttpHeaders) string {
	authValue := extractHeader(headers, "authorization")
	if authValue == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authValue, prefix) {
		return ""
	}
	return strings.TrimPrefix(authValue, prefix)
}

// extractHeader returns the value of the named header (case-insensitive) from the header map.
// RawValue (bytes) takes precedence over Value (string). agentgateway uses Value for
// pseudo-headers (:path, :method, :scheme, :authority) and RawValue for regular headers.
// Returns empty string if not found.
func extractHeader(headers *extprocv3.HttpHeaders, name string) string {
	if headers == nil || headers.Headers == nil {
		return ""
	}
	nameLower := strings.ToLower(name)
	for _, h := range headers.Headers.Headers {
		if strings.ToLower(h.Key) == nameLower {
			if len(h.RawValue) > 0 {
				return string(h.RawValue)
			}
			return h.Value
		}
	}
	return ""
}

// buildResourceURI constructs an absolute URI for use as the RFC 8693 resource parameter.
// If path is already an absolute URI (starts with http:// or https://), it is returned as-is.
// Otherwise, the URI is assembled from the HTTP/2 pseudo-headers: {scheme}://{authority}{path}.
// Returns an empty string if both path and authority are empty.
func buildResourceURI(scheme, authority, path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if authority == "" {
		return path // fall through to validation which will reject empty/relative paths
	}
	if scheme == "" {
		scheme = "https" // default to https if scheme is missing, per Envoy's behavior
	}
	return scheme + "://" + authority + path
}

// validateResourceURI checks that resourceURI is a non-empty absolute URI with
// http or https scheme, mitigating SSRF attacks per FR-004, FR-013.
func validateResourceURI(resourceURI string) error {
	if strings.TrimSpace(resourceURI) == "" {
		return status.Error(codes.InvalidArgument, "empty :path")
	}
	u, err := url.ParseRequestURI(resourceURI)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid :path: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return status.Errorf(codes.InvalidArgument, ":path must have http or https scheme, got %q", u.Scheme)
	}
	if u.Host == "" {
		return status.Error(codes.InvalidArgument, ":path must have a non-empty host")
	}
	return nil
}

// replaceAuthorizationHeader builds a ProcessingResponse that replaces the
// Authorization header with the provided value.
func replaceAuthorizationHeader(value string) *extprocv3.ProcessingResponse {
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extprocv3.HeadersResponse{
				Response: &extprocv3.CommonResponse{
					HeaderMutation: &extprocv3.HeaderMutation{
						SetHeaders: []*corev3.HeaderValueOption{
							{
								Header: &corev3.HeaderValue{
									Key:      "authorization",
									RawValue: []byte(value),
								},
							},
						},
					},
				},
			},
		},
	}
}

// immediateResponse builds a ProcessingResponse_ImmediateResponse with the given
// HTTP status code and JSON body. Used for error responses (500, 503).
func immediateResponse(code httpv3.StatusCode, body string) *extprocv3.ProcessingResponse {
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ImmediateResponse{
			ImmediateResponse: &extprocv3.ImmediateResponse{
				Status: &httpv3.HttpStatus{Code: code},
				Headers: &extprocv3.HeaderMutation{
					SetHeaders: []*corev3.HeaderValueOption{
						{
							Header: &corev3.HeaderValue{
								Key:      "content-type",
								RawValue: []byte("application/json"),
							},
						},
					},
				},
				Body: []byte(body),
			},
		},
	}
}

// urlElicitationResponse builds a ProcessingResponse_ImmediateResponse with HTTP 200 and
// a JSON-RPC 2.0 URLElicitationRequiredError body (error code -32042).
// Per MCP spec 2025-11-05: returned when a request cannot proceed until the user visits
// a URL for OAuth re-authentication.
// HTTP 200 is used because JSON-RPC errors always travel over HTTP 200.
// id is always null: the response is returned from the headers phase before the body is read,
// which is correct per JSON-RPC 2.0 §5 ("if the id cannot be determined, use null").
func (s *Server) urlElicitationResponse(brokerErr *BrokerExchangeError) *extprocv3.ProcessingResponse {
	elicitErr := mcp.URLElicitationRequiredError{
		Elicitations: []mcp.ElicitationParams{
			{
				Mode:          mcp.ElicitationModeURL,
				ElicitationID: uuid.New().String(),
				URL:           brokerErr.ErrorURI,
				Message:       brokerErr.Description,
			},
		},
	}
	jsonRPCErr := elicitErr.JSONRPCError()
	// Override the generated message with the broker-supplied description so that
	// the client receives context about why re-authentication is required.
	jsonRPCErr.Error.Message = brokerErr.Description

	body, err := json.Marshal(jsonRPCErr)
	if err != nil {
		s.logger.Error("failed to marshal URLElicitationRequiredError", "error", err)
		return immediateResponse(httpv3.StatusCode_InternalServerError,
			`{"error":"internal_error","error_description":"failed to marshal error response"}`)
	}
	return immediateResponse(httpv3.StatusCode_OK, string(body))
}

// passThrough builds a ProcessingResponse_RequestHeaders with no mutations,
// instructing Envoy to pass the request through unchanged.
func passThrough() *extprocv3.ProcessingResponse {
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extprocv3.HeadersResponse{},
		},
	}
}

// echoRequestBody echoes the request body bytes back unchanged using StreamedBodyResponse.
// StreamedBodyResponse is required by agentgateway; BodyMutation_Body causes body loss.
func echoRequestBody(body *extprocv3.HttpBody) *extprocv3.ProcessingResponse {
	streamed := &extprocv3.StreamedBodyResponse{}
	if body != nil {
		streamed.Body = body.Body
		streamed.EndOfStream = body.EndOfStream
	}
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestBody{
			RequestBody: &extprocv3.BodyResponse{
				Response: &extprocv3.CommonResponse{
					BodyMutation: &extprocv3.BodyMutation{
						Mutation: &extprocv3.BodyMutation_StreamedResponse{
							StreamedResponse: streamed,
						},
					},
				},
			},
		},
	}
}

// passThroughResponseHeaders builds a phase-specific pass-through response for response headers.
func passThroughResponseHeaders() *extprocv3.ProcessingResponse {
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ResponseHeaders{
			ResponseHeaders: &extprocv3.HeadersResponse{},
		},
	}
}

// echoResponseBody echoes the response body bytes back unchanged using StreamedBodyResponse.
// StreamedBodyResponse is required by agentgateway; BodyMutation_Body causes body loss.
func echoResponseBody(body *extprocv3.HttpBody) *extprocv3.ProcessingResponse {
	streamed := &extprocv3.StreamedBodyResponse{}
	if body != nil {
		streamed.Body = body.Body
		streamed.EndOfStream = body.EndOfStream
	}
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ResponseBody{
			ResponseBody: &extprocv3.BodyResponse{
				Response: &extprocv3.CommonResponse{
					BodyMutation: &extprocv3.BodyMutation{
						Mutation: &extprocv3.BodyMutation_StreamedResponse{
							StreamedResponse: streamed,
						},
					},
				},
			},
		},
	}
}

// passThroughRequestTrailers builds a phase-specific pass-through response for request trailers.
func passThroughRequestTrailers() *extprocv3.ProcessingResponse {
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestTrailers{
			RequestTrailers: &extprocv3.TrailersResponse{},
		},
	}
}

// passThroughResponseTrailers builds a phase-specific pass-through response for response trailers.
func passThroughResponseTrailers() *extprocv3.ProcessingResponse {
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ResponseTrailers{
			ResponseTrailers: &extprocv3.TrailersResponse{},
		},
	}
}
