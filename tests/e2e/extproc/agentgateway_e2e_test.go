// Package extproc_test contains an E2E integration test that validates the ExtProc
// Token Exchange service works correctly with the real agentgateway Docker image.
// This test is part of the ExtProc E2E suite (TestExtProcTokenExchange) and uses Ginkgo.
//
// Test architecture:
//
//	MCP Client (mcp-go) → agentgateway (Docker) → ExtProc (in-process gRPC, SUT) → Mock Identity Broker (httptest)
//	                       agentgateway (Docker) → Mock MCP Server (mcp-go, host)
//
// The test verifies that:
//  1. An MCP client sends a request with a Bearer token to agentgateway
//  2. agentgateway passes the request through ExtProc for token exchange
//  3. ExtProc exchanges the token via the mock identity broker
//  4. The exchanged token replaces the original in the Authorization header
//  5. The MCP server receives the exchanged token (not the original)
package extproc_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"

	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"

	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
	extprocserver "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/server"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/extproc/bootstrap"
)

// agentgwLogger writes structured test output to GinkgoWriter for test visibility.
var agentgwLogger = bootstrap.NewTestLogger()

// agentgwContextKey is used for storing request-scoped values in context.
type agentgwContextKey string

const agentgwAuthHeaderKey agentgwContextKey = "authorization"

const (
	agentgatewayImage = "cr.agentgateway.dev/agentgateway:0.12.0"

	// Test tokens used in agentgateway integration tests.
	agentgwOriginalBearerToken = "original-agent-bearer-token-e2e"
	agentgwExchangedToken      = "exchanged-downstream-token-e2e"
	agentgwMockAccessToken     = "mock-client-assertion-access-token"
)

// Agentgateway Integration describes the end-to-end flow through a real agentgateway
// Docker container: MCP client → agentgateway → ExtProc (SUT) → mock identity broker.
// The container setup is shared across all specs (Ordered + BeforeAll) since starting
// agentgateway is expensive.
var _ = Describe("Agentgateway Integration", Ordered, func() {
	var (
		ctx                  context.Context
		cancel               context.CancelFunc
		mockOAuth2Srv        *httptest.Server
		mockTokenExchangeSvr *agentgwMockTokenExchangeSrv
		extprocGRPC          *grpc.Server
		exchanger            *extprocserver.TokenExchanger
		mcpClient            *client.Client
	)

	BeforeAll(func() {
		_, dockerErr := testcontainers.ProviderDocker.GetProvider()
		if dockerErr != nil {
			Skip("Skipping agentgateway integration tests: Docker not available")
		}

		ctx, cancel = context.WithTimeout(context.Background(), 120*time.Second)

		// --- 1. Start mock identity broker (client_credentials + token exchange) ---
		mockOAuth2Srv = newAgentgwMockOAuth2Srv()
		mockTokenExchangeSvr = newAgentgwMockTokenExchangeSrv()

		// --- 2. Start mock MCP server using mcp-go (captures Authorization header) ---
		mcpListener := startAgentgwMCPServer()
		mcpPort := mcpListener.Addr().(*net.TCPAddr).Port
		agentgwLogger.Info("Mock MCP server listening", "port", mcpPort)

		// --- 3. Start ExtProc gRPC server (system under test) ---
		var extprocListener net.Listener
		extprocListener, extprocGRPC, exchanger = startAgentgwExtProc(
			mockOAuth2Srv.URL, mockTokenExchangeSvr.URL,
		)
		extprocPort := extprocListener.Addr().(*net.TCPAddr).Port
		agentgwLogger.Info("ExtProc gRPC server listening", "port", extprocPort)

		// --- 4. Start agentgateway Docker container ---
		agentgatewayPort := startAgentgwContainer(ctx, extprocPort, mcpPort)
		agentgwLogger.Info("agentgateway accessible on host port", "port", agentgatewayPort)

		// --- 5. Create MCP client and connect through agentgateway ---
		agentgatewayURL := fmt.Sprintf("http://localhost:%s", agentgatewayPort)
		mcpClient = connectAgentgwMCPClient(ctx, agentgatewayURL)

		DeferCleanup(func() {
			if mcpClient != nil {
				mcpClient.Close() //nolint:errcheck
			}
			if extprocGRPC != nil {
				extprocGRPC.GracefulStop()
			}
			if exchanger != nil {
				exchanger.Shutdown()
			}
			if mockTokenExchangeSvr != nil {
				mockTokenExchangeSvr.Close()
			}
			if mockOAuth2Srv != nil {
				mockOAuth2Srv.Close()
			}
			cancel()
		})
	})

	// US1-agentgw: Token exchange through agentgateway
	// Given an MCP client sends a Bearer token to agentgateway,
	// When agentgateway routes the request through ExtProc,
	// Then the MCP server receives the exchanged token (not the original).
	It("should exchange the Bearer token through agentgateway and deliver exchanged token to MCP server", func() {
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "whoami",
			},
		})
		Expect(err).NotTo(HaveOccurred(), "CallTool should succeed")
		Expect(result).NotTo(BeNil(), "CallTool result should not be nil")
		Expect(result.IsError).To(BeFalse(), "CallTool should not return an error result")
		Expect(result.Content).NotTo(BeEmpty(), "CallTool result should have content")

		// The MCP server's whoami tool returns the Authorization header it received.
		// It should contain the EXCHANGED token, not the original.
		text := agentgwExtractTextContent(result)
		agentgwLogger.Info("whoami response", "text", text)

		Expect(text).To(ContainSubstring("auth_scheme=Bearer"),
			"MCP server should have received a Bearer token")
		Expect(text).To(ContainSubstring(agentgwExchangedToken),
			"MCP server should have received the EXCHANGED token, not the original")
		Expect(text).NotTo(ContainSubstring(agentgwOriginalBearerToken),
			"MCP server must NOT receive the original Bearer token")

		Expect(mockTokenExchangeSvr.callCount()).To(BeNumerically(">=", 1),
			"Token exchange endpoint should have been called at least once")
	})

	// Echo tool: basic functional test through agentgateway
	It("should route the echo tool call through agentgateway to the MCP server", func() {
		result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "echo",
				Arguments: map[string]any{"message": "hello from e2e"},
			},
		})
		Expect(err).NotTo(HaveOccurred(), "CallTool echo should succeed")
		Expect(result).NotTo(BeNil())
		Expect(result.IsError).To(BeFalse())

		text := agentgwExtractTextContent(result)
		Expect(text).To(ContainSubstring("hello from e2e"),
			"echo tool should return the sent message")
	})
})

