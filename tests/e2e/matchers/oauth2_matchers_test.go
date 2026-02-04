// Package matchers provides custom Gomega matchers for OAuth2-specific assertions.
package matchers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/matchers"
)

func TestOAuth2Matchers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OAuth2 Matchers Suite")
}

var _ = Describe("HaveOAuth2Redirect", func() {
	It("should match a successful redirect response", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=auth-code&state=state-123")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).To(matchers.HaveOAuth2Redirect("https://callback.example.com"))
	})

	It("should match redirect with 303 See Other status", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=auth-code")
		w.WriteHeader(http.StatusSeeOther)

		Expect(w.Result()).To(matchers.HaveOAuth2Redirect("https://callback.example.com"))
	})

	It("should fail if response is not a redirect", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusOK)

		Expect(w.Result()).NotTo(matchers.HaveOAuth2Redirect(""))
	})

	It("should fail if Location header is missing", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).NotTo(matchers.HaveOAuth2Redirect(""))
	})

	It("should fail if Location header contains invalid URL", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "not a valid url ://")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).NotTo(matchers.HaveOAuth2Redirect(""))
	})
})

var _ = Describe("HaveOAuth2Error", func() {
	It("should match error parameter in redirect URL", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?error=invalid_request&error_description=Bad+Request")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).To(matchers.HaveOAuth2Error("invalid_request"))
	})

	It("should match any error if no specific code expected", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?error=access_denied")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).To(matchers.HaveOAuth2Error(""))
	})

	It("should fail if wrong error code", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?error=invalid_client")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).NotTo(matchers.HaveOAuth2Error("invalid_request"))
	})

	It("should fail if no error parameter present", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=auth-code")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).NotTo(matchers.HaveOAuth2Error("invalid_request"))
	})

	It("should fail if Location header missing", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).NotTo(matchers.HaveOAuth2Error("any_error"))
	})
})

var _ = Describe("HaveOAuth2SuccessRedirect", func() {
	It("should match successful redirect with code and state", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=auth-code-123&state=state-xyz")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).To(matchers.HaveOAuth2SuccessRedirect("state-xyz"))
	})

	It("should match redirect with any state if not specified", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=auth-code&state=any-state")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).To(matchers.HaveOAuth2SuccessRedirect(""))
	})

	It("should fail if state doesn't match", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=auth-code&state=state-abc")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).NotTo(matchers.HaveOAuth2SuccessRedirect("state-xyz"))
	})

	It("should fail if code parameter missing", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?state=state-123")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).NotTo(matchers.HaveOAuth2SuccessRedirect("state-123"))
	})

	It("should fail if error parameter present", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?error=access_denied&state=state-123")
		w.WriteHeader(http.StatusFound)

		Expect(w.Result()).NotTo(matchers.HaveOAuth2SuccessRedirect("state-123"))
	})

	It("should fail if not a redirect response", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusOK)

		Expect(w.Result()).NotTo(matchers.HaveOAuth2SuccessRedirect("state-123"))
	})
})

var _ = Describe("HaveMetadataField", func() {
	It("should match when metadata field is present", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")

		// Write body with fields that include the field names
		// Important: WriteString before WriteHeader to allow proper reading
		jsonBody := `{"issuer":"https://auth.example.com","authorization_endpoint":"https://auth.example.com/authorize","token_endpoint":"https://auth.example.com/token"}`
		_, _ = w.WriteString(jsonBody)
		w.WriteHeader(http.StatusOK)

		resp := w.Result()
		Expect(resp).To(matchers.HaveMetadataField("issuer"))

		// Create new responses for each test since body can only be read once
		w2 := httptest.NewRecorder()
		w2.Header().Set("Content-Type", "application/json")
		_, _ = w2.WriteString(jsonBody)
		w2.WriteHeader(http.StatusOK)

		Expect(w2.Result()).To(matchers.HaveMetadataField("authorization_endpoint"))

		w3 := httptest.NewRecorder()
		w3.Header().Set("Content-Type", "application/json")
		_, _ = w3.WriteString(jsonBody)
		w3.WriteHeader(http.StatusOK)

		Expect(w3.Result()).To(matchers.HaveMetadataField("token_endpoint"))
	})

	It("should fail when metadata field is missing", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")

		metadata := map[string]interface{}{
			"issuer": "https://auth.example.com",
		}
		_ = json.NewEncoder(w).Encode(metadata)
		w.WriteHeader(http.StatusOK)

		Expect(w.Result()).NotTo(matchers.HaveMetadataField("token_endpoint"))
	})

	It("should fail if not a 200 response", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		Expect(w.Result()).NotTo(matchers.HaveMetadataField("issuer"))
	})

	It("should fail if content type is not JSON", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		Expect(w.Result()).NotTo(matchers.HaveMetadataField("issuer"))
	})
})

var _ = Describe("HaveAllQueryParams", func() {
	It("should match when all expected query params are present", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=auth-code&state=state-123&session_state=sess-456")
		w.WriteHeader(http.StatusFound)

		expectedParams := map[string]string{
			"code":  "auth-code",
			"state": "state-123",
		}
		Expect(w.Result()).To(matchers.HaveAllQueryParams(expectedParams))
	})

	It("should fail if a query param is missing", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=auth-code")
		w.WriteHeader(http.StatusFound)

		expectedParams := map[string]string{
			"code":  "auth-code",
			"state": "state-123",
		}
		Expect(w.Result()).NotTo(matchers.HaveAllQueryParams(expectedParams))
	})

	It("should fail if a query param has wrong value", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=wrong-code&state=state-123")
		w.WriteHeader(http.StatusFound)

		expectedParams := map[string]string{
			"code":  "auth-code",
			"state": "state-123",
		}
		Expect(w.Result()).NotTo(matchers.HaveAllQueryParams(expectedParams))
	})

	It("should succeed with partial match when only some params specified", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=auth-code&state=state-123&extra=param")
		w.WriteHeader(http.StatusFound)

		expectedParams := map[string]string{
			"code": "auth-code",
		}
		Expect(w.Result()).To(matchers.HaveAllQueryParams(expectedParams))
	})

	It("should fail if Location header is missing", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusFound)

		expectedParams := map[string]string{
			"code": "auth-code",
		}
		Expect(w.Result()).NotTo(matchers.HaveAllQueryParams(expectedParams))
	})
})

