package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/authorization"
	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/server"
)

// testLogger returns a discard logger for unit tests.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// testConfig returns a minimal valid config for unit tests.
func testConfig() *extprocconfig.Config {
	return &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:       "https://idp.example.com/oauth2/token",
			Issuer:              "https://idp.example.com",
			ClientID:            "test-client",
			ClientSecret:        "test-secret",
			ClientAssertionType: "id_token",
			ExchangeTimeout:     5 * time.Second,
			TLS:                 extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
		CircuitBreaker: extprocconfig.CircuitBreakerConfig{
			Enabled:      true,
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}
}

// testConfigWithTelemetry returns a config with full telemetry enabled for observability tests.
func testConfigWithTelemetry() *extprocconfig.Config {
	cfg := testConfig()
	cfg.Telemetry = extprocconfig.TelemetryConfig{
		Enabled: true,
		Traces:  extprocconfig.TracesConfig{Enabled: true},
		Metrics: extprocconfig.MetricsConfig{Enabled: true},
	}
	return cfg
}

// mockExchanger is a controllable Exchanger for unit tests.
type mockExchanger struct {
	exchangeFunc   func(ctx context.Context, subjectToken, resourceURI string) (server.ExchangeResult, error)
	shutdownCalled bool
}

func (m *mockExchanger) Exchange(ctx context.Context, subjectToken, resourceURI string) (server.ExchangeResult, error) {
	return m.exchangeFunc(ctx, subjectToken, resourceURI)
}

func (m *mockExchanger) Shutdown() {
	m.shutdownCalled = true
}

// startTestServer registers the Server on a random in-process port and returns
// a connected client + cleanup function.
func startTestServer(t *testing.T, exchanger server.Exchanger) (extprocv3.ExternalProcessorClient, func()) {
	t.Helper()
	return startTestServerWithConfig(t, testConfig(), exchanger)
}

// startTestServerWithConfig registers the Server with the given config on a random in-process
// port and returns a connected client + cleanup function.
func startTestServerWithConfig(t *testing.T, cfg *extprocconfig.Config, exchanger server.Exchanger) (extprocv3.ExternalProcessorClient, func()) {
	t.Helper()

	svc := server.NewServer(cfg, exchanger, testLogger())

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	extprocv3.RegisterExternalProcessorServer(grpcSrv, svc)

	go func() {
		_ = grpcSrv.Serve(listener)
	}()

	conn, err := grpc.NewClient(listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := extprocv3.NewExternalProcessorClient(conn)

	cleanup := func() {
		_ = conn.Close()
		grpcSrv.GracefulStop()
	}
	return client, cleanup
}

// sendRequestHeaders opens a Process stream, sends a RequestHeaders message,
// and returns the first ProcessingResponse.
func sendRequestHeaders(t *testing.T, client extprocv3.ExternalProcessorClient, headers map[string]string) (*extprocv3.ProcessingResponse, error) {
	t.Helper()
	return sendRequestHeadersWithProtocol(t, client, headers, "")
}

// sendRequestHeadersWithProtocol is like sendRequestHeaders but also attaches
// agentgateway protocol metadata when protocol is non-empty.
func sendRequestHeadersWithProtocol(t *testing.T, client extprocv3.ExternalProcessorClient, headers map[string]string, protocol string) (*extprocv3.ProcessingResponse, error) {
	t.Helper()

	stream, err := client.Process(context.Background())
	require.NoError(t, err)

	headerList := make([]*corev3.HeaderValue, 0, len(headers))
	for k, v := range headers {
		headerList = append(headerList, &corev3.HeaderValue{Key: k, RawValue: []byte(v)})
	}

	req := &extprocv3.ProcessingRequest{
		Request: &extprocv3.ProcessingRequest_RequestHeaders{
			RequestHeaders: &extprocv3.HttpHeaders{
				Headers: &corev3.HeaderMap{Headers: headerList},
			},
		},
	}
	if protocol != "" {
		protocolStruct, sErr := structpb.NewStruct(map[string]any{"protocol": protocol})
		require.NoError(t, sErr)
		req.MetadataContext = &corev3.Metadata{
			FilterMetadata: map[string]*structpb.Struct{"agentgateway": protocolStruct},
		}
	}

	if err = stream.Send(req); err != nil {
		return nil, err
	}
	_ = stream.CloseSend()

	resp, err := stream.Recv()
	return resp, err
}

// sendRequestHeadersWithProtocolEOS sends a RequestHeaders message with EndOfStream=true
// (no body phase will follow), attaches agentgateway protocol metadata when protocol is
// non-empty, and returns the single ProcessingResponse.
func sendRequestHeadersWithProtocolEOS(t *testing.T, client extprocv3.ExternalProcessorClient, headers map[string]string, protocol string) (*extprocv3.ProcessingResponse, error) {
	t.Helper()

	stream, err := client.Process(context.Background())
	require.NoError(t, err)

	headerList := make([]*corev3.HeaderValue, 0, len(headers))
	for k, v := range headers {
		headerList = append(headerList, &corev3.HeaderValue{Key: k, RawValue: []byte(v)})
	}

	req := &extprocv3.ProcessingRequest{
		Request: &extprocv3.ProcessingRequest_RequestHeaders{
			RequestHeaders: &extprocv3.HttpHeaders{
				Headers:     &corev3.HeaderMap{Headers: headerList},
				EndOfStream: true,
			},
		},
	}
	if protocol != "" {
		protocolStruct, sErr := structpb.NewStruct(map[string]any{"protocol": protocol})
		require.NoError(t, sErr)
		req.MetadataContext = &corev3.Metadata{
			FilterMetadata: map[string]*structpb.Struct{"agentgateway": protocolStruct},
		}
	}

	if err = stream.Send(req); err != nil {
		return nil, err
	}
	_ = stream.CloseSend()

	resp, err := stream.Recv()
	return resp, err
}

// ---------------------------------------------------------------------------
// T015: ExtProc gRPC streaming logic — RequestHeaders processing
// ---------------------------------------------------------------------------

// Spec: FR-009 — No Bearer token → pass through unchanged
func TestServer_Process_NoBearerToken_PassThrough(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			t.Fatal("Exchange should not be called when no Bearer token is present")
			return server.ExchangeResult{}, nil
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":   "http://mcp-server:9003/mcp",
		":method": "POST",
	})
	require.NoError(t, err)

	// Expect a RequestHeaders response with no mutations (pass-through)
	headersResp, ok := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders)
	require.True(t, ok, "expected RequestHeaders response for pass-through")

	if headersResp.RequestHeaders != nil && headersResp.RequestHeaders.Response != nil {
		mutation := headersResp.RequestHeaders.Response.HeaderMutation
		if mutation != nil {
			assert.Empty(t, mutation.SetHeaders, "pass-through should not set any headers")
		}
	}
}

// Spec: FR-009 — Non-Bearer authorization header → pass through unchanged
func TestServer_Process_NonBearerAuth_PassThrough(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			t.Fatal("Exchange should not be called for non-Bearer auth")
			return server.ExchangeResult{}, nil
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://mcp-server:9003/mcp",
		"authorization": "Basic dXNlcjpwYXNz",
	})
	require.NoError(t, err)

	_, ok := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders)
	assert.True(t, ok, "non-Bearer auth should produce a pass-through RequestHeaders response")
}

// Spec: FR-002, FR-003, FR-007 — Valid Bearer token → exchanged token replaces Authorization
func TestServer_Process_BearerToken_ReplacesAuthorizationHeader(t *testing.T) {
	const subjectToken = "incoming-bearer-token"
	const exchangedToken = "exchanged-downstream-token"
	const resourceURI = "http://mcp-server:9003/mcp"

	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, st, ru string) (server.ExchangeResult, error) {
			assert.Equal(t, subjectToken, st, "Exchange must receive the incoming Bearer token")
			assert.Equal(t, resourceURI, ru, "Exchange must receive the :path as resource URI")
			return server.ExchangeResult{Token: exchangedToken}, nil
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         resourceURI,
		"authorization": "Bearer " + subjectToken,
	})
	require.NoError(t, err)

	headersResp, ok := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders)
	require.True(t, ok, "successful exchange should produce a RequestHeaders response")
	require.NotNil(t, headersResp.RequestHeaders)
	require.NotNil(t, headersResp.RequestHeaders.Response)
	require.NotNil(t, headersResp.RequestHeaders.Response.HeaderMutation)

	setHeaders := headersResp.RequestHeaders.Response.HeaderMutation.SetHeaders
	require.NotEmpty(t, setHeaders, "exchanged token must be set in headers")

	var authValue string
	for _, h := range setHeaders {
		if h.Header.Key == "authorization" {
			authValue = string(h.Header.RawValue)
		}
	}
	assert.Equal(t, "Bearer "+exchangedToken, authValue,
		"Authorization header must be replaced with exchanged token")
}

// Spec: FR-008, FR-010 — Token exchange failure → 500 ImmediateResponse
func TestServer_Process_ExchangeFailure_Returns500(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, errors.New("exchange endpoint returned 403")
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://mcp-server:9003/mcp",
		"authorization": "Bearer some-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "exchange failure must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_InternalServerError),
		int32(immResp.ImmediateResponse.Status.Code),
		"exchange failure must return HTTP 500")
	assert.Contains(t, string(immResp.ImmediateResponse.Body),
		"token_exchange_failed", "error body must contain token_exchange_failed")
}

