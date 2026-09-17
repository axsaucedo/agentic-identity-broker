package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"sigs.k8s.io/yaml"
)

const (
	defaultAgentgatewayImage       = "cr.agentgateway.dev/agentgateway:v1.5.0"
	gatewayPort                    = "4000/tcp"
	gatewayConfigPath              = "/config.yaml"
	gatewaySigningKeyContainerPath = "/etc/agentgateway/keys/gateway-signing.key"
)

// AgentgatewayOptions names the documented direct-route values substituted into
// the committed agentgateway configuration for one E2E environment.
type AgentgatewayOptions struct {
	DirectReferenceYAML      []byte
	BrokerURL                string
	BackendURL               string
	SigningKeyPath           string
	KeyID                    string
	Resource                 string
	ClientID                 string
	AssertionAudience        string
	ClientAuthAlgorithm      string
	OmitResource             bool
	ExpectConfigurationError bool
	DisableCache             bool
	SkipRouteLiveness        bool
}

// Agentgateway is a real pinned agentgateway instance serving the direct native
// token-exchange route. It owns only the container and its unused ExtProc stand-in.
type Agentgateway struct {
	BaseURL        string
	RenderedConfig []byte
	ExtProc        *ExtProcStandIn

	container testcontainers.Container
	closeOnce sync.Once
	closeErr  error
}

// AgentgatewayMCPClient is an initialized MCP client for one gateway route.
type AgentgatewayMCPClient struct {
	client *client.Client
}

type directRouteConfiguration struct {
	mcpTarget     map[string]any
	tokenExchange map[string]any
	clientAuth    map[string]any
	signingKey    map[string]any
}

var compiledAgentgatewayV150Schema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	schemaPath, err := repositoryPath("tests", "e2e", "gateway", "testdata", "agentgateway-config-v1.5.0.schema.json")
	if err != nil {
		return nil, err
	}
	schemaJSON, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("read pinned agentgateway v1.5.0 schema %q: %w", schemaPath, err)
	}

	var schemaDocument any
	if err := json.Unmarshal(schemaJSON, &schemaDocument); err != nil {
		return nil, fmt.Errorf("parse pinned agentgateway v1.5.0 schema: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	const schemaURL = "https://agentic-identity-broker.test/agentgateway/v1.5.0/config.json"
	if err := compiler.AddResource(schemaURL, schemaDocument); err != nil {
		return nil, fmt.Errorf("register pinned agentgateway v1.5.0 schema: %w", err)
	}
	schema, err := compiler.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("compile pinned agentgateway v1.5.0 schema: %w", err)
	}
	return schema, nil
})

// LoadAgentgatewayDirectReference reads and validates the exact direct-route
// configuration that an operator copies.
func LoadAgentgatewayDirectReference() ([]byte, error) {
	filePath, err := repositoryPath("examples", "agentgateway", "direct-token-exchange.yaml")
	if err != nil {
		return nil, err
	}

	config, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read committed agentgateway reference configuration %q: %w", filePath, err)
	}
	if err := ValidateAgentgatewayV150Config(config); err != nil {
		return nil, fmt.Errorf("validate committed agentgateway reference configuration: %w", err)
	}
	if _, _, err := validateDirectReference(config); err != nil {
		return nil, err
	}
	return config, nil
}