var _ = Describe("HaveStatusCode", func() {
	It("should match correct status code", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusOK)

		Expect(w.Result()).To(matchers.HaveStatusCode(http.StatusOK))
	})

	It("should match redirect status codes", func() {
		tests := []int{
			http.StatusMovedPermanently,
			http.StatusFound,
			http.StatusSeeOther,
			http.StatusTemporaryRedirect,
		}

		for _, code := range tests {
			w := httptest.NewRecorder()
			w.WriteHeader(code)
			Expect(w.Result()).To(matchers.HaveStatusCode(code))
		}
	})

	It("should fail if status code doesn't match", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusOK)

		Expect(w.Result()).NotTo(matchers.HaveStatusCode(http.StatusNotFound))
	})

	It("should provide helpful error message", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusInternalServerError)

		matcher := matchers.HaveStatusCode(http.StatusOK)
		matches, _ := matcher.Match(w.Result())
		Expect(matches).To(BeFalse())

		failMsg := matcher.FailureMessage(w.Result())
		Expect(failMsg).To(ContainSubstring("500"))
		Expect(failMsg).To(ContainSubstring("200"))
	})
})

var _ = Describe("Matcher edge cases", func() {
	It("should handle nil response gracefully", func() {
		var nilResp *http.Response

		matcher := matchers.HaveStatusCode(http.StatusOK)
		_, err := matcher.Match(nilResp)
		Expect(err).To(HaveOccurred())
	})

	It("should handle non-response types", func() {
		matcher := matchers.HaveStatusCode(http.StatusOK)
		_, err := matcher.Match("not a response")
		Expect(err).To(HaveOccurred())
	})

	It("should provide helpful negated failure messages", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com?code=auth-code&state=state-123")
		w.WriteHeader(http.StatusFound)

		matcher := matchers.HaveOAuth2Redirect("")
		matches, _ := matcher.Match(w.Result())
		Expect(matches).To(BeTrue())

		negMsg := matcher.NegatedFailureMessage(w.Result())
		Expect(negMsg).To(ContainSubstring("NOT"))
		Expect(negMsg).To(ContainSubstring("redirect"))
	})

	It("should show human-readable status names in failure messages", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusBadRequest) // 400

		matcher := matchers.HaveStatusCode(http.StatusConflict) // Expect 409
		matches, _ := matcher.Match(w.Result())
		Expect(matches).To(BeFalse())

		failMsg := matcher.FailureMessage(w.Result())
		Expect(failMsg).To(ContainSubstring("409 Conflict"))
		Expect(failMsg).To(ContainSubstring("400 Bad Request"))
	})
})

var _ = Describe("HaveTokenExchangeSuccess", func() {
	It("should match a valid RFC 8693 token exchange response", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		tokenResponse := map[string]interface{}{
			"access_token":      "test-token-123",
			"token_type":        "Bearer",
			"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			"expires_in":        3600,
		}
		_ = json.NewEncoder(w).Encode(tokenResponse)

		Expect(w.Result()).To(matchers.HaveTokenExchangeSuccess())
	})

	It("should validate specific field values when using With methods", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		tokenResponse := map[string]interface{}{
			"access_token":      "github-token-xyz",
			"token_type":        "Bearer",
			"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			"expires_in":        3600,
		}
		_ = json.NewEncoder(w).Encode(tokenResponse)

		Expect(w.Result()).To(matchers.HaveTokenExchangeSuccess().
			WithAccessToken("github-token-xyz").
			WithTokenType("Bearer").
			WithIssuedTokenType("urn:ietf:params:oauth:token-type:access_token").
			WithExpiresIn(3600))
	})

	It("should fail if Content-Type is not application/json", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("plain text response"))

		Expect(w.Result()).NotTo(matchers.HaveTokenExchangeSuccess())
	})

	It("should fail if required RFC 8693 fields are missing", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Missing issued_token_type field
		tokenResponse := map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
		}
		_ = json.NewEncoder(w).Encode(tokenResponse)

		Expect(w.Result()).NotTo(matchers.HaveTokenExchangeSuccess())
	})

	It("should fail if expected field values do not match", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		tokenResponse := map[string]interface{}{
			"access_token":      "actual-token",
			"token_type":        "Bearer",
			"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
		}
		_ = json.NewEncoder(w).Encode(tokenResponse)

		Expect(w.Result()).NotTo(matchers.HaveTokenExchangeSuccess().
			WithAccessToken("expected-different-token"))
	})

	It("should provide detailed failure messages", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Missing issued_token_type
		tokenResponse := map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
		}
		_ = json.NewEncoder(w).Encode(tokenResponse)

		matcher := matchers.HaveTokenExchangeSuccess()
		matches, _ := matcher.Match(w.Result())
		Expect(matches).To(BeFalse())

		failMsg := matcher.FailureMessage(w.Result())
		Expect(failMsg).To(ContainSubstring("RFC 8693"))
		Expect(failMsg).To(ContainSubstring("issued_token_type"))
	})
})