// Spec: FR-013 — Empty :path → 503 ImmediateResponse
func TestServer_Process_EmptyPath_Returns503(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			t.Fatal("Exchange should not be called with empty :path")
			return server.ExchangeResult{}, nil
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "",
		"authorization": "Bearer some-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "empty :path must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_ServiceUnavailable),
		int32(immResp.ImmediateResponse.Status.Code),
		"empty :path must return HTTP 503")
	assert.Contains(t, string(immResp.ImmediateResponse.Body),
		"invalid_resource", "error body must contain invalid_resource")
}

// Spec: FR-013 — Relative :path (not absolute URI) → 503 ImmediateResponse (SSRF mitigation)
func TestServer_Process_RelativePath_Returns503(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			t.Fatal("Exchange should not be called with relative :path")
			return server.ExchangeResult{}, nil
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "/mcp",
		"authorization": "Bearer some-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "relative :path must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_ServiceUnavailable),
		int32(immResp.ImmediateResponse.Status.Code),
		"relative :path must return HTTP 503")
}

// Spec: FR-013 — Non-http(s) :path → 503 ImmediateResponse (SSRF mitigation)
func TestServer_Process_NonHTTPPath_Returns503(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			t.Fatal("Exchange should not be called with non-http :path")
			return server.ExchangeResult{}, nil
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "grpc://mcp-server:9003/mcp",
		"authorization": "Bearer some-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "non-http :path must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_ServiceUnavailable),
		int32(immResp.ImmediateResponse.Status.Code))
}

// SR-001 — Malformed :path with sensitive query params must return 503 without leaking secrets.
// The actual query-string-in-error regression test is in security_test.go (white-box).
func TestServer_Process_MalformedPathWithSecrets_Returns503(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			t.Fatal("Exchange should not be called with malformed :path")
			return server.ExchangeResult{}, nil
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "https://example.com/%zz?access_token=secret123",
		"authorization": "Bearer some-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "malformed :path must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_ServiceUnavailable),
		int32(immResp.ImmediateResponse.Status.Code))
}

// SR-003 — Exchanged token details must not appear in error responses (information disclosure)
func TestServer_Process_ErrorResponse_DoesNotLeakTokenDetails(t *testing.T) {
	const internalError = "upstream returned 403: access_denied for client xyz"
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, errors.New(internalError)
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://mcp-server:9003/mcp",
		"authorization": "Bearer secret-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok)
	assert.NotContains(t, string(immResp.ImmediateResponse.Body), internalError,
		"internal error details must not be exposed in response body")
	assert.NotContains(t, string(immResp.ImmediateResponse.Body), "secret-token",
		"token values must not appear in error response body")
}

// Spec: FR-006 — expired client assertion → 503 ImmediateResponse
func TestServer_Process_ExpiredAssertion_Returns503(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, server.ErrAssertionExpired
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://mcp-server:9003/mcp",
		"authorization": "Bearer some-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "expired assertion must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_ServiceUnavailable),
		int32(immResp.ImmediateResponse.Status.Code),
		"expired assertion must return HTTP 503")
	assert.Contains(t, string(immResp.ImmediateResponse.Body),
		"service_unavailable", "error body must contain service_unavailable")
}

// Spec: FR-001 — Server must implement ExternalProcessorServer interface
func TestServer_ImplementsExternalProcessorServer(t *testing.T) {
	cfg := testConfig()
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, nil
		},
	}
	svc := server.NewServer(cfg, exchanger, testLogger())

	// Compile-time interface check
	var _ extprocv3.ExternalProcessorServer = svc
}

// ---------------------------------------------------------------------------
// T023: OPA authorization flow — server integration tests
// ---------------------------------------------------------------------------

// mockAuthorizer is a controllable authorization.Authorizer for unit tests.
type mockAuthorizer struct {
	evaluateFunc func(ctx context.Context, input authorization.OPAInput) (*authorization.OPADecision, error)
	stopCalled   bool
}

func (m *mockAuthorizer) Evaluate(ctx context.Context, input authorization.OPAInput) (*authorization.OPADecision, error) {
	return m.evaluateFunc(ctx, input)
}

func (m *mockAuthorizer) Stop(_ context.Context) {
	m.stopCalled = true
}

// startTestServerWithAuthorizer registers a Server (with OPA) on a random port.
func startTestServerWithAuthorizer(t *testing.T, exchanger server.Exchanger, auth authorization.Authorizer) (extprocv3.ExternalProcessorClient, func()) {
	t.Helper()

	cfg := testConfig()
	cfg.Authorization = extprocconfig.AuthorizationConfig{
		Enabled:           true,
		Policy:            extprocconfig.PolicyConfig{Package: "aib.extproc.authz", Decision: "result"},
		DefaultDecision:   "deny",
		EvaluationTimeout: 5 * time.Second,
		MaxBodySize:       1048576,
	}
	svc := server.NewServerWithAuthorizer(cfg, exchanger, auth, testLogger())

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	extprocv3.RegisterExternalProcessorServer(grpcSrv, svc)
	go func() { _ = grpcSrv.Serve(listener) }()

	conn, err := grpc.NewClient(listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	cleanup := func() {
		_ = conn.Close()
		grpcSrv.GracefulStop()
	}
	return extprocv3.NewExternalProcessorClient(conn), cleanup
}

func startTestServerWithAuthorizerConfig(t *testing.T, cfg *extprocconfig.Config, exchanger server.Exchanger, auth authorization.Authorizer) (extprocv3.ExternalProcessorClient, func()) {
	t.Helper()

	if !cfg.Authorization.Enabled {
		cfg.Authorization.Enabled = true
	}
	if cfg.Authorization.Policy.Package == "" {
		cfg.Authorization.Policy.Package = "aib.extproc.authz"
	}
	if cfg.Authorization.Policy.Decision == "" {
		cfg.Authorization.Policy.Decision = "result"
	}
	if cfg.Authorization.DefaultDecision == "" {
		cfg.Authorization.DefaultDecision = "deny"
	}
	if cfg.Authorization.EvaluationTimeout == 0 {
		cfg.Authorization.EvaluationTimeout = 5 * time.Second
	}
	if cfg.Authorization.MaxBodySize == 0 {
		cfg.Authorization.MaxBodySize = 1048576
	}
	svc := server.NewServerWithAuthorizer(cfg, exchanger, auth, testLogger())

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	extprocv3.RegisterExternalProcessorServer(grpcSrv, svc)
	go func() { _ = grpcSrv.Serve(listener) }()

	conn, err := grpc.NewClient(listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	cleanup := func() {
		_ = conn.Close()
		grpcSrv.GracefulStop()
	}
	return extprocv3.NewExternalProcessorClient(conn), cleanup
}

// sendHeadersThenBody sends RequestHeaders (with agentgateway metadata) followed by
// RequestBody on the same stream, returning the response to the body phase.
func sendHeadersThenBody(t *testing.T, client extprocv3.ExternalProcessorClient, headers map[string]string, body []byte) (*extprocv3.ProcessingResponse, *extprocv3.ProcessingResponse) {
	t.Helper()

	stream, err := client.Process(context.Background())
	require.NoError(t, err)

	headerList := make([]*corev3.HeaderValue, 0, len(headers))
	for k, v := range headers {
		headerList = append(headerList, &corev3.HeaderValue{Key: k, RawValue: []byte(v)})
	}

	// Build MetadataContext with agentgateway protocol = "mcp"
	protocolStruct, err := structpb.NewStruct(map[string]any{"protocol": "mcp"})
	require.NoError(t, err)
	metadata := &corev3.Metadata{
		FilterMetadata: map[string]*structpb.Struct{
			"agentgateway": protocolStruct,
		},
	}

	err = stream.Send(&extprocv3.ProcessingRequest{
		MetadataContext: metadata,
		Request: &extprocv3.ProcessingRequest_RequestHeaders{
			RequestHeaders: &extprocv3.HttpHeaders{
				Headers: &corev3.HeaderMap{Headers: headerList},
			},
		},
	})
	require.NoError(t, err)

	headersResp, err := stream.Recv()
	require.NoError(t, err)

	// If headers response is an ImmediateResponse, stop here
	if _, isImm := headersResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse); isImm {
		_ = stream.CloseSend()
		return headersResp, nil
	}

	// Send body
	err = stream.Send(&extprocv3.ProcessingRequest{
		Request: &extprocv3.ProcessingRequest_RequestBody{
			RequestBody: &extprocv3.HttpBody{
				Body:        body,
				EndOfStream: true,
			},
		},
	})
	require.NoError(t, err)
	_ = stream.CloseSend()

	bodyResp, err := stream.Recv()
	require.NoError(t, err)

	return headersResp, bodyResp
}

// sendHeadersThenBodyWithProtocol is like sendHeadersThenBody but uses the given protocol
// in the agentgateway MetadataContext instead of the default "mcp".
func sendHeadersThenBodyWithProtocol(t *testing.T, client extprocv3.ExternalProcessorClient, headers map[string]string, body []byte, protocol string) (*extprocv3.ProcessingResponse, *extprocv3.ProcessingResponse) {
	t.Helper()

	stream, err := client.Process(context.Background())
	require.NoError(t, err)

	headerList := make([]*corev3.HeaderValue, 0, len(headers))
	for k, v := range headers {
		headerList = append(headerList, &corev3.HeaderValue{Key: k, RawValue: []byte(v)})
	}

	protocolStruct, err := structpb.NewStruct(map[string]any{"protocol": protocol})
	require.NoError(t, err)
	metadata := &corev3.Metadata{
		FilterMetadata: map[string]*structpb.Struct{
			"agentgateway": protocolStruct,
		},
	}

	err = stream.Send(&extprocv3.ProcessingRequest{
		MetadataContext: metadata,
		Request: &extprocv3.ProcessingRequest_RequestHeaders{
			RequestHeaders: &extprocv3.HttpHeaders{
				Headers: &corev3.HeaderMap{Headers: headerList},
			},
		},
	})
	require.NoError(t, err)

	headersResp, err := stream.Recv()
	require.NoError(t, err)

	if _, isImm := headersResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse); isImm {
		_ = stream.CloseSend()
		return headersResp, nil
	}

	err = stream.Send(&extprocv3.ProcessingRequest{
		Request: &extprocv3.ProcessingRequest_RequestBody{
			RequestBody: &extprocv3.HttpBody{Body: body, EndOfStream: true},
		},
	})
	require.NoError(t, err)
	_ = stream.CloseSend()

	bodyResp, err := stream.Recv()
	require.NoError(t, err)

	return headersResp, bodyResp
}