// StartAgentgateway renders the committed direct route, starts the pinned
// agentgateway image, and confirms that the configured MCP route responds.
func StartAgentgateway(ctx context.Context, options AgentgatewayOptions) (*Agentgateway, error) {
	if ctx == nil {
		return nil, errors.New("start agentgateway: nil context")
	}

	referenceYAML := options.DirectReferenceYAML
	if len(referenceYAML) == 0 {
		var err error
		referenceYAML, err = LoadAgentgatewayDirectReference()
		if err != nil {
			return nil, err
		}
	}

	renderedConfig, err := renderDirectRoute(referenceYAML, options)
	if err != nil {
		return nil, err
	}

	brokerPort, err := brokerHostAccessPort(options.BrokerURL)
	if err != nil {
		return nil, err
	}
	backendPort, err := backendHostAccessPort(options.BackendURL)
	if err != nil {
		return nil, err
	}
	privateKey, err := openSigningKey(options.SigningKeyPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = privateKey.Close() }()

	if err := ensureTestcontainersCleanup(); err != nil {
		return nil, err
	}
	standIn, err := StartExtProcStandIn()
	if err != nil {
		return nil, fmt.Errorf("start ExtProc stand-in: %w", err)
	}
	cleanupStandIn := true
	defer func() {
		if cleanupStandIn {
			_ = standIn.Close()
		}
	}()

	image := agentgatewayImage()
	request := testcontainers.ContainerRequest{
		Image:           image,
		ExposedPorts:    []string{gatewayPort},
		Cmd:             []string{"-f", gatewayConfigPath},
		HostAccessPorts: uniquePorts(brokerPort, backendPort, standIn.port()),
		Files: []testcontainers.ContainerFile{
			{
				Reader:            bytes.NewReader(renderedConfig),
				ContainerFilePath: gatewayConfigPath,
				FileMode:          0o644,
			},
			{
				Reader:            privateKey,
				ContainerFilePath: gatewaySigningKeyContainerPath,
				FileMode:          0o644,
			},
		},
		WaitingFor: wait.ForListeningPort(gatewayPort).WithStartupTimeout(30 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("start agentgateway container using %q: %w", image, err)
	}
	cleanupContainer := true
	defer func() {
		if cleanupContainer {
			_ = container.Terminate(context.Background())
		}
	}()

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve agentgateway container host: %w", err)
	}
	mappedPort, err := container.MappedPort(ctx, gatewayPort)
	if err != nil {
		return nil, fmt.Errorf("resolve agentgateway container port: %w", err)
	}

	gateway := &Agentgateway{
		BaseURL:        "http://" + net.JoinHostPort(host, mappedPort.Port()),
		RenderedConfig: renderedConfig,
		ExtProc:        standIn,
		container:      container,
	}
	if !options.SkipRouteLiveness {
		if err := gateway.AssertRouteLiveness(ctx); err != nil {
			return nil, err
		}
	}

	cleanupContainer = false
	cleanupStandIn = false
	return gateway, nil
}

// CallWhoAmI calls the route's MCP whoami tool with bearerToken.
func (g *Agentgateway) CallWhoAmI(ctx context.Context, bearerToken string) (string, error) {
	mcpClient, err := g.OpenMCPClient(ctx, bearerToken)
	if err != nil {
		return "", err
	}
	defer func() { _ = mcpClient.Close() }()
	return mcpClient.CallWhoAmI(ctx)
}

// OpenMCPClient connects and initializes a gateway MCP session.
func (g *Agentgateway) OpenMCPClient(ctx context.Context, bearerToken string) (*AgentgatewayMCPClient, error) {
	if g == nil || g.BaseURL == "" {
		return nil, errors.New("open agentgateway MCP client: gateway is not initialized")
	}
	if strings.TrimSpace(bearerToken) == "" {
		return nil, errors.New("open agentgateway MCP client: bearer token is required")
	}
	mcpClient, err := client.NewStreamableHttpClient(
		g.BaseURL+"/mcp",
		transport.WithHTTPHeaders(map[string]string{"Authorization": "Bearer " + bearerToken}),
	)
	if err != nil {
		return nil, fmt.Errorf("create MCP client for agentgateway: %w", err)
	}
	if err := mcpClient.Start(ctx); err != nil {
		_ = mcpClient.Close()
		return nil, fmt.Errorf("start MCP client for agentgateway: %w", err)
	}
	initialize := mcp.InitializeRequest{}
	initialize.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initialize.Params.ClientInfo = mcp.Implementation{Name: "gateway-direct-e2e", Version: "1.0"}
	if _, err := mcpClient.Initialize(ctx, initialize); err != nil {
		_ = mcpClient.Close()
		return nil, fmt.Errorf("initialize MCP client through agentgateway: %w", err)
	}
	return &AgentgatewayMCPClient{client: mcpClient}, nil
}

