package approval

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/approval"
)

func TestCreateHandler_RequiresMetadataDescription(t *testing.T) {
	handler := NewCreateHandler(&approval.Service{})

	for _, body := range []string{
		`{"tool_name":"read_file","arguments":{}}`,
		`{"metadata":{},"tool_name":"read_file","arguments":{}}`,
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/approvals", strings.NewReader(body))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	}
}

func TestTraceparentFromRequest(t *testing.T) {
	valid := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	req := httptest.NewRequest(http.MethodPost, "/api/approvals", nil)
	req.Header.Set("Traceparent", valid)

	got := traceparentFromRequest(req)
	if got == nil || *got != valid {
		t.Fatalf("traceparentFromRequest() = %v, want %q", got, valid)
	}

	req.Header.Set("Traceparent", "not-a-traceparent")
	if got := traceparentFromRequest(req); got != nil {
		t.Fatalf("traceparentFromRequest() = %q, want nil", *got)
	}
}
