// Package server implements the Envoy ExtProc gRPC server for token exchange.
// The Server type handles the streaming Process RPC, extracts Bearer tokens,
// performs RFC 8693 token exchange, and replaces the Authorization header.
// In OPA mode, body-bearing requests exchange in the RequestHeaders phase and
// evaluate policy in the RequestBody phase; header-only requests evaluate first.
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocfilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/authorization"
	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
)

// agentgatewayProtocolMetadataKey is the filter_metadata namespace key used by agentgateway
// to pass the protocol type (e.g., "mcp", "a2a") to ExtProc via MetadataContext.
// The metadata structure is: MetadataContext.FilterMetadata["agentgateway"]["protocol"] = "<type>".
// Source: agentgateway ExtProc filter configuration (specs/020-extproc-opa-authorization/research.md).
const agentgatewayProtocolMetadataKey = "agentgateway"

// agentgatewayProtocolFieldKey is the field name within the agentgateway filter metadata struct
// that contains the protocol type value.
const agentgatewayProtocolFieldKey = "protocol"

// ExchangeResult holds the result of a successful token exchange.
type ExchangeResult struct {
	Token                 string
	GrantedPermissionSets map[string][]string
}

// Exchanger performs RFC 8693 token exchange with in-memory caching.
// Implementations must be safe for concurrent use.
type Exchanger interface {
	// Exchange exchanges subjectToken for a downstream token scoped to resourceURI.
	// ctx carries trace context and deadlines and must be passed to outbound HTTP requests.
	// Returns an ExchangeResult containing the token and any granted permission sets, or an error.
	Exchange(ctx context.Context, subjectToken, resourceURI string) (ExchangeResult, error)
	// Shutdown releases any resources held by the exchanger (e.g., background goroutines).
	Shutdown()
}

// Server implements the Envoy ExternalProcessorServer gRPC interface.
// It intercepts request headers, performs token exchange, and replaces
// the Authorization header before the request reaches the upstream.
// When an authorizer is configured, body-bearing requests exchange in the
// headers phase and evaluate OPA in the body phase; header-only requests do the reverse.
type Server struct {
	extprocv3.UnimplementedExternalProcessorServer
	cfg             *extprocconfig.Config
	exchanger       Exchanger
	authorizer      authorization.Authorizer // nil when OPA authorization is disabled
	logger          *slog.Logger
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
}

// requestState holds per-stream state accumulated from the RequestHeaders phase
// that is needed when processing the RequestBody phase (OPA mode only).
type requestState struct {
	bearerToken           string
	resourceURI           string
	headers               map[string]string
	protocol              string
	grantedPermissionSets map[string][]string
	requestContext        context.Context
	finishObservation     func(outcome, resourceURI, errorType string)
}

// NewServer creates a new ExtProc Server without OPA authorization.
// cfg provides the service configuration; exchanger performs token exchange;
// logger is used for structured logging.
// Metric instruments are obtained from the globally registered MeterProvider so
// that tests can inject a ManualReader-backed provider before calling NewServer.
func NewServer(cfg *extprocconfig.Config, exchanger Exchanger, logger *slog.Logger) *Server {
	meter := otel.GetMeterProvider().Meter("extproc")
	requestCounter, err := meter.Int64Counter("extproc.token_exchange.requests",
		metric.WithDescription("Total number of token exchange requests processed by ExtProc"))
	if err != nil {
		logger.Warn("failed to create request counter instrument", "error", err)
	}
	requestDuration, err := meter.Float64Histogram("extproc.token_exchange.duration",
		metric.WithDescription("Duration of token exchange requests in seconds"),
		metric.WithUnit("s"))
	if err != nil {
		logger.Warn("failed to create request duration instrument", "error", err)
	}
	return &Server{
		cfg:             cfg,
		exchanger:       exchanger,
		logger:          logger,
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
	}
}

// NewServerWithAuthorizer creates a new ExtProc Server with OPA authorization enabled.
// Body-bearing requests exchange in the headers phase, then OPA evaluates in the body phase.
// Header-only requests evaluate OPA first and exchange only after allow.
//
// This constructor is used by Phase 3+ implementation and E2E tests.
func NewServerWithAuthorizer(cfg *extprocconfig.Config, exchanger Exchanger, authorizer authorization.Authorizer, logger *slog.Logger) *Server {
	srv := NewServer(cfg, exchanger, logger)
	srv.authorizer = authorizer
	return srv
}

func (s *Server) extractTraceContext(ctx context.Context, headers *extprocv3.HttpHeaders) context.Context {
	if !s.cfg.Telemetry.Enabled {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, (*headerCarrier)(headers))
}