// CallWhoAmI calls the initialized gateway route's MCP whoami tool.
func (c *AgentgatewayMCPClient) CallWhoAmI(ctx context.Context) (string, error) {
	if c == nil || c.client == nil {
		return "", errors.New("call MCP whoami through agentgateway: client is not initialized")
	}
	result, err := c.client.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "whoami"}})
	if err != nil {
		return "", fmt.Errorf("call MCP whoami through agentgateway: %w", err)
	}
	if result == nil {
		return "", errors.New("call MCP whoami through agentgateway: empty result")
	}
	if result.IsError {
		return "", errors.New("call MCP whoami through agentgateway: backend returned an error")
	}
	for _, content := range result.Content {
		if text, ok := content.(mcp.TextContent); ok {
			return text.Text, nil
		}
	}
	return "", errors.New("call MCP whoami through agentgateway: response has no text content")
}

// Close closes the MCP session.
func (c *AgentgatewayMCPClient) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

// AssertRouteLiveness confirms the gateway serves the configured MCP route. A
// missing bearer can fail native authentication, but a 404 means no route served.
func (g *Agentgateway) AssertRouteLiveness(ctx context.Context) error {
	if g == nil || g.BaseURL == "" {
		return errors.New("check agentgateway route liveness: gateway is not initialized")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, g.BaseURL+"/mcp", nil)
	if err != nil {
		return fmt.Errorf("create agentgateway route-liveness request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("agentgateway MCP route is not live: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusNotFound {
		return errors.New("agentgateway MCP route is not live: received 404")
	}
	return nil
}

// Close stops the container and ExtProc stand-in. It is safe to call more than once.
func (g *Agentgateway) Close(ctx context.Context) error {
	if g == nil {
		return nil
	}
	g.closeOnce.Do(func() {
		var errs []error
		if g.container != nil {
			if err := g.container.Terminate(ctx); err != nil {
				errs = append(errs, fmt.Errorf("terminate agentgateway container: %w", err))
			}
		}
		if g.ExtProc != nil {
			if err := g.ExtProc.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		g.closeErr = errors.Join(errs...)
	})
	return g.closeErr
}

func renderDirectRoute(referenceYAML []byte, options AgentgatewayOptions) ([]byte, error) {
	if err := ValidateAgentgatewayV150Config(referenceYAML); err != nil {
		return nil, fmt.Errorf("validate direct-route reference configuration: %w", err)
	}
	document, directRoute, err := validateDirectReference(referenceYAML)
	if err != nil {
		return nil, err
	}

	brokerHost, err := brokerHost(options.BrokerURL)
	if err != nil {
		return nil, err
	}
	if _, err := backendHostAccessPort(options.BackendURL); err != nil {
		return nil, err
	}
	if err := validateRenderOptions(options); err != nil {
		return nil, err
	}

	directRoute.mcpTarget["host"] = options.BackendURL
	directRoute.tokenExchange["host"] = brokerHost
	if options.OmitResource {
		delete(directRoute.tokenExchange, "resources")
	} else {
		directRoute.tokenExchange["resources"] = []any{options.Resource}
	}
	directRoute.clientAuth["clientId"] = options.ClientID
	directRoute.clientAuth["assertionAudience"] = options.AssertionAudience
	directRoute.clientAuth["kid"] = options.KeyID
	if options.ClientAuthAlgorithm != "" {
		directRoute.clientAuth["alg"] = options.ClientAuthAlgorithm
	}
	directRoute.signingKey["file"] = gatewaySigningKeyContainerPath
	if options.DisableCache {
		directRoute.tokenExchange["cache"] = map[string]any{"maxEntries": 0}
	}

	rendered, err := yaml.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("marshal direct agentgateway configuration: %w", err)
	}
	if !options.ExpectConfigurationError {
		if err := ValidateAgentgatewayV150Config(rendered); err != nil {
			return nil, fmt.Errorf("validate rendered direct agentgateway configuration: %w", err)
		}
		if !options.OmitResource {
			if _, _, err := validateDirectReference(rendered); err != nil {
				return nil, fmt.Errorf("validate rendered direct agentgateway configuration invariants: %w", err)
			}
		}
	}
	return rendered, nil
}

func validateRenderOptions(options AgentgatewayOptions) error {
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "SigningKeyPath", value: options.SigningKeyPath},
		{name: "KeyID", value: options.KeyID},
		{name: "ClientID", value: options.ClientID},
		{name: "AssertionAudience", value: options.AssertionAudience},
	} {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("render direct agentgateway configuration: %s is required", field.name)
		}
	}
	if !options.OmitResource {
		if strings.TrimSpace(options.Resource) == "" {
			return errors.New("render direct agentgateway configuration: Resource is required")
		}
		if err := validateAbsoluteResource(options.Resource); err != nil {
			return fmt.Errorf("render direct agentgateway configuration: %w", err)
		}
	}
	if options.ClientAuthAlgorithm != "" && !options.ExpectConfigurationError {
		return errors.New("render direct agentgateway configuration: ClientAuthAlgorithm requires ExpectConfigurationError")
	}
	return nil
}