// --- Mock OAuth2 Server (client_credentials) ---

func newAgentgwMockOAuth2Srv() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, //nolint:errcheck
			`{"access_token":%q,"id_token":"mock-id-token","token_type":"Bearer","expires_in":3600}`,
			agentgwMockAccessToken,
		)
	})
	return httptest.NewServer(mux)
}

// --- Mock Token Exchange Server (identity broker) ---

type agentgwMockTokenExchangeSrv struct {
	*httptest.Server
	mu    sync.Mutex
	calls int
}

func newAgentgwMockTokenExchangeSrv() *agentgwMockTokenExchangeSrv {
	m := &agentgwMockTokenExchangeSrv{}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		m.mu.Lock()
		m.calls++
		m.mu.Unlock()

		agentgwLogger.Info("Token exchange called",
			"grant_type", r.FormValue("grant_type"),
			"subject_token", r.FormValue("subject_token"),
			"resource", r.FormValue("resource"),
		)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, //nolint:errcheck
			`{"access_token":%q,"issued_token_type":"urn:ietf:params:oauth:token-type:access_token","token_type":"Bearer","expires_in":3600}`,
			agentgwExchangedToken,
		)
	})
	m.Server = httptest.NewServer(mux)
	return m
}

func (m *agentgwMockTokenExchangeSrv) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// --- Mock MCP Server (using mcp-go) ---

// startAgentgwMCPServer creates an MCP server with "echo" and "whoami" tools using the mcp-go library.
// The "whoami" tool captures the Authorization header from the incoming HTTP request
// and returns it in the response, allowing the test to verify token exchange.
func startAgentgwMCPServer() net.Listener {
	mcpSvr := mcpserver.NewMCPServer(
		"e2e-mock-mcp-server", "1.0.0",
		mcpserver.WithToolCapabilities(false),
	)

	// Register "echo" tool
	echoTool := mcp.NewTool("echo",
		mcp.WithDescription("Echoes the provided message back to the caller."),
		mcp.WithString("message", mcp.Description("The message to echo"), mcp.Required()),
	)
	mcpSvr.AddTool(echoTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		msg := req.GetString("message", "")
		return mcp.NewToolResultText(fmt.Sprintf("echo: %s", msg)), nil
	})

	// Register "whoami" tool — captures Authorization header from context
	whoamiTool := mcp.NewTool("whoami",
		mcp.WithDescription("Returns the Authorization token visible to the MCP server."),
	)
	mcpSvr.AddTool(whoamiTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		authHeader, _ := ctx.Value(agentgwAuthHeaderKey).(string)
		scheme := "none"
		token := ""
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			scheme = parts[0]
			if len(parts) == 2 {
				token = parts[1]
			}
		}
		return mcp.NewToolResultText(fmt.Sprintf("auth_scheme=%s\ntoken=%s", scheme, token)), nil
	})

	// Create the streamable HTTP server with context injection for Authorization header
	httpSrv := mcpserver.NewStreamableHTTPServer(mcpSvr,
		mcpserver.WithEndpointPath("/mcp"),
		mcpserver.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			return context.WithValue(ctx, agentgwAuthHeaderKey, r.Header.Get("Authorization"))
		}),
	)

	// Listen on a random port (0.0.0.0 so Docker container can reach us)
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	Expect(err).NotTo(HaveOccurred(), "failed to create MCP server listener")

	go func() {
		srv := &http.Server{Handler: httpSrv}
		if serveErr := srv.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			agentgwLogger.Error("MCP server error", "err", serveErr)
		}
	}()

	return listener
}

