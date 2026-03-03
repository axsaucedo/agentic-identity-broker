// Package helpers provides gRPC client helpers and request builders for ExtProc E2E tests.
// These utilities simplify construction of ProcessingRequest messages and assertion of
// ProcessingResponse messages.
package helpers

import (
	"context"
	"fmt"
	"io"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	. "github.com/onsi/gomega" //nolint:staticcheck
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ProcessingRequestBuilder builds Envoy ProcessingRequest messages for testing.
// Implements the builder pattern for fluent test construction.
type ProcessingRequestBuilder struct {
	headers map[string]string
}

// NewRequestHeaders creates a builder initialized with common pseudo-headers for ExtProc testing.
// The method path, authority, and scheme are set to sensible defaults.
func NewRequestHeaders() *ProcessingRequestBuilder {
	return &ProcessingRequestBuilder{
		headers: map[string]string{
			":method":    "GET",
			":scheme":    "http",
			":authority": "mcp-server:9003",
		},
	}
}

// WithPath sets the :path pseudo-header (used as resource parameter in token exchange).
func (b *ProcessingRequestBuilder) WithPath(path string) *ProcessingRequestBuilder {
	b.headers[":path"] = path
	return b
}

// WithBearerToken sets the Authorization header with the given Bearer token.
func (b *ProcessingRequestBuilder) WithBearerToken(token string) *ProcessingRequestBuilder {
	b.headers["authorization"] = fmt.Sprintf("Bearer %s", token)
	return b
}

// WithHeader sets an arbitrary header value.
func (b *ProcessingRequestBuilder) WithHeader(key, value string) *ProcessingRequestBuilder {
	b.headers[key] = value
	return b
}

// WithoutAuthorizationHeader removes the Authorization header (simulates requests without Bearer).
func (b *ProcessingRequestBuilder) WithoutAuthorizationHeader() *ProcessingRequestBuilder {
	delete(b.headers, "authorization")
	return b
}

// Build constructs the ProcessingRequest with request headers phase.
func (b *ProcessingRequestBuilder) Build() *extprocv3.ProcessingRequest {
	headers := make([]*corev3.HeaderValue, 0, len(b.headers))
	for k, v := range b.headers {
		headers = append(headers, &corev3.HeaderValue{
			Key:      k,
			RawValue: []byte(v),
		})
	}

	return &extprocv3.ProcessingRequest{
		Request: &extprocv3.ProcessingRequest_RequestHeaders{
			RequestHeaders: &extprocv3.HttpHeaders{
				Headers: &corev3.HeaderMap{
					Headers: headers,
				},
			},
		},
	}
}

// SendRequestHeaders sends a RequestHeaders phase message to ExtProc and returns the response.
// This is the primary processing phase where token exchange occurs.
// Handles the streaming gRPC protocol: sends one message, receives one response.
func SendRequestHeaders(
	ctx context.Context,
	client extprocv3.ExternalProcessorClient,
	req *extprocv3.ProcessingRequest,
) *extprocv3.ProcessingResponse {
	// Create streaming context with timeout
	streamCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	stream, err := client.Process(streamCtx)
	Expect(err).NotTo(HaveOccurred(), "failed to open gRPC stream")

	// Send the request headers
	err = stream.Send(req)
	Expect(err).NotTo(HaveOccurred(), "failed to send ProcessingRequest")

	// Close the send side (we're done sending headers for this request)
	err = stream.CloseSend()
	Expect(err).NotTo(HaveOccurred(), "failed to close send")

	// Receive the response
	resp, err := stream.Recv()
	Expect(err).NotTo(HaveOccurred(), "failed to receive ProcessingResponse")

	return resp
}

// SendRequestHeadersWithError sends a request headers message and expects to receive an error
// or an ImmediateResponse (failure path). Returns the response without failing the test.
func SendRequestHeadersWithError(
	ctx context.Context,
	client extprocv3.ExternalProcessorClient,
	req *extprocv3.ProcessingRequest,
) (*extprocv3.ProcessingResponse, error) {
	streamCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	stream, err := client.Process(streamCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to open gRPC stream: %w", err)
	}

	if err := stream.Send(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if err := stream.CloseSend(); err != nil {
		return nil, fmt.Errorf("failed to close send: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("stream closed without response (EOF)")
		}
		return nil, fmt.Errorf("stream recv error: %w", err)
	}

	return resp, nil
}

// ConnectToExtProc creates a gRPC connection to the ExtProc server at addr.
// Returns the client and connection. Caller must call conn.Close() when done.
func ConnectToExtProc(addr string) (extprocv3.ExternalProcessorClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred(), "failed to connect to ExtProc server at "+addr)
	return extprocv3.NewExternalProcessorClient(conn), conn
}

// ExtractMutatedAuthorizationHeader extracts the Authorization header value from a
// HeadersResponse that mutates headers. Returns empty string if not found.
func ExtractMutatedAuthorizationHeader(resp *extprocv3.ProcessingResponse) string {
	headersResp, ok := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders)
	if !ok {
		return ""
	}

	if headersResp.RequestHeaders == nil ||
		headersResp.RequestHeaders.Response == nil ||
		headersResp.RequestHeaders.Response.HeaderMutation == nil {
		return ""
	}

	for _, hvo := range headersResp.RequestHeaders.Response.HeaderMutation.SetHeaders {
		if hvo.Header != nil && hvo.Header.Key == "authorization" {
			return string(hvo.Header.RawValue)
		}
	}

	return ""
}

// ExtractImmediateResponseStatus extracts the HTTP status code from an ImmediateResponse.
// Returns 0 if the response is not an ImmediateResponse.
func ExtractImmediateResponseStatus(resp *extprocv3.ProcessingResponse) uint32 {
	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	if !ok {
		return 0
	}

	if immResp.ImmediateResponse == nil || immResp.ImmediateResponse.Status == nil {
		return 0
	}

	return uint32(immResp.ImmediateResponse.Status.Code)
}

// ExtractImmediateResponseBody extracts the body from an ImmediateResponse.
// Returns empty string if not an ImmediateResponse or no body.
func ExtractImmediateResponseBody(resp *extprocv3.ProcessingResponse) string {
	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	if !ok {
		return ""
	}

	if immResp.ImmediateResponse == nil {
		return ""
	}

	return string(immResp.ImmediateResponse.Body)
}

// IsPassThroughResponse returns true if the response is a pass-through (empty HeadersResponse
// with no mutations). This is the expected response when no Bearer token is present.
func IsPassThroughResponse(resp *extprocv3.ProcessingResponse) bool {
	headersResp, ok := resp.Response.(*extprocv3.ProcessingResponse_RequestHeaders)
	if !ok {
		return false
	}

	if headersResp.RequestHeaders == nil {
		return true
	}

	// Pass-through has no header mutations
	if headersResp.RequestHeaders.Response == nil {
		return true
	}

	mutation := headersResp.RequestHeaders.Response.HeaderMutation
	return mutation == nil || (len(mutation.SetHeaders) == 0 && len(mutation.RemoveHeaders) == 0)
}