func validateDirectReference(configYAML []byte) (map[string]any, *directRouteConfiguration, error) {
	var document map[string]any
	if err := yaml.Unmarshal(configYAML, &document); err != nil {
		return nil, nil, fmt.Errorf("parse direct agentgateway reference configuration: %w", err)
	}
	if len(document) == 0 {
		return nil, nil, errors.New("direct agentgateway reference configuration must not be empty")
	}
	if containsConfigKey(document, "extProc") {
		return nil, nil, errors.New("direct agentgateway reference configuration must not configure extProc")
	}
	for _, forbiddenKey := range []string{"clientSecret", "secretRef"} {
		if containsConfigKey(document, forbiddenKey) {
			return nil, nil, fmt.Errorf("direct agentgateway reference configuration must not configure %s", forbiddenKey)
		}
	}
	for _, forbiddenValue := range []string{"clientSecretBasic", "clientSecretPost", "jwtBearer"} {
		if containsConfigString(document, forbiddenValue) {
			return nil, nil, fmt.Errorf("direct agentgateway reference configuration must not use %q", forbiddenValue)
		}
	}
	if containsConfigStringPrefix(document, "-----BEGIN") {
		return nil, nil, errors.New("direct agentgateway reference configuration must not embed private key material")
	}

	directRoute, err := findDirectRoute(document)
	if err != nil {
		return nil, nil, err
	}
	if err := validateDirectRouteValues(directRoute); err != nil {
		return nil, nil, err
	}
	return document, directRoute, nil
}