func (s *Server) beginTokenExchangeObservation(ctx context.Context) (context.Context, func(outcome, resourceURI, errorType string)) {
	start := time.Now()
	telemetryEnabled := s.cfg.Telemetry.Enabled
	tracesEnabled := telemetryEnabled && s.cfg.Telemetry.Traces.Enabled
	metricsEnabled := telemetryEnabled && s.cfg.Telemetry.Metrics.Enabled

	var span trace.Span
	if tracesEnabled {
		ctx, span = otel.Tracer("extproc").Start(ctx, "extproc.token_exchange")
	} else {
		span = trace.SpanFromContext(ctx)
	}

	finished := false
	return ctx, func(outcome, resourceURI, errorType string) {
		if finished {
			return
		}
		finished = true

		if resourceURI != "" {
			span.SetAttributes(attribute.String("resource.uri", sanitizeURIForTelemetry(resourceURI)))
		}
		if errorType != "" {
			span.SetAttributes(attribute.String("error.type", errorType))
		}
		span.SetAttributes(attribute.String("outcome", outcome))
		span.End()

		if metricsEnabled {
			outcomeAttr := metric.WithAttributes(attribute.String("outcome", outcome))
			metricCtx := context.WithoutCancel(ctx)
			if s.requestCounter != nil {
				s.requestCounter.Add(metricCtx, 1, outcomeAttr)
			}
			if s.requestDuration != nil {
				s.requestDuration.Record(metricCtx, time.Since(start).Seconds(), outcomeAttr)
			}
		}
	}
}

// Process implements the streaming ExtProc gRPC RPC.
// When OPA is disabled (authorizer == nil):
//   - RequestHeaders: extract Bearer + path → token exchange → replace header
//   - All other phases: pass through unchanged
//
// When OPA is enabled (authorizer != nil) and Bearer token is present:
//   - Body-bearing requests: RequestHeaders performs token exchange, sets header mutation, and requests BUFFERED body; RequestBody evaluates OPA and echoes or denies
//   - Header-only requests: RequestHeaders evaluates OPA first, then performs token exchange on allow
//   - All other phases: pass through unchanged
func (s *Server) Process(stream extprocv3.ExternalProcessor_ProcessServer) error {
	// Per-stream OPA state populated by a preceding RequestHeaders message on the same
	// ExtProc stream. It remains nil for OPA-disabled requests, Bearer-less passthrough,
	// and any RequestBody message that arrives without an earlier headers phase.
	var state *requestState

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			if status.Code(err) == codes.Canceled {
				return nil
			}
			s.logger.DebugContext(stream.Context(), "stream recv error", "error", err)
			return err
		}

		var resp *extprocv3.ProcessingResponse

		switch msg := req.Request.(type) {
		case *extprocv3.ProcessingRequest_RequestHeaders:
			if s.authorizer != nil {
				resp, state = s.processRequestHeadersOPA(stream.Context(), msg.RequestHeaders.EndOfStream, req, msg.RequestHeaders)
				if state != nil && msg.RequestHeaders.EndOfStream {
					// No body phase will follow. For MCP only GET (SSE streams) and POST
					// (JSON-RPC) are valid transports. POST without a body is malformed (400).
					// All other methods are unsupported (405).
					if state.protocol == "mcp" && !isMCPHeaderOnlyMethod(state.headers[":method"]) {
						outcome := "invalid_request"
						errorType := "invalid_method"
						if isBodyBearingMethod(state.headers[":method"]) {
							resp = immediateResponse(httpv3.StatusCode_BadRequest,
								`{"error":"invalid_request","error_description":"MCP request must have a body"}`)
							errorType = "missing_body"
						} else {
							resp = immediateResponseWithHeaders(httpv3.StatusCode_MethodNotAllowed,
								`{"error":"invalid_request","error_description":"method not supported for MCP"}`,
								map[string]string{"allow": "GET, POST"})
						}
						if state.finishObservation != nil {
							state.finishObservation(outcome, state.resourceURI, errorType)
						}
						state = nil
					} else {
						requestCtx := stream.Context()
						if state.requestContext != nil {
							requestCtx = state.requestContext
						}
						resp = s.processHeadersOnlyOPA(requestCtx, state)
						state = nil
					}
				}
			} else {
				resp = s.processRequestHeaders(stream.Context(), req, msg.RequestHeaders)
			}

		case *extprocv3.ProcessingRequest_RequestBody:
			if state != nil {
				bodyCtx := stream.Context()
				if state.requestContext != nil {
					bodyCtx = state.requestContext
				}
				resp = s.processRequestBody(bodyCtx, state, msg.RequestBody)
				state = nil // consumed
			} else {
				// No OPA state from a preceding RequestHeaders phase on this stream: echo the
				// body unchanged. This preserves pass-through behavior for non-OPA traffic and
				// for any body message that arrives without prior per-stream state.
				resp = echoRequestBody(msg.RequestBody)
			}

		case *extprocv3.ProcessingRequest_ResponseHeaders:
			resp = passThroughResponseHeaders()

		case *extprocv3.ProcessingRequest_ResponseBody:
			resp = echoResponseBody(msg.ResponseBody)

		case *extprocv3.ProcessingRequest_RequestTrailers:
			resp = passThroughRequestTrailers()

		case *extprocv3.ProcessingRequest_ResponseTrailers:
			resp = passThroughResponseTrailers()

		default:
			resp = passThrough()
		}

		if err := stream.Send(resp); err != nil {
			s.logger.InfoContext(stream.Context(), "stream send error", "error", err)
			return err
		}

		// After an ImmediateResponse the stream is complete.
		if _, isImmediate := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse); isImmediate {
			return nil
		}
	}
}