// --- ExtProc gRPC Server (System Under Test) ---

func startAgentgwExtProc(
	oauth2URL, tokenExchangeURL string,
) (net.Listener, *grpc.Server, *extprocserver.TokenExchanger) {
	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{
			Bind:                 "0.0.0.0",
			Port:                 0, // will use listener port
			MaxConcurrentStreams: 100,
		},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             tokenExchangeURL + "/oauth2/token",
			Issuer:                    oauth2URL,
			ClientID:                  "e2e-extproc-client",
			ClientSecret:              "e2e-client-secret",
			ClientCredentialsEndpoint: oauth2URL + "/oauth/token",
			ClientAssertionType:       "access_token",
			ExchangeTimeout:           10 * time.Second,
			TLS: extprocconfig.TLSConfig{
				AllowHTTP: true,
			},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
		Log: extprocconfig.LogConfig{
			Level:  "debug",
			Format: "text",
		},
	}

	slogLogger := bootstrap.NewTestLogger()
	exchanger, err := extprocserver.NewTokenExchanger(cfg, slogLogger)
	Expect(err).NotTo(HaveOccurred(), "failed to create token exchanger")

	svc := extprocserver.NewServer(cfg, exchanger, slogLogger)

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	Expect(err).NotTo(HaveOccurred(), "failed to create ExtProc listener")

	grpcSrv := grpc.NewServer()
	extprocv3.RegisterExternalProcessorServer(grpcSrv, svc)

	go func() {
		if serveErr := grpcSrv.Serve(listener); serveErr != nil {
			agentgwLogger.Error("ExtProc gRPC server error", "err", serveErr)
		}
	}()

	return listener, grpcSrv, exchanger
}

// --- agentgateway Docker Container ---

func startAgentgwContainer(ctx context.Context, extprocPort, mcpPort int) string {
	// Generate agentgateway config pointing to host services
	configYAML := fmt.Sprintf(`binds:
- port: 4000
  listeners:
  - routes:
    - policies:
        extProc:
          host: "host.testcontainers.internal:%d"
          failureMode: failClosed
      backends:
      - mcp:
          targets:
          - name: tools
            mcp:
              host: http://host.testcontainers.internal:%d/mcp
`, extprocPort, mcpPort)

	agentgwLogger.Info("agentgateway config", "yaml", configYAML)

	req := testcontainers.ContainerRequest{
		Image:           agentgatewayImage,
		ExposedPorts:    []string{"4000/tcp"},
		Cmd:             []string{"-f", "/config.yaml"},
		HostAccessPorts: []int{extprocPort, mcpPort},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            strings.NewReader(configYAML),
				ContainerFilePath: "/config.yaml",
				FileMode:          0644,
			},
		},
		WaitingFor: wait.ForListeningPort("4000/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	Expect(err).NotTo(HaveOccurred(), "failed to start agentgateway container")

	DeferCleanup(func() {
		if logs, logErr := container.Logs(ctx); logErr == nil {
			buf := make([]byte, 4096)
			n, _ := logs.Read(buf)
			agentgwLogger.Info("agentgateway container logs", "logs", string(buf[:n]))
			logs.Close() //nolint:errcheck
		}
		if termErr := container.Terminate(ctx); termErr != nil {
			agentgwLogger.Error("failed to terminate agentgateway container", "err", termErr)
		}
	})

	mappedPort, err := container.MappedPort(ctx, "4000")
	Expect(err).NotTo(HaveOccurred(), "failed to get mapped port")

	return mappedPort.Port()
}

// --- MCP Client ---

func connectAgentgwMCPClient(ctx context.Context, agentgatewayURL string) *client.Client {
	mcpClient, err := client.NewStreamableHttpClient(
		agentgatewayURL+"/mcp",
		transport.WithHTTPHeaders(map[string]string{
			"Authorization": "Bearer " + agentgwOriginalBearerToken,
		}),
	)
	Expect(err).NotTo(HaveOccurred(), "failed to create MCP client")

	err = mcpClient.Start(ctx)
	Expect(err).NotTo(HaveOccurred(), "failed to start MCP client")

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "e2e-test-client",
		Version: "1.0.0",
	}

	_, err = mcpClient.Initialize(ctx, initReq)
	Expect(err).NotTo(HaveOccurred(), "MCP Initialize should succeed")

	return mcpClient
}

// --- Helpers ---

func agentgwExtractTextContent(result *mcp.CallToolResult) string {
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}