func findDirectRoute(document map[string]any) (*directRouteConfiguration, error) {
	binds, err := configSlice(document, "binds", "direct agentgateway reference configuration")
	if err != nil || len(binds) == 0 {
		return nil, errors.New("direct agentgateway reference configuration must contain a bind")
	}

	var matches []*directRouteConfiguration
	for bindIndex, bindValue := range binds {
		bind, err := configObject(bindValue, fmt.Sprintf("binds[%d]", bindIndex))
		if err != nil {
			return nil, err
		}
		listeners, err := configSlice(bind, "listeners", fmt.Sprintf("binds[%d]", bindIndex))
		if err != nil {
			return nil, err
		}
		for listenerIndex, listenerValue := range listeners {
			listener, err := configObject(listenerValue, fmt.Sprintf("binds[%d].listeners[%d]", bindIndex, listenerIndex))
			if err != nil {
				return nil, err
			}
			routes, err := configSlice(listener, "routes", fmt.Sprintf("binds[%d].listeners[%d]", bindIndex, listenerIndex))
			if err != nil {
				return nil, err
			}
			for routeIndex, routeValue := range routes {
				route, err := configObject(routeValue, fmt.Sprintf("binds[%d].listeners[%d].routes[%d]", bindIndex, listenerIndex, routeIndex))
				if err != nil {
					return nil, err
				}
				backends, err := configSlice(route, "backends", fmt.Sprintf("binds[%d].listeners[%d].routes[%d]", bindIndex, listenerIndex, routeIndex))
				if err != nil {
					return nil, err
				}
				for backendIndex, backendValue := range backends {
					backend, err := configObject(backendValue, fmt.Sprintf("binds[%d].listeners[%d].routes[%d].backends[%d]", bindIndex, listenerIndex, routeIndex, backendIndex))
					if err != nil {
						return nil, err
					}
					candidate, ok, err := directMCPRoute(route, backend)
					if err != nil {
						return nil, err
					}
					if ok {
						matches = append(matches, candidate)
					}
				}
			}
		}
	}
	if len(matches) != 1 {
		return nil, fmt.Errorf("direct agentgateway reference configuration must contain exactly one MCP oauthTokenExchange backend, found %d", len(matches))
	}
	return matches[0], nil
}

func directMCPRoute(_ map[string]any, backend map[string]any) (*directRouteConfiguration, bool, error) {
	mcpConfig, ok := backend["mcp"].(map[string]any)
	if !ok {
		return nil, false, nil
	}
	policies, ok := backend["policies"].(map[string]any)
	if !ok {
		return nil, false, nil
	}
	backendAuth, ok := policies["backendAuth"].(map[string]any)
	if !ok {
		return nil, false, nil
	}
	tokenExchange, ok := backendAuth["oauthTokenExchange"].(map[string]any)
	if !ok {
		return nil, false, nil
	}
	if len(backendAuth) != 1 {
		return nil, false, errors.New("direct agentgateway reference configuration must configure oauthTokenExchange as its only backend-auth method")
	}
	targets, err := configSlice(mcpConfig, "targets", "direct MCP backend")
	if err != nil {
		return nil, false, err
	}
	if len(targets) != 1 {
		return nil, false, fmt.Errorf("direct MCP backend must contain exactly one target, found %d", len(targets))
	}
	target, err := configObject(targets[0], "direct MCP backend target")
	if err != nil {
		return nil, false, err
	}
	targetMCP, ok := target["mcp"].(map[string]any)
	if !ok {
		return nil, false, errors.New("direct MCP backend target must use the streamable MCP transport")
	}
	clientAuth, ok := tokenExchange["clientAuth"].(map[string]any)
	if !ok {
		return nil, false, errors.New("direct oauthTokenExchange policy must contain clientAuth")
	}
	signingKey, ok := clientAuth["signingKey"].(map[string]any)
	if !ok {
		return nil, false, errors.New("direct oauthTokenExchange clientAuth must source signingKey from a file")
	}
	return &directRouteConfiguration{
		mcpTarget:     targetMCP,
		tokenExchange: tokenExchange,
		clientAuth:    clientAuth,
		signingKey:    signingKey,
	}, true, nil
}