// processRequestHeaders implements the primary ExtProc processing phase (OPA disabled).
// It handles:
//   - No Bearer token → pass through (FR-009)
//   - Empty/invalid :path → 503 ImmediateResponse (FR-013)
//   - Successful exchange → replace Authorization header (FR-007)
//   - Re-auth required (broker error_uri) → URLElicitationRequiredError for MCP, 503 for other protocols
//   - Exchange failure → 500 ImmediateResponse (FR-008, FR-010)
//
// ctx carries trace context and must be passed to all blocking operations.
func (s *Server) processRequestHeaders(ctx context.Context, req *extprocv3.ProcessingRequest, headers *extprocv3.HttpHeaders) *extprocv3.ProcessingResponse {
	start := time.Now()
	outcome := "success"

	telemetryEnabled := s.cfg.Telemetry.Enabled
	tracesEnabled := telemetryEnabled && s.cfg.Telemetry.Traces.Enabled
	metricsEnabled := telemetryEnabled && s.cfg.Telemetry.Metrics.Enabled

	// Extract trace context from incoming request headers before starting span.
	// In ExtProc, traceparent/baggage live in HttpHeaders, not gRPC stream context.
	// Use the globally registered propagator so all configured formats (tracecontext,
	// b3, b3multi, ottrace, baggage) are honoured, not just the hard-coded defaults.
	if telemetryEnabled {
		carrier := (*headerCarrier)(headers)
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	}

	// FR-003: Start span for token exchange with extracted trace context.
	// When traces are disabled, use a no-op span so downstream span.SetAttributes
	// calls remain safe without additional guards throughout the function.
	var span trace.Span
	if tracesEnabled {
		ctx, span = otel.Tracer("extproc").Start(ctx, "extproc.token_exchange")
	} else {
		span = trace.SpanFromContext(ctx)
	}
	defer func() {
		span.SetAttributes(attribute.String("outcome", outcome))
		span.End()
		if metricsEnabled {
			outcomeAttr := metric.WithAttributes(attribute.String("outcome", outcome))
			// Use context.WithoutCancel for metric recording — metrics must be recorded
			// even if the gRPC stream was cancelled (client disconnect).
			metricCtx := context.WithoutCancel(ctx)
			if s.requestCounter != nil {
				s.requestCounter.Add(metricCtx, 1, outcomeAttr)
			}
			if s.requestDuration != nil {
				s.requestDuration.Record(metricCtx, time.Since(start).Seconds(), outcomeAttr)
			}
		}
	}()

	bearerToken := extractBearerToken(headers)
	if bearerToken == "" {
		// FR-009: No Bearer token — pass through without modification.
		outcome = "passthrough"
		s.logger.DebugContext(ctx, "no Bearer token — passing through")
		return passThrough()
	}

	path := extractHeader(headers, ":path")
	scheme := extractHeader(headers, ":scheme")
	authority := extractHeader(headers, ":authority")
	resourceURI := buildResourceURI(scheme, authority, path)

	// Sanitize the URI for telemetry (SR-001: no token values in query strings).
	sanitizedURI := sanitizeURIForTelemetry(resourceURI)

	if err := validateResourceURI(resourceURI); err != nil {
		s.logger.WarnContext(ctx, "extproc: invalid resource URI — rejecting with 503",
			"scheme", scheme,
			"authority", authority,
			"resource_uri", sanitizedURI,
			"error", err)
		outcome = "invalid_resource"
		span.SetAttributes(
			attribute.String("resource.uri", sanitizedURI),
			attribute.String("error.type", "invalid_resource"),
		)
		return immediateResponse(httpv3.StatusCode_ServiceUnavailable,
			`{"error":"invalid_resource","error_description":"request URI is empty or invalid"}`)
	}

	span.SetAttributes(attribute.String("resource.uri", sanitizedURI))

	// Extract protocol; default to "mcp" when absent. The non-OPA token-exchange
	// path is MCP-only by design — deployments without protocol metadata are
	// pre-OPA MCP clients. Explicit non-MCP values (e.g. "a2a") are respected.
	protocol, _ := extractProtocolFromMetadata(req)
	if protocol == "" {
		protocol = "mcp"
	}

	result, err := s.exchanger.Exchange(ctx, bearerToken, resourceURI)
	if err != nil {
		resp, mappedOutcome, mappedErrorType := s.exchangeErrorResponse(ctx, "", protocol, resourceURI, err)
		outcome = mappedOutcome
		span.SetAttributes(attribute.String("error.type", mappedErrorType))
		return resp
	}
	s.logger.DebugContext(ctx, "token exchanged successfully", "resource", sanitizedURI)
	return replaceAuthorizationHeader("Bearer " + result.Token)
}