// Spec: FR-009 — No Bearer token with OPA enabled → pass through in headers phase
func TestServer_OPA_NoBearerToken_PassThrough(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			t.Fatal("Evaluate must not be called when no Bearer token is present")
			return nil, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			t.Fatal("Exchange must not be called when no Bearer token is present")
			return server.ExchangeResult{}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	headersResp, _ := sendHeadersThenBody(t, client, map[string]string{
		":path":   "http://mcp-server:9003/mcp",
		":method": "POST",
	}, nil)

	_, ok := headersResp.Response.(*extprocv3.ProcessingResponse_RequestHeaders)
	assert.True(t, ok, "no-Bearer-token with OPA enabled must pass through at headers phase")
}

// Spec: OPA disabled (nil authorizer) → direct token exchange in headers phase (no body buffering)
func TestServer_OPA_Disabled_DirectExchange(t *testing.T) {
	const exchangedToken = "direct-exchanged-token"
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: exchangedToken}, nil
		},
	}
	// NewServer (no authorizer) = OPA disabled
	cfg := testConfig()
	svc := server.NewServer(cfg, exchanger, testLogger())

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcSrv := grpc.NewServer()
	extprocv3.RegisterExternalProcessorServer(grpcSrv, svc)
	go func() { _ = grpcSrv.Serve(listener) }()
	conn, err := grpc.NewClient(listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() { _ = conn.Close(); grpcSrv.GracefulStop() }()
	client := extprocv3.NewExternalProcessorClient(conn)

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://mcp-server:9003/mcp",
		"authorization": "Bearer incoming-token",
	})
	require.NoError(t, err)

	headersResp, ok := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders)
	require.True(t, ok, "OPA disabled: must get RequestHeaders response (not body buffering)")
	require.NotNil(t, headersResp.RequestHeaders.Response)
	assert.Nil(t, resp.ModeOverride,
		"OPA disabled: headers response must not request body buffering")

	var authValue string
	for _, h := range headersResp.RequestHeaders.Response.HeaderMutation.SetHeaders {
		if h.Header.Key == "authorization" {
			authValue = string(h.Header.RawValue)
		}
	}
	assert.Equal(t, "Bearer "+exchangedToken, authValue)
}

// Spec: OPA enabled + allow → token exchange completes, Authorization header replaced
func TestServer_OPA_Enabled_Allow_ExchangeCompletes(t *testing.T) {
	const exchangedToken = "opa-allowed-token"
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: exchangedToken}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	_, bodyResp := sendHeadersThenBody(t, client, map[string]string{
		":method":       "POST",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer incoming-token",
	}, []byte(`{"jsonrpc":"2.0","method":"tools/call","params":{"name":"list_files"}}`))

	require.NotNil(t, bodyResp, "expected body response after OPA allow")
	// Must NOT be an ImmediateResponse (deny)
	_, isImmediate := bodyResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	assert.False(t, isImmediate, "OPA allow must not produce ImmediateResponse")
}

// Spec: OPA enabled + deny → 403 ImmediateResponse with JSON access_denied body
// Note: token exchange now occurs eagerly in the RequestHeaders phase (before OPA evaluation)
// to support proxies that only apply header mutations from HeadersResponse. When OPA denies,
// the 403 ImmediateResponse from the body phase prevents the already-mutated header from
// reaching the backend.
func TestServer_OPA_Enabled_Deny_Returns403(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{
				Action:  "deny",
				Reasons: []string{"tool is destructive"},
			}, nil
		},
	}
	exchanger := &mockExchanger{
		// Exchange IS called in the headers phase (eager exchange); the subsequent OPA
		// denial in the body phase returns 403 and prevents the token from reaching the backend.
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "pre-exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	_, bodyResp := sendHeadersThenBody(t, client, map[string]string{
		":method":       "POST",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer incoming-token",
	}, []byte(`{"jsonrpc":"2.0","method":"tools/call","params":{"name":"delete_file"}}`))

	require.NotNil(t, bodyResp, "expected response in body phase after OPA deny")
	immResp, ok := bodyResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "OPA deny must produce ImmediateResponse")
	assert.Equal(t, int32(403), int32(immResp.ImmediateResponse.Status.Code))
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "access_denied")
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "tool is destructive")
}

// Spec: OPA enabled + header-only request (end_of_stream=true in headers phase) + allow →
// OPA is evaluated with type="mcp_headers_only" and token exchange completes in the headers
// phase. The original bearer token is NOT forwarded unchanged — Authorization header is replaced.
func TestServer_OPA_HeadersOnly_EndOfStream_Allow_ExchangeCompletes(t *testing.T) {
	var capturedInput authorization.OPAInput
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, input authorization.OPAInput) (*authorization.OPADecision, error) {
			capturedInput = input
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	resp, err := sendRequestHeadersWithProtocolEOS(t, client, map[string]string{
		":method":       "GET",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer incoming-token",
	}, "mcp")
	require.NoError(t, err)

	headersResp, ok := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders)
	require.True(t, ok, "OPA allow on header-only request must return RequestHeaders response with token mutation")
	require.NotNil(t, headersResp.RequestHeaders)
	var authHeader string
	for _, h := range headersResp.RequestHeaders.Response.HeaderMutation.SetHeaders {
		if h.Header.Key == "authorization" {
			authHeader = string(h.Header.RawValue)
		}
	}
	assert.Equal(t, "Bearer exchanged-token", authHeader)
	require.NotNil(t, capturedInput, "OPA authorizer must be called for header-only request")
	assert.Equal(t, "mcp_headers_only", capturedInput["type"],
		"MCP header-only request must produce type=mcp_headers_only in OPA input, not type=unknown")
}

// Spec: OPA enabled + header-only request (end_of_stream=true) + deny →
// OPA is evaluated and the request is rejected with 403.
func TestServer_OPA_HeadersOnly_EndOfStream_Deny_Returns403(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "deny", Reasons: []string{"not authorized"}}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			t.Fatal("Exchange must not be called when OPA denies")
			return server.ExchangeResult{}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	resp, err := sendRequestHeadersWithProtocolEOS(t, client, map[string]string{
		":method":       "GET",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer incoming-token",
	}, "mcp")
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "OPA deny on header-only request must return ImmediateResponse")
	assert.Equal(t, int32(403), int32(immResp.ImmediateResponse.Status.Code))
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "access_denied")
}

// Spec: OPA allow + header-only + broker returns 401 + error_uri →
// URLElicitationRequiredError must be returned (same as body-phase path).
func TestServer_OPA_HeadersOnly_EndOfStream_ReAuth_Returns401WithElicitation(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	reAuthURL := "https://idp.example.com/reauth"
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, &server.BrokerExchangeError{
				StatusCode: 401,
				Code:       "reauth_required",
				ErrorURI:   reAuthURL,
			}
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	resp, err := sendRequestHeadersWithProtocolEOS(t, client, map[string]string{
		":method":       "GET",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer incoming-token",
	}, "mcp")
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "re-auth on header-only MCP must produce ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code),
		"MCP re-auth must return HTTP 200 with JSON-RPC -32042")
	assert.Contains(t, string(immResp.ImmediateResponse.Body), reAuthURL)
}

// Spec: OPA allow + header-only + broker returns 500 + error_uri →
// transient 5xx must NOT return elicitation; returns normal exchange failure (500).
func TestServer_OPA_HeadersOnly_EndOfStream_5xxWithErrorURI_Returns500NotElicitation(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, &server.BrokerExchangeError{
				StatusCode: 500,
				Code:       "server_error",
				ErrorURI:   "https://idp.example.com/reauth",
			}
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	resp, err := sendRequestHeadersWithProtocolEOS(t, client, map[string]string{
		":method":       "GET",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer incoming-token",
	}, "mcp")
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "5xx with error_uri on header-only must produce ImmediateResponse")
	assert.NotEqual(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code),
		"transient 5xx must not return elicitation (HTTP 200)")
	assert.Equal(t, int32(httpv3.StatusCode_InternalServerError), int32(immResp.ImmediateResponse.Status.Code),
		"transient 5xx with error_uri on header-only must return 500")
}