func validateDirectRouteValues(route *directRouteConfiguration) error {
	if value, ok := route.tokenExchange["host"].(string); !ok || strings.TrimSpace(value) == "" {
		return errors.New("direct oauthTokenExchange policy must contain a host")
	}
	if value, ok := route.tokenExchange["path"].(string); !ok || value != "/oauth2/token" {
		return errors.New("direct oauthTokenExchange policy path must be /oauth2/token")
	}
	if grantType, exists := route.tokenExchange["grantType"]; exists {
		if grantType != "tokenExchange" {
			return errors.New("direct oauthTokenExchange policy must use the RFC 8693 tokenExchange grant")
		}
	}
	resources, err := configSlice(route.tokenExchange, "resources", "direct oauthTokenExchange policy")
	if err != nil {
		return err
	}
	if len(resources) != 1 {
		return fmt.Errorf("direct oauthTokenExchange policy must contain exactly one resource, found %d", len(resources))
	}
	resource, ok := resources[0].(string)
	if !ok {
		return errors.New("direct oauthTokenExchange resource must be a string")
	}
	if err := validateAbsoluteResource(resource); err != nil {
		return err
	}
	if route.clientAuth["method"] != "privateKeyJwt" {
		return errors.New("direct oauthTokenExchange clientAuth.method must be privateKeyJwt")
	}
	for _, field := range []string{"clientId", "assertionAudience"} {
		value, ok := route.clientAuth[field].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return fmt.Errorf("direct oauthTokenExchange clientAuth.%s must be a non-empty string", field)
		}
	}
	if len(route.signingKey) != 1 {
		return errors.New("direct oauthTokenExchange signingKey must contain only a file reference")
	}
	keyPath, ok := route.signingKey["file"].(string)
	if !ok || !path.IsAbs(keyPath) {
		return errors.New("direct oauthTokenExchange signingKey.file must be an absolute container path")
	}
	if alg, exists := route.clientAuth["alg"]; exists {
		if alg != "RS256" && alg != "RS384" && alg != "RS512" {
			return errors.New("direct oauthTokenExchange clientAuth.alg must be compatible with the RSA gateway signing key")
		}
	}
	return nil
}

func validateAbsoluteResource(value string) error {
	resource, err := url.ParseRequestURI(value)
	if err != nil || resource.Scheme == "" || resource.Host == "" || resource.Fragment != "" {
		return fmt.Errorf("direct oauthTokenExchange resource must be an absolute URI without a fragment: %q", value)
	}
	return nil
}

// ValidateAgentgatewayV150Config validates YAML against the schema pinned to the
// agentgateway v1.5.0 test image.
func ValidateAgentgatewayV150Config(configYAML []byte) error {
	schema, err := compiledAgentgatewayV150Schema()
	if err != nil {
		return err
	}

	configJSON, err := yaml.YAMLToJSON(configYAML)
	if err != nil {
		return fmt.Errorf("convert agentgateway configuration to JSON: %w", err)
	}
	var configDocument any
	if err := json.Unmarshal(configJSON, &configDocument); err != nil {
		return fmt.Errorf("parse agentgateway configuration JSON: %w", err)
	}
	if err := schema.Validate(configDocument); err != nil {
		return fmt.Errorf("agentgateway configuration does not satisfy the pinned v1.5.0 schema: %w", err)
	}
	return nil
}

func brokerHost(endpoint string) (string, error) {
	parsed, _, err := parseContainerEndpoint("BrokerURL", endpoint)
	if err != nil {
		return "", err
	}
	if parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("render direct agentgateway configuration: BrokerURL must not include a path, query, or fragment")
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}

func brokerHostAccessPort(endpoint string) (int, error) {
	_, port, err := parseContainerEndpoint("BrokerURL", endpoint)
	return port, err
}

func backendHostAccessPort(endpoint string) (int, error) {
	parsed, port, err := parseContainerEndpoint("BackendURL", endpoint)
	if err != nil {
		return 0, err
	}
	if parsed.Path == "" || parsed.Path == "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return 0, errors.New("render direct agentgateway configuration: BackendURL must identify an MCP endpoint without a query or fragment")
	}
	return port, nil
}