// processRequestHeadersOPA handles the RequestHeaders phase when OPA is enabled.
//
// For body-bearing requests (endOfStream=false): token exchange runs eagerly here so
// that the Authorization mutation is part of the headers-phase response. Proxies such
// as agentgateway only apply header mutations from the first (headers-phase) ExtProc
// response; body-phase mutations are silently dropped by those proxies.
//
// For header-only requests (endOfStream=true): exchange is deferred to
// processHeadersOnlyOPA so that OPA can gate the exchange before any broker call.
//
// Returns (ImmediateResponse, nil) on validation/exchange error, or (response, state).
func (s *Server) processRequestHeadersOPA(ctx context.Context, endOfStream bool, req *extprocv3.ProcessingRequest, headers *extprocv3.HttpHeaders) (*extprocv3.ProcessingResponse, *requestState) {
	ctx = s.extractTraceContext(ctx, headers)

	bearerToken := extractBearerToken(headers)
	if bearerToken == "" {
		s.logger.DebugContext(ctx, "OPA mode: no Bearer token — passing through")
		return passThrough(), nil
	}

	ctx, finishObservation := s.beginTokenExchangeObservation(ctx)
	finishNow := true
	outcome := "success"
	errorType := ""
	path := extractHeader(headers, ":path")
	scheme := extractHeader(headers, ":scheme")
	authority := extractHeader(headers, ":authority")
	resourceURI := buildResourceURI(scheme, authority, path)
	defer func() {
		if finishNow {
			finishObservation(outcome, resourceURI, errorType)
		}
	}()

	if err := validateResourceURI(resourceURI); err != nil {
		outcome = "invalid_resource"
		errorType = "invalid_resource"
		s.logger.WarnContext(ctx, "extproc OPA: invalid resource URI — rejecting with 503",
			"path", path, "scheme", scheme, "authority", authority,
			"resource_uri", sanitizeURIForTelemetry(resourceURI), "error", err)
		return immediateResponse(httpv3.StatusCode_ServiceUnavailable,
			`{"error":"invalid_resource","error_description":"request URI is empty or invalid"}`), nil
	}

	protocol, ok := extractProtocolFromMetadata(req)
	if !ok {
		outcome = "authorization_denied"
		errorType = "missing_protocol_metadata"
		s.logger.WarnContext(ctx, "OPA: protocol metadata absent — rejecting with 403 (misconfiguration)",
			"resource", sanitizeURIForTelemetry(resourceURI))
		return immediateResponse(httpv3.StatusCode_Forbidden,
			`{"error":"access_denied","error_description":"protocol metadata is required when authorization is enabled"}`), nil
	}

	headerMap := extractAllHeaders(headers)

	if protocol == "mcp" && !isMCPHeaderOnlyMethod(headerMap[":method"]) && !isBodyBearingMethod(headerMap[":method"]) {
		outcome = "invalid_request"
		errorType = "invalid_method"
		return immediateResponseWithHeaders(httpv3.StatusCode_MethodNotAllowed,
			`{"error":"invalid_request","error_description":"method not supported for MCP"}`,
			map[string]string{"allow": "GET, POST"}), nil
	}

	state := &requestState{
		bearerToken:       bearerToken,
		resourceURI:       resourceURI,
		headers:           headerMap,
		protocol:          protocol,
		requestContext:    ctx,
		finishObservation: finishObservation,
	}

	if endOfStream {
		finishNow = false
		return passThrough(), state
	}

	exchangeResult, exchErr := s.exchanger.Exchange(ctx, bearerToken, resourceURI)
	if exchErr != nil {
		resp, mappedOutcome, mappedErrorType := s.exchangeErrorResponse(ctx, "OPA: headers-phase", protocol, resourceURI, exchErr)
		outcome = mappedOutcome
		errorType = mappedErrorType
		return resp, nil
	}
	state.grantedPermissionSets = exchangeResult.GrantedPermissionSets
	s.logger.DebugContext(ctx, "OPA: token exchanged in headers phase, buffering body for OPA evaluation",
		"resource", sanitizeURIForTelemetry(resourceURI))
	return requestBodyBufferingResponseWithAuth("Bearer " + exchangeResult.Token), state
}