// Spec: Only GET is a valid MCP header-only transport (SSE streams, upgrade handshakes).
// Body-bearing methods (POST/PUT/PATCH) without a body return 400 (malformed JSON-RPC).
// POST with no body returns 400. All other non-GET methods return 405 with Allow: GET, POST.
func TestServer_OPA_HeadersOnly_MCP_InvalidMethods(t *testing.T) {
	type testCase struct {
		method    string
		wantCode  httpv3.StatusCode
		wantAllow bool
	}
	cases := []testCase{
		// POST without a body is malformed JSON-RPC → 400
		{"POST", httpv3.StatusCode_BadRequest, false},
		// Unsupported methods → 405 with Allow: GET, POST
		{"PUT", httpv3.StatusCode_MethodNotAllowed, true},
		{"PATCH", httpv3.StatusCode_MethodNotAllowed, true},
		{"DELETE", httpv3.StatusCode_MethodNotAllowed, true},
		{"HEAD", httpv3.StatusCode_MethodNotAllowed, true},
		{"OPTIONS", httpv3.StatusCode_MethodNotAllowed, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.method, func(t *testing.T) {
			auth := &mockAuthorizer{
				evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
					return &authorization.OPADecision{Action: "allow"}, nil
				},
			}
			exchanger := &mockExchanger{
				exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
					return server.ExchangeResult{Token: "exchanged-token"}, nil
				},
			}
			client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
			defer cleanup()

			resp, err := sendRequestHeadersWithProtocolEOS(t, client, map[string]string{
				":method":       tc.method,
				":path":         "http://mcp-server:9003/mcp",
				":authority":    "mcp-server:9003",
				":scheme":       "http",
				"authorization": "Bearer incoming-token",
			}, "mcp")
			require.NoError(t, err)

			immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
			require.True(t, ok, "%s with no body must produce an ImmediateResponse", tc.method)
			assert.Equal(t, int32(tc.wantCode), int32(immResp.ImmediateResponse.Status.Code),
				"%s: unexpected status code", tc.method)
			if tc.wantAllow {
				var allowHeader string
				for _, h := range immResp.ImmediateResponse.Headers.GetSetHeaders() {
					if h.Header.Key == "allow" {
						allowHeader = string(h.Header.RawValue)
					}
				}
				assert.Equal(t, "GET, POST", allowHeader, "%s 405 response must include Allow: GET, POST", tc.method)
			}
		})
	}
}

// Spec: GET is the valid MCP header-only method for SSE streams. An EOS GET must
// succeed through the full OPA+exchange path and return mcp_headers_only input type.
func TestServer_OPA_HeadersOnly_MCP_GET_Succeeds(t *testing.T) {
	var capturedMethod string
	var hasRequestWrapper bool
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, input authorization.OPAInput) (*authorization.OPADecision, error) {
			// Verify input.attributes.request.http.method (envoy-plugin path)
			if attrs, ok := input["attributes"].(map[string]any); ok {
				if req, ok := attrs["request"].(map[string]any); ok {
					if http, ok := req["http"].(map[string]any); ok {
						capturedMethod, _ = http["method"].(string)
					}
				}
			}
			_, hasRequestWrapper = input["request"]
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	resp, err := sendRequestHeadersWithProtocolEOS(t, client, map[string]string{
		":method":       "GET",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer incoming-token",
	}, "mcp")
	require.NoError(t, err)

	headersResp, ok := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders)
	require.True(t, ok, "MCP GET EOS must return RequestHeaders (token mutated)")
	var authHeader string
	for _, h := range headersResp.RequestHeaders.Response.HeaderMutation.SetHeaders {
		if h.Header.Key == "authorization" {
			authHeader = string(h.Header.RawValue)
		}
	}
	assert.Equal(t, "Bearer exchanged-token", authHeader)
	assert.Equal(t, "GET", capturedMethod,
		"mcp_headers_only OPA input must expose GET method via input.attributes.request.http.method")
	assert.False(t, hasRequestWrapper,
		"mcp_headers_only OPA input must not duplicate HTTP request fields at top level")
}

// GET is a header-only MCP transport; a body-bearing GET must be rejected in the body phase.
func TestServer_OPA_MCP_GET_WithBody_Returns405(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	_, bodyResp := sendHeadersThenBodyWithProtocol(t, client, map[string]string{
		":method":       "GET",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer incoming-token",
	}, []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{}}`), "mcp")

	require.NotNil(t, bodyResp)
	immResp, ok := bodyResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "body-bearing MCP GET must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_MethodNotAllowed), int32(immResp.ImmediateResponse.Status.Code))
	var allowHeader string
	for _, h := range immResp.ImmediateResponse.Headers.GetSetHeaders() {
		if h.Header.Key == "allow" {
			allowHeader = string(h.Header.RawValue)
		}
	}
	assert.Equal(t, "GET, POST", allowHeader)
}

// PUT and PATCH are not valid MCP transports even when a body is present — they
// must be rejected with 405 before buffering, not processed as JSON-RPC.
func TestServer_OPA_MCP_InvalidBodyMethods_Return405(t *testing.T) {
	for _, method := range []string{"PUT", "PATCH"} {
		method := method
		t.Run(method, func(t *testing.T) {
			auth := &mockAuthorizer{
				evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
					return &authorization.OPADecision{Action: "allow"}, nil
				},
			}
			exchanger := &mockExchanger{
				exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
					return server.ExchangeResult{Token: "exchanged-token"}, nil
				},
			}
			client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
			defer cleanup()

			headersResp, _ := sendHeadersThenBodyWithProtocol(t, client, map[string]string{
				":method":       method,
				":path":         "http://mcp-server:9003/mcp",
				":authority":    "mcp-server:9003",
				":scheme":       "http",
				"authorization": "Bearer incoming-token",
			}, []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{}}`), "mcp")

			immResp, ok := headersResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
			require.True(t, ok, "%s MCP request must be rejected with ImmediateResponse", method)
			assert.Equal(t, int32(httpv3.StatusCode_MethodNotAllowed), int32(immResp.ImmediateResponse.Status.Code))
			var allowHeader string
			for _, h := range immResp.ImmediateResponse.Headers.GetSetHeaders() {
				if h.Header.Key == "allow" {
					allowHeader = string(h.Header.RawValue)
				}
			}
			assert.Equal(t, "GET, POST", allowHeader)
		})
	}
}

// Spec: Process must handle stream correctly — CloseSend triggers clean completion
func TestServer_Process_StreamHandledCleanly(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "tok"}, nil
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	stream, err := client.Process(context.Background())
	require.NoError(t, err)

	// Close without sending anything — server must handle gracefully
	err = stream.CloseSend()
	require.NoError(t, err)

	_, err = stream.Recv()
	// EOF or a non-error close is acceptable
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			assert.NotEqual(t, codes.Internal, st.Code(),
				"unclean close should not produce Internal error")
		}
	}
}

// ---------------------------------------------------------------------------
// MCP URL Elicitation — BrokerExchangeError with ErrorURI
// ---------------------------------------------------------------------------

// Spec: When broker returns error_uri and protocol is "mcp", headers phase immediately
// returns HTTP 200 with JSON-RPC -32042 URLElicitationRequiredError. id is null because
// the request body has not been read yet; this is correct per JSON-RPC 2.0 §5.
func TestServer_Process_BrokerErrorWithURI_ReturnsElicitationFromHeadersPhase(t *testing.T) {
	reAuthURL := "https://broker.example.com/api/third-party/svc-123/oauth2/authorize"
	const description = "User session has expired. Please re-authenticate."
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, &server.BrokerExchangeError{
				StatusCode:  401,
				Code:        "invalid_grant",
				Description: description,
				ErrorURI:    reAuthURL,
			}
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeadersWithProtocol(t, client, map[string]string{
		":path":         "http://mcp-server:9003/mcp",
		"authorization": "Bearer subject-token",
	}, "mcp")
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "headers phase must return ImmediateResponse for elicitation")
	assert.Equal(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code),
		"elicitation must use HTTP 200 (JSON-RPC errors always travel over HTTP 200)")

	var envelope struct {
		JSONRPC string `json:"jsonrpc"`
		ID      any    `json:"id"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Elicitations []struct {
					Mode          string `json:"mode"`
					ElicitationID string `json:"elicitationId"`
					URL           string `json:"url"`
					Message       string `json:"message"`
				} `json:"elicitations"`
			} `json:"data"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(immResp.ImmediateResponse.Body, &envelope))

	assert.Equal(t, "2.0", envelope.JSONRPC)
	assert.Nil(t, envelope.ID, "id must be JSON null (body not yet available in headers phase)")
	assert.Equal(t, -32042, envelope.Error.Code, "must use JSON-RPC error code -32042")
	assert.Equal(t, description, envelope.Error.Message)
	require.Len(t, envelope.Error.Data.Elicitations, 1)
	assert.Equal(t, "url", envelope.Error.Data.Elicitations[0].Mode)
	assert.Equal(t, reAuthURL, envelope.Error.Data.Elicitations[0].URL)
	assert.NotEmpty(t, envelope.Error.Data.Elicitations[0].ElicitationID, "elicitationId must be set")
}

