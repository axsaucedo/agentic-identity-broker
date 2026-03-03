package server_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"net"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"

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
	}
}

// mockExchanger is a controllable Exchanger for unit tests.
type mockExchanger struct {
	exchangeFunc   func(subjectToken, resourceURI string) (string, error)
	shutdownCalled bool
}

func (m *mockExchanger) Exchange(subjectToken, resourceURI string) (string, error) {
	return m.exchangeFunc(subjectToken, resourceURI)
}

func (m *mockExchanger) Shutdown() {
	m.shutdownCalled = true
}

// startTestServer registers the Server on a random in-process port and returns
// a connected client + cleanup function.
func startTestServer(t *testing.T, exchanger server.Exchanger) (extprocv3.ExternalProcessorClient, func()) {
	t.Helper()

	cfg := testConfig()
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

	stream, err := client.Process(context.Background())
	require.NoError(t, err)

	headerList := make([]*corev3.HeaderValue, 0, len(headers))
	for k, v := range headers {
		headerList = append(headerList, &corev3.HeaderValue{Key: k, RawValue: []byte(v)})
	}

	err = stream.Send(&extprocv3.ProcessingRequest{
		Request: &extprocv3.ProcessingRequest_RequestHeaders{
			RequestHeaders: &extprocv3.HttpHeaders{
				Headers: &corev3.HeaderMap{Headers: headerList},
			},
		},
	})
	if err != nil {
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
		exchangeFunc: func(_, _ string) (string, error) {
			t.Fatal("Exchange should not be called when no Bearer token is present")
			return "", nil
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
		exchangeFunc: func(_, _ string) (string, error) {
			t.Fatal("Exchange should not be called for non-Bearer auth")
			return "", nil
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
		exchangeFunc: func(st, ru string) (string, error) {
			assert.Equal(t, subjectToken, st, "Exchange must receive the incoming Bearer token")
			assert.Equal(t, resourceURI, ru, "Exchange must receive the :path as resource URI")
			return exchangedToken, nil
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
		exchangeFunc: func(_, _ string) (string, error) {
			return "", errors.New("exchange endpoint returned 403")
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
		exchangeFunc: func(_, _ string) (string, error) {
			t.Fatal("Exchange should not be called with empty :path")
			return "", nil
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
		exchangeFunc: func(_, _ string) (string, error) {
			t.Fatal("Exchange should not be called with relative :path")
			return "", nil
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
		exchangeFunc: func(_, _ string) (string, error) {
			t.Fatal("Exchange should not be called with non-http :path")
			return "", nil
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

// SR-003 — Exchanged token details must not appear in error responses (information disclosure)
func TestServer_Process_ErrorResponse_DoesNotLeakTokenDetails(t *testing.T) {
	const internalError = "upstream returned 403: access_denied for client xyz"
	exchanger := &mockExchanger{
		exchangeFunc: func(_, _ string) (string, error) {
			return "", errors.New(internalError)
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
		exchangeFunc: func(_, _ string) (string, error) {
			return "", server.ErrAssertionExpired
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
		exchangeFunc: func(_, _ string) (string, error) { return "", nil },
	}
	svc := server.NewServer(cfg, exchanger, testLogger())

	// Compile-time interface check
	var _ extprocv3.ExternalProcessorServer = svc
}

// Spec: Process must handle stream correctly — CloseSend triggers clean completion
func TestServer_Process_StreamHandledCleanly(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_, _ string) (string, error) { return "tok", nil },
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