// processRequestBody handles the RequestBody phase when OPA is enabled.
// Token exchange has already occurred in the headers phase; this phase only
// evaluates OPA policy using the body and returns:
//   - allow: echo the request body unchanged (Authorization was already set)
//   - deny: 403 ImmediateResponse with JSON access_denied body
func (s *Server) processRequestBody(ctx context.Context, state *requestState, body *extprocv3.HttpBody) *extprocv3.ProcessingResponse {
	var bodyBytes []byte
	if body != nil {
		bodyBytes = body.Body
	}
	sanitizedURI := sanitizeURIForTelemetry(state.resourceURI)

	if state.protocol == "mcp" && isMCPHeaderOnlyMethod(state.headers[":method"]) {
		return immediateResponseWithHeaders(httpv3.StatusCode_MethodNotAllowed,
			`{"error":"invalid_request","error_description":"method not supported for MCP"}`,
			map[string]string{"allow": "GET, POST"})
	}

	maxSize := s.cfg.Authorization.MaxBodySize
	if maxSize > 0 && len(bodyBytes) > maxSize {
		s.logger.WarnContext(ctx, "OPA: request body exceeds max_body_size — rejecting",
			"resource", sanitizedURI, "body_size", len(bodyBytes), "max_body_size", maxSize)
		return immediateResponse(httpv3.StatusCode_Forbidden,
			`{"error":"request_too_large","error_description":"request body exceeds maximum allowed size"}`)
	}

	if state.protocol == "mcp" {
		if trimmed := bytes.TrimLeft(bodyBytes, " \t\r\n"); len(trimmed) > 0 && trimmed[0] == '[' {
			return s.processRequestBodyBatch(ctx, state, bodyBytes, body)
		}
	}

	opaInput, buildErr := authorization.BuildOPAInput(state.protocol, bodyBytes, state.headers, state.grantedPermissionSets)
	if buildErr != nil {
		s.logger.WarnContext(ctx, "OPA: failed to parse request body — denying", "resource", sanitizedURI, "error", buildErr)
		return immediateResponse(httpv3.StatusCode_Forbidden,
			`{"error":"access_denied","error_description":"failed to parse request protocol"}`)
	}

	decision, err := s.authorizer.Evaluate(ctx, opaInput)
	if err != nil {
		s.logger.ErrorContext(ctx, "OPA evaluation error — denying", "resource", sanitizedURI, "error", err)
		return immediateResponse(httpv3.StatusCode_Forbidden,
			`{"error":"access_denied","error_description":"authorization evaluation failed"}`)
	}

	if decision.Action != "allow" {
		s.logger.InfoContext(ctx, "OPA denied request", "reasons", decision.Reasons, "resource", sanitizedURI)
		return accessDeniedResponse(decision.Reasons)
	}

	s.logger.DebugContext(ctx, "OPA allowed request, echoing body", "resource", sanitizedURI)
	return echoRequestBody(body)
}

// processHeadersOnlyOPA handles header-only requests in OPA mode (end_of_stream=true in
// the headers phase, meaning no body phase will follow). It evaluates OPA with an empty
// body and performs token exchange, returning a headers-phase response on success.
// Deny and exchange errors return ImmediateResponse, identical to the body path.
func (s *Server) processHeadersOnlyOPA(ctx context.Context, state *requestState) *extprocv3.ProcessingResponse {
	outcome := "success"
	errorType := ""
	if state.finishObservation != nil {
		defer func() {
			state.finishObservation(outcome, state.resourceURI, errorType)
		}()
	}
	sanitizedURI := sanitizeURIForTelemetry(state.resourceURI)

	opaInput, buildErr := authorization.BuildOPAInputHeadersOnly(state.protocol, state.headers)
	if buildErr != nil {
		outcome = "authorization_denied"
		errorType = "invalid_request"
		s.logger.WarnContext(ctx, "OPA: failed to build input for header-only request — denying", "resource", sanitizedURI, "error", buildErr)
		return immediateResponse(httpv3.StatusCode_Forbidden,
			`{"error":"access_denied","error_description":"failed to parse request protocol"}`)
	}

	decision, err := s.authorizer.Evaluate(ctx, opaInput)
	if err != nil {
		outcome = "authorization_denied"
		errorType = "evaluation_error"
		s.logger.ErrorContext(ctx, "OPA evaluation error — denying", "resource", sanitizedURI, "error", err)
		return immediateResponse(httpv3.StatusCode_Forbidden,
			`{"error":"access_denied","error_description":"authorization evaluation failed"}`)
	}

	if decision.Action != "allow" {
		outcome = "authorization_denied"
		errorType = "access_denied"
		s.logger.InfoContext(ctx, "OPA denied header-only request", "reasons", decision.Reasons, "resource", sanitizedURI)
		return accessDeniedResponse(decision.Reasons)
	}

	exchangeResult, exchErr := s.exchanger.Exchange(ctx, state.bearerToken, state.resourceURI)
	if exchErr != nil {
		resp, mappedOutcome, mappedErrorType := s.exchangeErrorResponse(ctx, "OPA: header-only", state.protocol, state.resourceURI, exchErr)
		outcome = mappedOutcome
		errorType = mappedErrorType
		return resp
	}
	s.logger.DebugContext(ctx, "OPA allowed header-only request, token exchanged successfully", "resource", sanitizedURI)
	return replaceAuthorizationHeader("Bearer " + exchangeResult.Token)
}