// Spec: When broker returns error_uri and protocol metadata is absent, non-OPA mode
// returns URLElicitationRequiredError (same as protocol="mcp"). Absent metadata is
// treated as MCP for backward compatibility with token-exchange-only deployments.
func TestServer_Process_BrokerErrorWithURI_AbsentMetadata_ReturnsElicitationFromHeadersPhase(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, &server.BrokerExchangeError{
				StatusCode:  401,
				Code:        "invalid_grant",
				Description: "session expired",
				ErrorURI:    "https://broker.example.com/api/third-party/svc-123/oauth2/authorize",
			}
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://api-server:9003/v1/resource",
		"authorization": "Bearer subject-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "absent metadata must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code),
		"absent metadata is treated as MCP — elicitation uses HTTP 200")
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "-32042",
		"absent metadata must return URLElicitationRequiredError (JSON-RPC -32042), not 503")
}

// Spec: absent metadata with any path returns URLElicitationRequiredError — protocol ""
// is treated as MCP for backward compatibility regardless of the request path.
func TestServer_Process_BrokerErrorWithURI_AbsentMetadataNonMCPPath_ReturnsElicitation(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, &server.BrokerExchangeError{
				StatusCode:  401,
				Code:        "invalid_grant",
				Description: "session expired",
				ErrorURI:    "https://broker.example.com/api/third-party/svc-999/oauth2/authorize",
			}
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://a2a-server:9003/a2a",
		":authority":    "a2a-server:9003",
		"authorization": "Bearer subject-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "absent metadata must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code),
		"absent metadata returns URLElicitationRequiredError (HTTP 200)")
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "-32042",
		"absent metadata must return URLElicitationRequiredError (JSON-RPC -32042)")
}

// Spec: absent metadata with /mcp path returns URLElicitationRequiredError — protocol ""
// is treated as MCP for backward compatibility with direct connections.
func TestServer_Process_BrokerErrorWithURI_AbsentMetadataMCPPath_ReturnsElicitation(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, &server.BrokerExchangeError{
				StatusCode:  401,
				Code:        "invalid_grant",
				Description: "session expired",
				ErrorURI:    "https://broker.example.com/api/third-party/svc-123/oauth2/authorize",
			}
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://mcp-server:9003/mcp",
		"authorization": "Bearer subject-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "absent metadata must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code),
		"absent metadata is treated as MCP — elicitation uses HTTP 200")
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "-32042",
		"absent metadata is treated as MCP — elicitation returns JSON-RPC -32042")
}

// Spec: When broker returns error_uri and protocol is explicitly non-MCP (e.g. "a2a"),
// headers phase must return 503 — URLElicitationRequiredError is MCP-specific.
func TestServer_Process_BrokerErrorWithURI_NonMCP_Returns503FromHeadersPhase(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, &server.BrokerExchangeError{
				StatusCode:  401,
				Code:        "invalid_grant",
				Description: "session expired",
				ErrorURI:    "https://idp.example.com/reauth",
			}
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeadersWithProtocol(t, client, map[string]string{
		":path":         "http://a2a-server:9003/a2a",
		"authorization": "Bearer subject-token",
	}, "a2a")
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "non-MCP re-auth must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_ServiceUnavailable), int32(immResp.ImmediateResponse.Status.Code),
		"non-MCP re-auth in headers phase must return 503")
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "service_unavailable")
}

// Spec: A transient 5xx broker error carrying error_uri must NOT trigger re-auth
// elicitation — it must be treated as a normal exchange failure (503), even when
// the exchange cache is empty (cache miss). Transient errors must never surface
// a spurious re-auth URL.
func TestServer_Process_5xxBrokerErrorWithURI_Returns503NotElicitation(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, &server.BrokerExchangeError{
				StatusCode: 500,
				Code:       "server_error",
				ErrorURI:   "https://idp.example.com/reauth",
			}
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeadersWithProtocol(t, client, map[string]string{
		":path":         "http://mcp-server:9003/mcp",
		"authorization": "Bearer subject-token",
	}, "mcp")
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "5xx with error_uri must produce an ImmediateResponse")
	assert.NotEqual(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code),
		"transient 5xx with error_uri must not return elicitation (HTTP 200)")
	assert.Equal(t, int32(httpv3.StatusCode_InternalServerError), int32(immResp.ImmediateResponse.Status.Code),
		"transient 5xx with error_uri must return 500, not elicitation")
}

// ---------------------------------------------------------------------------
// FR-003: OPA mode — absent protocol metadata
// ---------------------------------------------------------------------------

// Spec: FR-003 — when OPA is enabled and agentgateway protocol metadata is absent,
// the request MUST be rejected with 403 without calling the exchanger or authorizer.
func TestServer_OPA_AbsentMetadata_Returns403(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			t.Fatal("Evaluate must not be called when protocol metadata is absent")
			return nil, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			t.Fatal("Exchange must not be called when protocol metadata is absent")
			return server.ExchangeResult{}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	stream, err := client.Process(context.Background())
	require.NoError(t, err)

	// Send RequestHeaders without MetadataContext.
	err = stream.Send(&extprocv3.ProcessingRequest{
		// MetadataContext intentionally omitted.
		Request: &extprocv3.ProcessingRequest_RequestHeaders{
			RequestHeaders: &extprocv3.HttpHeaders{
				Headers: &corev3.HeaderMap{Headers: []*corev3.HeaderValue{
					{Key: ":path", RawValue: []byte("/mcp")},
					{Key: ":scheme", RawValue: []byte("https")},
					{Key: ":authority", RawValue: []byte("mcp-server:9003")},
					{Key: "authorization", RawValue: []byte("Bearer subject-token")},
				}},
			},
		},
	})
	require.NoError(t, err)

	resp, err := stream.Recv()
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "absent metadata must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_Forbidden), int32(immResp.ImmediateResponse.Status.Code))
}

// Spec: Exchange fails with re-auth → elicitation is returned in the headers-phase response.
// Token exchange now occurs eagerly in the headers phase (before the body is seen), so the
// JSON-RPC id cannot be extracted from the request body; the elicitation carries id=null.
func TestServer_OPA_BodyPhaseReauth_ReturnsElicitationInHeadersPhase(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	reAuthErr := &server.BrokerExchangeError{
		StatusCode:  401,
		Code:        "reauth_required",
		Description: "session expired",
		ErrorURI:    "https://idp.example.com/reauth",
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, reAuthErr
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	headersResp, bodyResp := sendHeadersThenBody(t, client, map[string]string{
		":method":       "POST",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer subject-token",
	}, []byte(`{"jsonrpc":"2.0","method":"tools/call","id":42,"params":{"name":"list_files"}}`))

	// Exchange failed in headers phase → elicitation is in headersResp, no body phase.
	assert.Nil(t, bodyResp, "re-auth in headers phase must not produce a body-phase response")
	immResp, ok := headersResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "re-auth must produce an ImmediateResponse in the headers phase")
	assert.Equal(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code))
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "idp.example.com/reauth")

	var rpcErr map[string]any
	require.NoError(t, json.Unmarshal(immResp.ImmediateResponse.Body, &rpcErr))
	assert.Nil(t, rpcErr["id"], "headers-phase re-auth elicitation carries id=null (body not yet seen)")
}

// Spec: An MCP body-phase request with an empty body must be rejected with 403 (fail-closed).
// The empty-body path in buildMCPInput is reserved for EOS header-only requests via
// BuildOPAInputHeadersOnly; a zero-length body reaching the body phase is malformed.
// Token exchange occurs eagerly in the headers phase (before the body is seen), so
// exchange IS called. The 403 is returned from the body phase after OPA rejects the
// empty body.
func TestServer_OPA_BodyPhase_EmptyMCPBody_Denied(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			t.Fatal("OPA authorizer must not be called for malformed (empty-body) MCP requests")
			return nil, nil
		},
	}
	exchanger := &mockExchanger{
		// Exchange IS called eagerly in the headers phase before the body arrives.
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "pre-exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	_, bodyResp := sendHeadersThenBody(t, client, map[string]string{
		":method":       "POST",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer subject-token",
	}, []byte{})

	require.NotNil(t, bodyResp, "empty-body MCP request must produce a body-phase response")
	immResp, ok := bodyResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "empty-body MCP request must produce ImmediateResponse")
	assert.Equal(t, int32(403), int32(immResp.ImmediateResponse.Status.Code),
		"empty-body MCP request must be denied with 403")
}

// Spec: Bodies larger than authorization.max_body_size must be rejected before OPA evaluation.
func TestServer_OPA_BodyPhase_RequestTooLarge_DeniesWithoutAuthorizerCall(t *testing.T) {
	cfg := testConfig()
	cfg.Authorization.Enabled = true
	cfg.Authorization.MaxBodySize = 10

	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			t.Fatal("OPA authorizer must not be called when the body exceeds max_body_size")
			return nil, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "pre-exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizerConfig(t, cfg, exchanger, auth)
	defer cleanup()

	_, bodyResp := sendHeadersThenBody(t, client, map[string]string{
		":method":       "POST",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer subject-token",
	}, []byte("01234567890"))

	require.NotNil(t, bodyResp, "oversized request must produce a body-phase response")
	immResp, ok := bodyResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "oversized request must produce ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_Forbidden), int32(immResp.ImmediateResponse.Status.Code))
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "request_too_large")
}