func parseContainerEndpoint(field, endpoint string) (*url.URL, int, error) {
	parsed, err := url.ParseRequestURI(endpoint)
	if err != nil || parsed == nil {
		return nil, 0, fmt.Errorf("render direct agentgateway configuration: %s must be an absolute HTTP URL: %q", field, endpoint)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return nil, 0, fmt.Errorf("render direct agentgateway configuration: %s must be an absolute HTTP URL without user credentials: %q", field, endpoint)
	}
	portValue := parsed.Port()
	if portValue == "" {
		return nil, 0, fmt.Errorf("render direct agentgateway configuration: %s must include a host-access port: %q", field, endpoint)
	}
	port, err := strconv.Atoi(portValue)
	if err != nil || port < 1 || port > 65535 {
		return nil, 0, fmt.Errorf("render direct agentgateway configuration: %s has invalid port %q", field, portValue)
	}
	return parsed, port, nil
}

func openSigningKey(keyPath string) (*os.File, error) {
	if strings.TrimSpace(keyPath) == "" {
		return nil, errors.New("start agentgateway: SigningKeyPath is required")
	}
	key, err := os.Open(keyPath)
	if err != nil {
		return nil, fmt.Errorf("open gateway signing key %q: %w", keyPath, err)
	}
	info, err := key.Stat()
	if err != nil {
		_ = key.Close()
		return nil, fmt.Errorf("stat gateway signing key %q: %w", keyPath, err)
	}
	if !info.Mode().IsRegular() {
		_ = key.Close()
		return nil, fmt.Errorf("gateway signing key %q must be a regular file", keyPath)
	}
	return key, nil
}

func uniquePorts(ports ...int) []int {
	seen := make(map[int]struct{}, len(ports))
	unique := make([]int, 0, len(ports))
	for _, port := range ports {
		if _, ok := seen[port]; ok {
			continue
		}
		seen[port] = struct{}{}
		unique = append(unique, port)
	}
	return unique
}

func configSlice(object map[string]any, key, parent string) ([]any, error) {
	value, ok := object[key]
	if !ok {
		return nil, fmt.Errorf("%s must contain %q", parent, key)
	}
	values, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s.%s must be an array", parent, key)
	}
	return values, nil
}

func configObject(value any, name string) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("direct agentgateway reference configuration %s must be an object", name)
	}
	return object, nil
}

func containsConfigKey(value any, key string) bool {
	switch typed := value.(type) {
	case map[string]any:
		for currentKey, currentValue := range typed {
			if currentKey == key || containsConfigKey(currentValue, key) {
				return true
			}
		}
	case []any:
		for _, currentValue := range typed {
			if containsConfigKey(currentValue, key) {
				return true
			}
		}
	}
	return false
}

func containsConfigString(value any, wanted string) bool {
	switch typed := value.(type) {
	case map[string]any:
		for _, currentValue := range typed {
			if containsConfigString(currentValue, wanted) {
				return true
			}
		}
	case []any:
		for _, currentValue := range typed {
			if containsConfigString(currentValue, wanted) {
				return true
			}
		}
	case string:
		return typed == wanted
	}
	return false
}

func containsConfigStringPrefix(value any, prefix string) bool {
	switch typed := value.(type) {
	case map[string]any:
		for _, currentValue := range typed {
			if containsConfigStringPrefix(currentValue, prefix) {
				return true
			}
		}
	case []any:
		for _, currentValue := range typed {
			if containsConfigStringPrefix(currentValue, prefix) {
				return true
			}
		}
	case string:
		return strings.HasPrefix(typed, prefix)
	}
	return false
}

func ensureTestcontainersCleanup() error {
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") != "" {
		return nil
	}
	if err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"); err != nil {
		return fmt.Errorf("set TESTCONTAINERS_RYUK_DISABLED: %w", err)
	}
	return nil
}

func agentgatewayImage() string {
	if image := os.Getenv("AGENTGATEWAY_IMAGE"); image != "" {
		return image
	}
	return defaultAgentgatewayImage
}

func repositoryPath(elements ...string) (string, error) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("locate agentgateway test support source")
	}
	parts := append([]string{filepath.Dir(sourceFile), "..", "..", "..", ".."}, elements...)
	return filepath.Join(parts...), nil
}