// processRequestBodyBatch evaluates a JSON-RPC batch body (FR-023).
// Each element is evaluated independently; if any is denied the entire batch is denied
// with a 403 response that aggregates reasons from all denying messages.
// Original raw JSON bytes are passed to BuildOPAInput to preserve any extra top-level fields.
func (s *Server) processRequestBodyBatch(ctx context.Context, state *requestState, bodyBytes []byte, body *extprocv3.HttpBody) *extprocv3.ProcessingResponse {
	var rawMessages []json.RawMessage
	sanitizedURI := sanitizeURIForTelemetry(state.resourceURI)
	if err := json.Unmarshal(bodyBytes, &rawMessages); err != nil {
		s.logger.WarnContext(ctx, "OPA: failed to parse batch body — denying", "resource", sanitizedURI, "error", err)
		return immediateResponse(httpv3.StatusCode_Forbidden,
			`{"error":"access_denied","error_description":"failed to parse batch request"}`)
	}
	if len(rawMessages) == 0 {
		s.logger.WarnContext(ctx, "OPA: empty batch body — rejecting as malformed", "resource", sanitizedURI)
		return immediateResponse(httpv3.StatusCode_Forbidden,
			`{"error":"access_denied","error_description":"empty batch is not valid JSON-RPC 2.0"}`)
	}

	var (
		denied      bool
		denyReasons []string
	)
	for i, raw := range rawMessages {
		if _, err := authorization.ParseMCPMessage(raw); err != nil {
			s.logger.WarnContext(ctx, "OPA: invalid batch element — denying", "resource", sanitizedURI, "index", i, "error", err)
			denied = true
			denyReasons = append(denyReasons, "batch element could not be evaluated")
			continue
		}
		opaInput, buildErr := authorization.BuildOPAInput(state.protocol, raw, state.headers, state.grantedPermissionSets)
		if buildErr != nil {
			s.logger.WarnContext(ctx, "OPA: failed to build input for batch element — denying", "resource", sanitizedURI, "index", i, "error", buildErr)
			denied = true
			denyReasons = append(denyReasons, "failed to parse batch element")
			continue
		}
		decision, evalErr := s.authorizer.Evaluate(ctx, opaInput)
		if evalErr != nil {
			s.logger.ErrorContext(ctx, "OPA evaluation error for batch element — denying", "resource", sanitizedURI, "index", i, "error", evalErr)
			denied = true
			denyReasons = append(denyReasons, "authorization evaluation failed")
			continue
		}
		if decision.Action != "allow" {
			denied = true
			if len(decision.Reasons) > 0 {
				denyReasons = append(denyReasons, decision.Reasons...)
			} else {
				denyReasons = append(denyReasons, "access denied")
			}
		}
	}

	if denied {
		s.logger.InfoContext(ctx, "OPA denied batch request", "reasons", denyReasons, "resource", sanitizedURI)
		return accessDeniedResponse(denyReasons)
	}

	s.logger.DebugContext(ctx, "OPA allowed batch, echoing body", "resource", sanitizedURI)
	return echoRequestBody(body)
}

// tokenExchangeErrorResponse maps a token exchange error to the appropriate ImmediateResponse.
// Handles ErrAssertionExpired (503), ErrCircuitOpen (503), and generic failures (500).
// BrokerExchangeError with error_uri is handled by callers before reaching this function.
func (s *Server) tokenExchangeErrorResponse(ctx context.Context, err error, resourceURI string) *extprocv3.ProcessingResponse {
	sanitizedURI := sanitizeURIForTelemetry(resourceURI)
	if errors.Is(err, ErrAssertionExpired) {
		s.logger.ErrorContext(ctx, "token exchange failed: client assertion expired — background refresh may have failed",
			"resource", sanitizedURI)
		return immediateResponse(httpv3.StatusCode_ServiceUnavailable,
			`{"error":"service_unavailable","error_description":"client assertion expired"}`)
	}
	if errors.Is(err, ErrCircuitOpen) {
		s.logger.WarnContext(ctx, "token exchange rejected: circuit breaker is open",
			"resource", sanitizedURI)
		return immediateResponse(httpv3.StatusCode_ServiceUnavailable,
			`{"error":"service_unavailable","error_description":"circuit breaker is open"}`)
	}
	s.logger.ErrorContext(ctx, "token exchange failed", "resource", sanitizedURI, "error", err)
	return immediateResponse(httpv3.StatusCode_InternalServerError,
		`{"error":"token_exchange_failed","error_description":"token exchange request failed"}`)
}

func accessDeniedResponse(reasons []string) *extprocv3.ProcessingResponse {
	body403, _ := json.Marshal(map[string]string{
		"error":             "access_denied",
		"error_description": strings.Join(reasons, "; "),
	})
	return immediateResponse(httpv3.StatusCode_Forbidden, string(body403))
}

func (s *Server) exchangeErrorResponse(ctx context.Context, phase, protocol, resourceURI string, err error) (*extprocv3.ProcessingResponse, string, string) {
	sanitizedURI := sanitizeURIForTelemetry(resourceURI)
	var brokerErr *BrokerExchangeError
	if errors.As(err, &brokerErr) && brokerErr.ErrorURI != "" && !isTransientBrokerError(err) {
		if protocol == "mcp" {
			msg := "token exchange requires re-authentication — returning URLElicitationRequiredError"
			if phase != "" {
				msg = phase + " token exchange requires re-auth — returning URLElicitationRequiredError"
			}
			s.logger.InfoContext(ctx, msg,
				"resource", sanitizedURI,
				"code", brokerErr.Code,
				"error_uri", brokerErr.ErrorURI)
			return urlElicitationResponse(brokerErr, nil), "exchange_failure", brokerErr.Code
		}

		msg := "token exchange requires re-authentication but protocol is non-MCP — returning 503"
		if phase != "" {
			msg = phase + " token exchange requires re-auth but protocol is non-MCP — returning 503"
		}
		s.logger.WarnContext(ctx, msg, "resource", sanitizedURI, "protocol", protocol)
		return immediateResponse(httpv3.StatusCode_ServiceUnavailable,
			`{"error":"service_unavailable","error_description":"token exchange requires re-authentication"}`), "exchange_failure", brokerErr.Code
	}

	switch {
	case errors.Is(err, ErrAssertionExpired):
		return s.tokenExchangeErrorResponse(ctx, err, resourceURI), "assertion_expired", "assertion_expired"
	case errors.Is(err, ErrCircuitOpen):
		return s.tokenExchangeErrorResponse(ctx, err, resourceURI), "circuit_open", "circuit_open"
	default:
		return s.tokenExchangeErrorResponse(ctx, err, resourceURI), "exchange_failure", "exchange_failure"
	}
}