// Spec: A batch where every element is allowed must evaluate each element and echo the full body.
func TestServer_OPA_BatchBodyPhase_Allow_EchoesBody(t *testing.T) {
	var seenToolNames []string
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, input authorization.OPAInput) (*authorization.OPADecision, error) {
			mcp, ok := input["mcp"].(*authorization.MCPInput)
			require.True(t, ok, "batch element must expose parsed MCP input")
			seenToolNames = append(seenToolNames, mcp.ToolName)
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "batch-token"}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	batch := []byte(`[{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"list_files","arguments":{}}},{"jsonrpc":"2.0","method":"tools/call","id":2,"params":{"name":"read_config","arguments":{}}}]`)
	_, bodyResp := sendHeadersThenBody(t, client, map[string]string{
		":method":       "POST",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer subject-token",
	}, batch)

	require.NotNil(t, bodyResp)
	requestBodyResp, ok := bodyResp.Response.(*extprocv3.ProcessingResponse_RequestBody)
	require.True(t, ok, "allowed batch must echo a RequestBody response")
	streamed, ok := requestBodyResp.RequestBody.Response.BodyMutation.Mutation.(*extprocv3.BodyMutation_StreamedResponse)
	require.True(t, ok, "allowed batch must use streamed body echo")
	assert.Equal(t, batch, streamed.StreamedResponse.Body)
	assert.Equal(t, []string{"list_files", "read_config"}, seenToolNames)
}

// Spec: A denied batch must aggregate reasons from every denying element into one 403.
func TestServer_OPA_BatchBodyPhase_Deny_AggregatesReasons(t *testing.T) {
	var evaluateCalls int
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, input authorization.OPAInput) (*authorization.OPADecision, error) {
			evaluateCalls++
			mcp, ok := input["mcp"].(*authorization.MCPInput)
			require.True(t, ok, "batch element must expose parsed MCP input")
			return &authorization.OPADecision{Action: "deny", Reasons: []string{"denied tool: " + mcp.ToolName}}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "batch-token"}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	batch := []byte(`[{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"delete_repo","arguments":{}}},{"jsonrpc":"2.0","method":"tools/call","id":2,"params":{"name":"drop_db","arguments":{}}}]`)
	_, bodyResp := sendHeadersThenBody(t, client, map[string]string{
		":method":       "POST",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer subject-token",
	}, batch)

	require.NotNil(t, bodyResp)
	immResp, ok := bodyResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "denied batch must produce ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_Forbidden), int32(immResp.ImmediateResponse.Status.Code))
	assert.Equal(t, 2, evaluateCalls)
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "denied tool: delete_repo")
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "denied tool: drop_db")
}

// Spec: Exchange fails with re-auth for a batch request → elicitation is returned in the
// headers-phase response. Token exchange occurs eagerly in the headers phase (before the
// batch body is seen), so the elicitation carries id=null regardless of batch element ids.
func TestServer_OPA_BatchBodyPhaseReauth_ReturnsElicitationInHeadersPhase(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	reAuthErr := &server.BrokerExchangeError{
		StatusCode:  401,
		Code:        "reauth_required",
		Description: "session expired",
		ErrorURI:    "https://idp.example.com/reauth",
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, reAuthErr
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	batch := []byte(`[{"jsonrpc":"2.0","method":"tools/call","id":7,"params":{"name":"a"}},{"jsonrpc":"2.0","method":"tools/call","id":8,"params":{"name":"b"}}]`)
	headersResp, bodyResp := sendHeadersThenBody(t, client, map[string]string{
		":method":       "POST",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer subject-token",
	}, batch)

	// Exchange failed in headers phase → elicitation is in headersResp, no body phase.
	assert.Nil(t, bodyResp, "re-auth in headers phase must not produce a body-phase response")
	immResp, ok := headersResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "batch re-auth must produce an ImmediateResponse in the headers phase")
	assert.Equal(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code))
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "idp.example.com/reauth")

	var rpcErr map[string]any
	require.NoError(t, json.Unmarshal(immResp.ImmediateResponse.Body, &rpcErr))
	assert.Nil(t, rpcErr["id"], "headers-phase re-auth elicitation carries id=null (batch not yet seen)")
}

// Spec: Exchange fails with re-auth for a batch starting with a notification → elicitation
// is returned in the headers-phase response with id=null. Token exchange occurs eagerly in
// the headers phase, so the batch body (and its element ids) is never seen.
func TestServer_OPA_BatchBodyPhaseReauth_NotificationFirstUsesNextID(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	reAuthErr := &server.BrokerExchangeError{
		StatusCode:  401,
		Code:        "reauth_required",
		Description: "session expired",
		ErrorURI:    "https://idp.example.com/reauth",
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, reAuthErr
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	// First element is a notification (no id field); second carries id=42.
	batch := []byte(`[{"jsonrpc":"2.0","method":"notifications/progress"},{"jsonrpc":"2.0","method":"tools/call","id":42,"params":{"name":"a"}}]`)
	headersResp, bodyResp := sendHeadersThenBody(t, client, map[string]string{
		":method":       "POST",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer subject-token",
	}, batch)

	// Exchange failed in headers phase → elicitation is in headersResp, no body phase.
	assert.Nil(t, bodyResp, "re-auth in headers phase must not produce a body-phase response")
	immResp, ok := headersResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "batch re-auth must produce an ImmediateResponse in the headers phase")
	assert.Equal(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code))

	var rpcErr map[string]any
	require.NoError(t, json.Unmarshal(immResp.ImmediateResponse.Body, &rpcErr))
	assert.Nil(t, rpcErr["id"], "headers-phase re-auth elicitation carries id=null (batch not yet seen)")
}

// Spec: A transient 5xx broker error with error_uri on an MCP batch request must NOT
// trigger re-auth elicitation — it must return a normal exchange failure (500).
// Token exchange occurs eagerly in the headers phase, so the 500 is returned in the
// headers-phase response (not the body phase).
func TestServer_OPA_BatchBodyPhase_5xxWithErrorURI_Returns500NotElicitation(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, &server.BrokerExchangeError{
				StatusCode: 500,
				Code:       "server_error",
				ErrorURI:   "https://idp.example.com/reauth",
			}
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	batch := []byte(`[{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"a"}}]`)
	headersResp, bodyResp := sendHeadersThenBody(t, client, map[string]string{
		":method":       "POST",
		":path":         "http://mcp-server:9003/mcp",
		":authority":    "mcp-server:9003",
		":scheme":       "http",
		"authorization": "Bearer subject-token",
	}, batch)

	// Exchange failed in headers phase → error is in headersResp, no body phase.
	assert.Nil(t, bodyResp, "5xx exchange error in headers phase must not produce a body-phase response")
	immResp, ok := headersResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "5xx with error_uri on batch must produce an ImmediateResponse in the headers phase")
	assert.NotEqual(t, int32(httpv3.StatusCode_OK), int32(immResp.ImmediateResponse.Status.Code),
		"transient 5xx with error_uri must not return elicitation (HTTP 200)")
	assert.Equal(t, int32(httpv3.StatusCode_InternalServerError), int32(immResp.ImmediateResponse.Status.Code),
		"transient 5xx with error_uri must return 500, not elicitation")
}