// extractProtocolFromMetadata extracts the agentgateway protocol value from the
// MetadataContext FilterMetadata.
//
// Returns (protocol, true) when the agentgateway metadata key is present and contains
// a non-empty protocol string. The returned protocol may be any value (e.g. "mcp",
// "a2a", or an unrecognised string).
//
// Returns ("", false) when the metadata is entirely absent — i.e. MetadataContext is
// nil, the "agentgateway" key is missing, or the protocol field is absent or empty.
// Callers in OPA mode MUST reject the request with 403 on a false return (FR-003).
//
// The metadata structure is: FilterMetadata["agentgateway"]["protocol"] = "<type>".
func extractProtocolFromMetadata(req *extprocv3.ProcessingRequest) (string, bool) {
	if req.MetadataContext == nil {
		return "", false
	}
	agwMeta, ok := req.MetadataContext.FilterMetadata[agentgatewayProtocolMetadataKey]
	if !ok || agwMeta == nil {
		return "", false
	}
	fields := agwMeta.GetFields()
	if fields == nil {
		return "", false
	}
	protoVal, ok := fields[agentgatewayProtocolFieldKey]
	if !ok || protoVal == nil {
		return "", false
	}
	// GetStringValue returns "" for non-string protobuf Values.
	v := protoVal.GetStringValue()
	if v == "" {
		return "", false
	}
	return v, true
}

// extractAllHeaders returns all request headers as a lowercase-key map.
func extractAllHeaders(headers *extprocv3.HttpHeaders) map[string]string {
	result := make(map[string]string)
	if headers == nil || headers.Headers == nil {
		return result
	}
	for _, h := range headers.Headers.Headers {
		key := strings.ToLower(h.Key)
		if len(h.RawValue) > 0 {
			result[key] = string(h.RawValue)
		} else {
			result[key] = h.Value
		}
	}
	return result
}

// requestBodyBufferingResponseWithAuth returns a HeadersResponse that sets the exchanged
// Authorization header and instructs Envoy/proxy to buffer the full request body for the
// body phase. By including the auth mutation in the headers-phase response, proxies that
// only apply header mutations from the first ExtProc response (e.g. agentgateway) will
// correctly forward the exchanged token to the upstream.
func requestBodyBufferingResponseWithAuth(authValue string) *extprocv3.ProcessingResponse {
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extprocv3.HeadersResponse{
				Response: &extprocv3.CommonResponse{
					HeaderMutation: &extprocv3.HeaderMutation{
						SetHeaders: []*corev3.HeaderValueOption{
							{
								Header: &corev3.HeaderValue{
									Key:      "authorization",
									RawValue: []byte(authValue),
								},
							},
						},
					},
				},
			},
		},
		ModeOverride: &extprocfilterv3.ProcessingMode{
			RequestBodyMode: extprocfilterv3.ProcessingMode_BUFFERED,
		},
	}
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
// extractHeader returns the first matching header value (case-insensitive key match),
// preferring RawValue over Value. Returns on first key match, even if both RawValue
// and Value are empty — matching http.Header.Get() first-match semantics. Envoy
// guarantees at least one of RawValue/Value is populated for headers it forwards.
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
		// SR-001: Do not echo the raw URI in error messages — it may contain
		// sensitive query parameters. Return a generic parse failure instead.
		return status.Error(codes.InvalidArgument, "invalid :path: URI parse error")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		// SR-001: Do not echo the parsed scheme — input may contain
		// unexpected URI schemes (e.g. data:, javascript:).
		return status.Error(codes.InvalidArgument, ":path must have http or https scheme")
	}
	if u.Host == "" {
		return status.Error(codes.InvalidArgument, ":path must have a non-empty host")
	}
	return nil
}

// headerCarrier adapts ExtProc headers to the OTel TextMapCarrier interface.
// Get is key-aware: list-valued propagation headers (baggage, tracestate) are
// comma-joined per RFC 9110 §5.2; single-valued headers (traceparent, b3, etc.)
// return first-match only, consistent with http.Header.Get().
type headerCarrier extprocv3.HttpHeaders

// listValuedPropagationHeaders are propagation headers whose spec allows
// multiple header fields to be combined with commas (RFC 9110 §5.2).
var listValuedPropagationHeaders = map[string]bool{
	"baggage":    true,
	"tracestate": true,
}

func (c *headerCarrier) Get(key string) string {
	if !listValuedPropagationHeaders[strings.ToLower(key)] {
		return extractHeader((*extprocv3.HttpHeaders)(c), key)
	}
	if c == nil || c.Headers == nil {
		return ""
	}
	keyLower := strings.ToLower(key)
	var vals []string
	for _, h := range c.Headers.Headers {
		if strings.ToLower(h.Key) == keyLower {
			if len(h.RawValue) > 0 {
				vals = append(vals, string(h.RawValue))
			} else if h.Value != "" {
				vals = append(vals, h.Value)
			}
		}
	}
	return strings.Join(vals, ",")
}

func (c *headerCarrier) Set(key string, value string) {
	// Extraction-only carrier: Set is intentionally a no-op because ExtProc
	// responses don't inject trace headers. If Inject() is called on this
	// carrier (via otel.GetTextMapPropagator().Inject()), injected headers
	// will be silently dropped. Implement Set with header mutation if
	// response header injection becomes needed.
}

func (c *headerCarrier) Keys() []string {
	if c == nil || c.Headers == nil {
		return []string{}
	}
	seen := make(map[string]struct{}, len(c.Headers.Headers))
	keys := make([]string, 0, len(c.Headers.Headers))
	for _, h := range c.Headers.Headers {
		lower := strings.ToLower(h.Key)
		if _, dup := seen[lower]; !dup {
			seen[lower] = struct{}{}
			keys = append(keys, h.Key)
		}
	}
	return keys
}

// sanitizeURIForTelemetry removes query strings and fragments from a URI
// before recording it in telemetry, to prevent leaking tokens or other
// sensitive parameters in violation of SR-001.
// On parse failure it strips everything from '?' or '#' onwards conservatively,
// rather than returning the raw URI which may contain sensitive query parameters.
func sanitizeURIForTelemetry(resourceURI string) string {
	u, err := url.ParseRequestURI(resourceURI)
	if err != nil {
		// Conservative fallback: strip query string and fragment by truncating at
		// the first '?' or '#' to avoid leaking sensitive parameters in span attributes.
		if i := strings.IndexAny(resourceURI, "?#"); i >= 0 {
			return resourceURI[:i]
		}
		return resourceURI
	}
	// Create a copy with no query or fragment
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
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
// HTTP status code and JSON body. Used for error responses (403, 500, 503).
// isMCPHeaderOnlyMethod reports whether an HTTP method is a valid MCP transport
// that legitimately carries no request body. Only GET is explicitly allowed:
// it is used for SSE stream connections and protocol upgrade handshakes.
func isMCPHeaderOnlyMethod(method string) bool {
	return strings.ToUpper(method) == "GET"
}

// isBodyBearingMethod reports whether an HTTP method is the MCP JSON-RPC transport method.
// MCP uses POST exclusively for JSON-RPC messages.
func isBodyBearingMethod(method string) bool {
	return strings.ToUpper(method) == "POST"
}

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

// immediateResponseWithHeaders builds an ImmediateResponse with extra response headers
// in addition to the standard content-type. Used for 405 responses that require Allow.
func immediateResponseWithHeaders(code httpv3.StatusCode, body string, extra map[string]string) *extprocv3.ProcessingResponse {
	headers := []*corev3.HeaderValueOption{
		{Header: &corev3.HeaderValue{Key: "content-type", RawValue: []byte("application/json")}},
	}
	for k, v := range extra {
		headers = append(headers, &corev3.HeaderValueOption{
			Header: &corev3.HeaderValue{Key: k, RawValue: []byte(v)},
		})
	}
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ImmediateResponse{
			ImmediateResponse: &extprocv3.ImmediateResponse{
				Status:  &httpv3.HttpStatus{Code: code},
				Headers: &extprocv3.HeaderMutation{SetHeaders: headers},
				Body:    []byte(body),
			},
		},
	}
}

// urlElicitationResponse builds a ProcessingResponse_ImmediateResponse with HTTP 200 and
// a JSON-RPC 2.0 URLElicitationRequiredError body (error code -32042).
// Per MCP spec 2025-11-05: returned when a request cannot proceed until the user visits
// a URL for OAuth re-authentication.
// HTTP 200 is used because JSON-RPC errors always travel over HTTP 200.
// rawID is the raw JSON bytes of the request ID (e.g. `42`, `"req-1"`, `null`).
// Pass nil to emit null — correct when the request ID is unknown.
func urlElicitationResponse(brokerErr *BrokerExchangeError, rawID json.RawMessage) *extprocv3.ProcessingResponse {
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
	// Set the request ID using the raw JSON token so any valid JSON-RPC ID type
	// (string, integer, null) is preserved exactly without numeric precision loss.
	// json.RawMessage.MarshalJSON returns its bytes verbatim, so mcp.NewRequestId
	// will serialise the ID token unchanged.
	if rawID != nil {
		jsonRPCErr.ID = mcp.NewRequestId(rawID)
	}
	// Override the generated message with the broker-supplied description so that
	// the client receives context about why re-authentication is required.
	jsonRPCErr.Error.Message = brokerErr.Description

	// Marshal the response. json.Marshal cannot fail for this struct: all fields are strings,
	// ints, or slices thereof — no encoding/json.Marshaler implementations that could error.
	body, _ := json.Marshal(jsonRPCErr)
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