// Spec: re-auth returns elicitation in the headers phase with id=null. Token exchange
// occurs eagerly before the body is seen, so the request body (including its id field)
// is never inspected; id is always null in headers-phase elicitation responses.
func TestServer_OPA_BodyPhaseReauth_InvalidIDTypesFallBackToNull(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	reAuthErr := &server.BrokerExchangeError{
		StatusCode:  401,
		Code:        "reauth_required",
		Description: "session expired",
		ErrorURI:    "https://idp.example.com/reauth",
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, reAuthErr
		},
	}

	for _, tc := range []struct {
		name string
		body []byte
	}{
		{"object id", []byte(`{"jsonrpc":"2.0","method":"tools/call","id":{},"params":{"name":"a"}}`)},
		{"array id", []byte(`{"jsonrpc":"2.0","method":"tools/call","id":[],"params":{"name":"a"}}`)},
		{"boolean true id", []byte(`{"jsonrpc":"2.0","method":"tools/call","id":true,"params":{"name":"a"}}`)},
		{"boolean false id", []byte(`{"jsonrpc":"2.0","method":"tools/call","id":false,"params":{"name":"a"}}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
			defer cleanup()

			headersResp, bodyResp := sendHeadersThenBody(t, client, map[string]string{
				":method":       "POST",
				":path":         "http://mcp-server:9003/mcp",
				":authority":    "mcp-server:9003",
				":scheme":       "http",
				"authorization": "Bearer subject-token",
			}, tc.body)

			// Exchange failed in headers phase → elicitation is in headersResp, no body phase.
			assert.Nil(t, bodyResp)
			immResp, ok := headersResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
			require.True(t, ok)

			var rpcErr map[string]any
			require.NoError(t, json.Unmarshal(immResp.ImmediateResponse.Body, &rpcErr))
			assert.Nil(t, rpcErr["id"], "headers-phase re-auth elicitation carries id=null")
		})
	}
}

// Spec: When OPA allows a non-MCP request and Exchange returns error_uri,
// the response must NOT be a URLElicitationRequiredError (MCP-specific).
// It should fall through to the generic token-exchange error path.
func TestServer_OPA_NonMCPBodyPhaseReauth_DoesNotReturnElicitation(t *testing.T) {
	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	reAuthErr := &server.BrokerExchangeError{
		StatusCode:  401,
		Code:        "reauth_required",
		Description: "session expired",
		ErrorURI:    "https://idp.example.com/reauth",
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, reAuthErr
		},
	}
	client, cleanup := startTestServerWithAuthorizer(t, exchanger, auth)
	defer cleanup()

	// Send as "a2a" protocol (non-MCP) — Exchange returns error_uri in headers phase.
	headersResp, bodyResp := sendHeadersThenBodyWithProtocol(t, client, map[string]string{
		":path":         "http://a2a-server:9003/a2a",
		":authority":    "a2a-server:9003",
		":scheme":       "http",
		"authorization": "Bearer subject-token",
	}, []byte(`{"some":"a2a-body"}`), "a2a")

	// Exchange failed in headers phase → error is in headersResp, no body phase.
	assert.Nil(t, bodyResp, "non-MCP re-auth in headers phase must not produce a body-phase response")
	immResp, ok := headersResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "non-MCP re-auth must produce an ImmediateResponse in the headers phase")
	assert.Equal(t, int32(httpv3.StatusCode_ServiceUnavailable), int32(immResp.ImmediateResponse.Status.Code),
		"non-MCP re-auth must return 503")
	assert.Contains(t, string(immResp.ImmediateResponse.Body), "service_unavailable",
		"non-MCP re-auth body must contain service_unavailable")
}

// Spec: BrokerExchangeError without ErrorURI still returns 500 (no elicitation).
func TestServer_Process_BrokerErrorWithoutURI_Returns500(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, &server.BrokerExchangeError{
				StatusCode:  403,
				Code:        "access_denied",
				Description: "agent does not have a grant",
				// ErrorURI intentionally empty
			}
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://mcp-server:9003/mcp",
		"authorization": "Bearer some-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "broker error without ErrorURI must produce an ImmediateResponse directly")
	assert.Equal(t, int32(httpv3.StatusCode_InternalServerError),
		int32(immResp.ImmediateResponse.Status.Code))
}

// ---------------------------------------------------------------------------
// T021: Span creation in processRequestHeaders
// ---------------------------------------------------------------------------

// Spec: US1 S2 — When traces are enabled, processRequestHeaders creates a span
// with resource.uri and outcome attributes set correctly
func TestServer_ProcessRequestHeaders_CreatesSpanWithAttributes(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prevTP) })

	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithConfig(t, testConfigWithTelemetry(), exchanger)
	defer cleanup()

	_, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "/api/resource",
		":scheme":       "https",
		":authority":    "example.com",
		"authorization": "Bearer test-token",
	})
	require.NoError(t, err)

	tp.ForceFlush(context.Background()) //nolint:errcheck
	spans := spanRecorder.Ended()
	require.NotEmpty(t, spans, "at least one span must be recorded")

	var found bool
	for _, s := range spans {
		if s.Name() == "extproc.token_exchange" {
			found = true
			attrs := attributeMap(s.Attributes())
			assert.Equal(t, "https://example.com/api/resource", attrs["resource.uri"], "resource.uri must be set")
			assert.Equal(t, "success", attrs["outcome"], "outcome must be 'success'")
			break
		}
	}
	assert.True(t, found, "span 'extproc.token_exchange' must exist")
}

// Spec: US1 S2 — Span outcome attribute must be set based on exchange result
func TestServer_ProcessRequestHeaders_SpanOutcomeOnSuccess(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prevTP) })

	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "success-token"}, nil
		},
	}
	client, cleanup := startTestServerWithConfig(t, testConfigWithTelemetry(), exchanger)
	defer cleanup()

	_, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://example.com/api",
		"authorization": "Bearer test-token",
	})
	require.NoError(t, err)

	tp.ForceFlush(context.Background()) //nolint:errcheck
	spans := spanRecorder.Ended()

	var outcomeAttr string
	for _, s := range spans {
		if s.Name() == "extproc.token_exchange" {
			outcomeAttr = attributeMap(s.Attributes())["outcome"]
		}
	}
	assert.Equal(t, "success", outcomeAttr, "span outcome must be 'success' on successful exchange")
}

// Spec: US1 S2 — Span outcome attribute must be set on exchange_failure
func TestServer_ProcessRequestHeaders_SpanOutcomeOnFailure(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prevTP) })

	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{}, errors.New("token exchange failed")
		},
	}
	client, cleanup := startTestServerWithConfig(t, testConfigWithTelemetry(), exchanger)
	defer cleanup()

	_, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://example.com/api",
		"authorization": "Bearer test-token",
	})
	require.NoError(t, err)

	tp.ForceFlush(context.Background()) //nolint:errcheck
	spans := spanRecorder.Ended()

	var outcomeAttr string
	for _, s := range spans {
		if s.Name() == "extproc.token_exchange" {
			outcomeAttr = attributeMap(s.Attributes())["outcome"]
		}
	}
	assert.Equal(t, "exchange_failure", outcomeAttr, "span outcome must be 'exchange_failure' on error")
}

// Spec: US3 S1 — Metrics are recorded with outcome attribute
func TestServer_ProcessRequestHeaders_MetricsRecordOutcome(t *testing.T) {
	metricReader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricReader))
	prevMP := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	t.Cleanup(func() { otel.SetMeterProvider(prevMP) })

	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			return server.ExchangeResult{Token: "exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithConfig(t, testConfigWithTelemetry(), exchanger)
	defer cleanup()

	_, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://example.com/api",
		"authorization": "Bearer test-token",
	})
	require.NoError(t, err)

	var rm metricdata.ResourceMetrics
	require.NoError(t, metricReader.Collect(context.Background(), &rm))

	var foundCounterWithOutcome, foundHistogramWithOutcome bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch m.Name {
			case "extproc.token_exchange.requests":
				if sum, ok := m.Data.(metricdata.Sum[int64]); ok {
					for _, dp := range sum.DataPoints {
						for _, attr := range dp.Attributes.ToSlice() {
							if string(attr.Key) == "outcome" && attr.Value.AsString() == "success" {
								foundCounterWithOutcome = true
							}
						}
					}
				}
			case "extproc.token_exchange.duration":
				if hist, ok := m.Data.(metricdata.Histogram[float64]); ok {
					for _, dp := range hist.DataPoints {
						for _, attr := range dp.Attributes.ToSlice() {
							if string(attr.Key) == "outcome" && attr.Value.AsString() == "success" {
								foundHistogramWithOutcome = true
							}
						}
					}
				}
			}
		}
	}
	assert.True(t, foundCounterWithOutcome, "counter must have outcome=success attribute")
	assert.True(t, foundHistogramWithOutcome, "histogram must have outcome=success attribute")
}

// Regression test: metrics must still be recorded when the gRPC stream context
// is cancelled (client disconnect) because the deferred closure uses
// context.WithoutCancel. Without that guard, the OTel SDK receives a cancelled
// context which could silently drop metric data points.
func TestServer_ProcessRequestHeaders_MetricsRecordedWhenStreamCancelled(t *testing.T) {
	metricReader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricReader))
	prevMP := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	t.Cleanup(func() { otel.SetMeterProvider(prevMP) })

	exchangeStarted := make(chan struct{})
	exchangeContinue := make(chan struct{})

	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			close(exchangeStarted)
			<-exchangeContinue
			return server.ExchangeResult{Token: "exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithConfig(t, testConfigWithTelemetry(), exchanger)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := client.Process(ctx)
	require.NoError(t, err)

	err = stream.Send(&extprocv3.ProcessingRequest{
		Request: &extprocv3.ProcessingRequest_RequestHeaders{
			RequestHeaders: &extprocv3.HttpHeaders{
				Headers: &corev3.HeaderMap{Headers: []*corev3.HeaderValue{
					{Key: ":path", RawValue: []byte("http://example.com/api")},
					{Key: "authorization", RawValue: []byte("Bearer test-token")},
				}},
			},
		},
	})
	require.NoError(t, err)
	_ = stream.CloseSend()

	// Wait for exchange to start, cancel stream context (simulate client disconnect),
	// then let the exchange complete.
	<-exchangeStarted
	cancel()
	close(exchangeContinue)

	// Poll until metrics appear — avoids flaky time.Sleep on slow CI runners.
	var foundCounter, foundHistogram bool
	require.Eventually(t, func() bool {
		var rm metricdata.ResourceMetrics
		if err := metricReader.Collect(context.Background(), &rm); err != nil {
			return false
		}
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				switch m.Name {
				case "extproc.token_exchange.requests":
					if sum, ok := m.Data.(metricdata.Sum[int64]); ok && len(sum.DataPoints) > 0 {
						foundCounter = true
					}
				case "extproc.token_exchange.duration":
					if hist, ok := m.Data.(metricdata.Histogram[float64]); ok && len(hist.DataPoints) > 0 {
						foundHistogram = true
					}
				}
			}
		}
		return foundCounter && foundHistogram
	}, 2*time.Second, 10*time.Millisecond,
		"counter and histogram must be recorded even when stream context is cancelled")
}

func TestServer_OPA_BodyBearing_EmitsTelemetryAndPropagatesTrace(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prevTP) })
	prevProp := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(prevProp) })

	metricReader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricReader))
	prevMP := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	t.Cleanup(func() { otel.SetMeterProvider(prevMP) })

	const traceparentHeader = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	const expectedTraceID = "4bf92f3577b34da6a3ce929d0e0e4736"
	var seenTraceID string

	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(ctx context.Context, _, _ string) (server.ExchangeResult, error) {
			seenTraceID = trace.SpanContextFromContext(ctx).TraceID().String()
			return server.ExchangeResult{Token: "exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizerConfig(t, testConfigWithTelemetry(), exchanger, auth)
	defer cleanup()

	_, bodyResp := sendHeadersThenBodyWithProtocol(t, client, map[string]string{
		":method":       "POST",
		":path":         "https://example.com/api/resource?access_token=secret",
		"authorization": "Bearer test-token",
		"traceparent":   traceparentHeader,
	}, []byte(`{"jsonrpc":"2.0","method":"tools/call","params":{"name":"list_files"}}`), "mcp")
	require.NotNil(t, bodyResp)
	_, isImmediate := bodyResp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	assert.False(t, isImmediate, "allow path should not return ImmediateResponse")

	tp.ForceFlush(context.Background()) //nolint:errcheck
	spans := spanRecorder.Ended()
	require.NotEmpty(t, spans, "OPA exchange path must emit a span")
	assert.Equal(t, expectedTraceID, seenTraceID, "exchange call should receive trace context extracted from request headers")

	var foundSpan bool
	for _, s := range spans {
		if s.Name() == "extproc.token_exchange" {
			foundSpan = true
			attrs := attributeMap(s.Attributes())
			assert.Equal(t, "https://example.com/api/resource", attrs["resource.uri"])
			assert.Equal(t, "success", attrs["outcome"])
		}
	}
	assert.True(t, foundSpan, "OPA exchange path must emit extproc.token_exchange span")

	var rm metricdata.ResourceMetrics
	require.NoError(t, metricReader.Collect(context.Background(), &rm))
	var foundCounterWithOutcome, foundHistogramWithOutcome bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch m.Name {
			case "extproc.token_exchange.requests":
				if sum, ok := m.Data.(metricdata.Sum[int64]); ok {
					for _, dp := range sum.DataPoints {
						for _, attr := range dp.Attributes.ToSlice() {
							if string(attr.Key) == "outcome" && attr.Value.AsString() == "success" {
								foundCounterWithOutcome = true
							}
						}
					}
				}
			case "extproc.token_exchange.duration":
				if hist, ok := m.Data.(metricdata.Histogram[float64]); ok {
					for _, dp := range hist.DataPoints {
						for _, attr := range dp.Attributes.ToSlice() {
							if string(attr.Key) == "outcome" && attr.Value.AsString() == "success" {
								foundHistogramWithOutcome = true
							}
						}
					}
				}
			}
		}
	}
	assert.True(t, foundCounterWithOutcome, "OPA exchange path must record counter with outcome=success")
	assert.True(t, foundHistogramWithOutcome, "OPA exchange path must record histogram with outcome=success")
}

func TestServer_OPA_HeadersOnly_EmitsTelemetryAndPropagatesTrace(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	prevProp := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(prevProp) })
	t.Cleanup(func() { otel.SetTracerProvider(prevTP) })

	metricReader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricReader))
	prevMP := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	t.Cleanup(func() { otel.SetMeterProvider(prevMP) })

	const traceparentHeader = "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01"
	const expectedTraceID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var seenTraceID string

	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "allow"}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(ctx context.Context, _, _ string) (server.ExchangeResult, error) {
			seenTraceID = trace.SpanContextFromContext(ctx).TraceID().String()
			return server.ExchangeResult{Token: "exchanged-token"}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizerConfig(t, testConfigWithTelemetry(), exchanger, auth)
	defer cleanup()

	resp, err := sendRequestHeadersWithProtocolEOS(t, client, map[string]string{
		":method":       "GET",
		":path":         "https://example.com/mcp?client_secret=secret",
		"authorization": "Bearer test-token",
		"traceparent":   traceparentHeader,
	}, "mcp")
	require.NoError(t, err)
	require.NotNil(t, resp)

	tp.ForceFlush(context.Background()) //nolint:errcheck
	spans := spanRecorder.Ended()
	require.NotEmpty(t, spans, "OPA header-only exchange path must emit a span")
	assert.Equal(t, expectedTraceID, seenTraceID, "header-only exchange should receive trace context extracted from request headers")

	var foundSpan bool
	for _, s := range spans {
		if s.Name() == "extproc.token_exchange" {
			foundSpan = true
			attrs := attributeMap(s.Attributes())
			assert.Equal(t, "https://example.com/mcp", attrs["resource.uri"])
			assert.Equal(t, "success", attrs["outcome"])
		}
	}
	assert.True(t, foundSpan, "OPA header-only exchange path must emit extproc.token_exchange span")

	var rm metricdata.ResourceMetrics
	require.NoError(t, metricReader.Collect(context.Background(), &rm))
	var foundCounterWithOutcome, foundHistogramWithOutcome bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch m.Name {
			case "extproc.token_exchange.requests":
				if sum, ok := m.Data.(metricdata.Sum[int64]); ok {
					for _, dp := range sum.DataPoints {
						for _, attr := range dp.Attributes.ToSlice() {
							if string(attr.Key) == "outcome" && attr.Value.AsString() == "success" {
								foundCounterWithOutcome = true
							}
						}
					}
				}
			case "extproc.token_exchange.duration":
				if hist, ok := m.Data.(metricdata.Histogram[float64]); ok {
					for _, dp := range hist.DataPoints {
						for _, attr := range dp.Attributes.ToSlice() {
							if string(attr.Key) == "outcome" && attr.Value.AsString() == "success" {
								foundHistogramWithOutcome = true
							}
						}
					}
				}
			}
		}
	}
	assert.True(t, foundCounterWithOutcome, "OPA header-only path must record counter with outcome=success")
	assert.True(t, foundHistogramWithOutcome, "OPA header-only path must record histogram with outcome=success")
}

func TestServer_OPA_HeadersOnly_Deny_RecordsAuthorizationOutcome(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prevTP) })

	metricReader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricReader))
	prevMP := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	t.Cleanup(func() { otel.SetMeterProvider(prevMP) })

	auth := &mockAuthorizer{
		evaluateFunc: func(_ context.Context, _ authorization.OPAInput) (*authorization.OPADecision, error) {
			return &authorization.OPADecision{Action: "deny", Reasons: []string{"tool denied"}}, nil
		},
	}
	exchanger := &mockExchanger{
		exchangeFunc: func(_ context.Context, _, _ string) (server.ExchangeResult, error) {
			t.Fatal("Exchange must not be called when header-only OPA denies")
			return server.ExchangeResult{}, nil
		},
	}
	client, cleanup := startTestServerWithAuthorizerConfig(t, testConfigWithTelemetry(), exchanger, auth)
	defer cleanup()

	resp, err := sendRequestHeadersWithProtocolEOS(t, client, map[string]string{
		":method":       "GET",
		":path":         "https://example.com/mcp?token=secret",
		"authorization": "Bearer test-token",
	}, "mcp")
	require.NoError(t, err)
	require.NotNil(t, resp)

	tp.ForceFlush(context.Background()) //nolint:errcheck
	spans := spanRecorder.Ended()
	require.NotEmpty(t, spans, "OPA header-only deny path must emit a span")

	var foundSpan bool
	for _, s := range spans {
		if s.Name() == "extproc.token_exchange" {
			foundSpan = true
			attrs := attributeMap(s.Attributes())
			assert.Equal(t, "https://example.com/mcp", attrs["resource.uri"])
			assert.Equal(t, "authorization_denied", attrs["outcome"])
			assert.Equal(t, "access_denied", attrs["error.type"])
		}
	}
	assert.True(t, foundSpan, "OPA header-only deny path must emit extproc.token_exchange span")

	var rm metricdata.ResourceMetrics
	require.NoError(t, metricReader.Collect(context.Background(), &rm))
	var foundCounterWithOutcome, foundHistogramWithOutcome bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch m.Name {
			case "extproc.token_exchange.requests":
				if sum, ok := m.Data.(metricdata.Sum[int64]); ok {
					for _, dp := range sum.DataPoints {
						for _, attr := range dp.Attributes.ToSlice() {
							if string(attr.Key) == "outcome" && attr.Value.AsString() == "authorization_denied" {
								foundCounterWithOutcome = true
							}
						}
					}
				}
			case "extproc.token_exchange.duration":
				if hist, ok := m.Data.(metricdata.Histogram[float64]); ok {
					for _, dp := range hist.DataPoints {
						for _, attr := range dp.Attributes.ToSlice() {
							if string(attr.Key) == "outcome" && attr.Value.AsString() == "authorization_denied" {
								foundHistogramWithOutcome = true
							}
						}
					}
				}
			}
		}
	}
	assert.True(t, foundCounterWithOutcome, "OPA header-only deny path must record counter with outcome=authorization_denied")
	assert.True(t, foundHistogramWithOutcome, "OPA header-only deny path must record histogram with outcome=authorization_denied")
}

// attributeMap converts a slice of key-value attributes to a map for easy assertions.
func attributeMap(attrs []attribute.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, a := range attrs {
		m[string(a.Key)] = a.Value.String()
	}
	return m
}
