// Package helpers provides HTTP utilities for E2E testing.
package helpers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
)

func TestHTTPHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Helpers Suite")
}

var _ = Describe("ExtractRedirectURL", func() {
	It("should extract valid redirect URL from Location header", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "https://callback.example.com/redirect?code=123&state=abc")
		w.WriteHeader(http.StatusFound)

		extractedURL, err := helpers.ExtractRedirectURL(w.Result())

		Expect(err).NotTo(HaveOccurred())
		Expect(extractedURL).NotTo(BeNil())
		Expect(extractedURL.Scheme).To(Equal("https"))
		Expect(extractedURL.Host).To(Equal("callback.example.com"))
		Expect(extractedURL.Path).To(Equal("/redirect"))
	})

	It("should fail if response is nil", func() {
		var nilResp *http.Response
		_, err := helpers.ExtractRedirectURL(nilResp)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("nil"))
	})

	It("should fail if Location header is missing", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusOK)

		_, err := helpers.ExtractRedirectURL(w.Result())

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Location"))
	})

	It("should fail if Location header contains invalid URL", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Location", "not a valid url ://")
		w.WriteHeader(http.StatusFound)

		_, err := helpers.ExtractRedirectURL(w.Result())

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("parse"))
	})
})

var _ = Describe("ParseOAuth2ErrorFromRedirect", func() {
	It("should extract error and description from redirect URL", func() {
		// Parse the URL manually for test
		testURL := "https://callback.example.com?error=access_denied&error_description=User+denied"
		parsedURL, _ := parseTestURL(testURL)

		errorCode, errorDesc := helpers.ParseOAuth2ErrorFromRedirect(parsedURL)

		Expect(errorCode).To(Equal("access_denied"))
		Expect(errorDesc).To(Equal("User denied"))
	})

	It("should return empty strings if no error parameter", func() {
		parsedURL, _ := parseTestURL("https://callback.example.com?code=auth-code&state=state-123")

		errorCode, errorDesc := helpers.ParseOAuth2ErrorFromRedirect(parsedURL)

		Expect(errorCode).To(Equal(""))
		Expect(errorDesc).To(Equal(""))
	})

	It("should handle nil URL gracefully", func() {
		errorCode, errorDesc := helpers.ParseOAuth2ErrorFromRedirect(nil)

		Expect(errorCode).To(Equal(""))
		Expect(errorDesc).To(Equal(""))
	})

	It("should extract error without description", func() {
		parsedURL, _ := parseTestURL("https://callback.example.com?error=invalid_request")

		errorCode, errorDesc := helpers.ParseOAuth2ErrorFromRedirect(parsedURL)

		Expect(errorCode).To(Equal("invalid_request"))
		Expect(errorDesc).To(Equal(""))
	})
})

var _ = Describe("CreateAuthorizationRequest", func() {
	It("should create valid OAuth2 authorization request URL", func() {
		baseURL := "https://broker.example.com/oauth2/authorize"
		clientID := "client-123"
		redirectURI := "https://app.example.com/callback"
		state := "state-abc"

		requestURL := helpers.CreateAuthorizationRequest(baseURL, clientID, redirectURI, state)

		Expect(requestURL).To(ContainSubstring("client_id=client-123"))
		Expect(requestURL).To(ContainSubstring("redirect_uri="))
		Expect(requestURL).To(ContainSubstring("response_type=code"))
		Expect(requestURL).To(ContainSubstring("state=state-abc"))
	})

	It("should URL-encode redirect URI", func() {
		baseURL := "https://broker.example.com/oauth2/authorize"
		redirectURI := "https://app.example.com/callback?param=value"

		requestURL := helpers.CreateAuthorizationRequest(baseURL, "client-1", redirectURI, "state")

		// Should be URL-encoded
		Expect(requestURL).To(ContainSubstring("redirect_uri="))
		// Should not contain unencoded ? in the parameter value itself
		Expect(strings.Count(requestURL, "?")).To(Equal(1))
	})

	It("should include all required OAuth2 parameters", func() {
		requestURL := helpers.CreateAuthorizationRequest(
			"https://auth.example.com/authorize",
			"app-client",
			"https://app.example.com/callback",
			"random-state-123",
		)

		Expect(requestURL).To(ContainSubstring("client_id="))
		Expect(requestURL).To(ContainSubstring("redirect_uri="))
		Expect(requestURL).To(ContainSubstring("response_type=code"))
		Expect(requestURL).To(ContainSubstring("state="))
	})
})

var _ = Describe("ReadResponseBody", func() {
	It("should read response body successfully", func() {
		w := httptest.NewRecorder()
		_, _ = w.WriteString("test response body")
		w.WriteHeader(http.StatusOK)

		body, err := helpers.ReadResponseBody(w.Result())

		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(ContainSubstring("test response body"))
	})

	It("should handle empty response body", func() {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusNoContent)

		body, err := helpers.ReadResponseBody(w.Result())

		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(Equal(""))
	})

	It("should fail if response is nil", func() {
		var nilResp *http.Response
		_, err := helpers.ReadResponseBody(nilResp)

		Expect(err).To(HaveOccurred())
	})

	It("should read JSON response body", func() {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.WriteString(`{"error":"invalid_request","error_description":"Bad request"}`)
		w.WriteHeader(http.StatusBadRequest)

		body, err := helpers.ReadResponseBody(w.Result())

		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(ContainSubstring("error"))
		Expect(body).To(ContainSubstring("invalid_request"))
	})
})

var _ = Describe("ExtractQueryParam", func() {
	It("should extract single query parameter", func() {
		parsedURL, _ := parseTestURL("https://example.com?code=auth-code&state=state-123")

		code := helpers.ExtractQueryParam(parsedURL, "code")

		Expect(code).To(Equal("auth-code"))
	})

	It("should return empty string if parameter not found", func() {
		parsedURL, _ := parseTestURL("https://example.com?code=auth-code")

		state := helpers.ExtractQueryParam(parsedURL, "state")

		Expect(state).To(Equal(""))
	})

	It("should handle nil URL gracefully", func() {
		param := helpers.ExtractQueryParam(nil, "code")

		Expect(param).To(Equal(""))
	})
})

var _ = Describe("ExtractAllQueryParams", func() {
	It("should extract all query parameters into map", func() {
		parsedURL, _ := parseTestURL("https://example.com?code=auth-code&state=state-123&session=sess-456")

		params := helpers.ExtractAllQueryParams(parsedURL)

		Expect(params).To(HaveKeyWithValue("code", "auth-code"))
		Expect(params).To(HaveKeyWithValue("state", "state-123"))
		Expect(params).To(HaveKeyWithValue("session", "sess-456"))
	})

	It("should return empty map for nil URL", func() {
		params := helpers.ExtractAllQueryParams(nil)

		Expect(params).To(HaveLen(0))
	})

	It("should handle URL with no query parameters", func() {
		parsedURL, _ := parseTestURL("https://example.com/path")

		params := helpers.ExtractAllQueryParams(parsedURL)

		Expect(params).To(HaveLen(0))
	})
})

var _ = Describe("HTTP Status Code Checks", func() {
	It("should identify redirect status codes", func() {
		Expect(helpers.IsHTTPRedirect(http.StatusMovedPermanently)).To(BeTrue())
		Expect(helpers.IsHTTPRedirect(http.StatusFound)).To(BeTrue())
		Expect(helpers.IsHTTPRedirect(http.StatusSeeOther)).To(BeTrue())
		Expect(helpers.IsHTTPRedirect(http.StatusTemporaryRedirect)).To(BeTrue())

		Expect(helpers.IsHTTPRedirect(http.StatusOK)).To(BeFalse())
		Expect(helpers.IsHTTPRedirect(http.StatusBadRequest)).To(BeFalse())
	})

	It("should identify client error status codes", func() {
		Expect(helpers.IsHTTPClientError(http.StatusBadRequest)).To(BeTrue())
		Expect(helpers.IsHTTPClientError(http.StatusUnauthorized)).To(BeTrue())
		Expect(helpers.IsHTTPClientError(http.StatusNotFound)).To(BeTrue())

		Expect(helpers.IsHTTPClientError(http.StatusOK)).To(BeFalse())
		Expect(helpers.IsHTTPClientError(http.StatusFound)).To(BeFalse())
		Expect(helpers.IsHTTPClientError(http.StatusInternalServerError)).To(BeFalse())
	})

	It("should identify server error status codes", func() {
		Expect(helpers.IsHTTPServerError(http.StatusInternalServerError)).To(BeTrue())
		Expect(helpers.IsHTTPServerError(http.StatusBadGateway)).To(BeTrue())
		Expect(helpers.IsHTTPServerError(http.StatusServiceUnavailable)).To(BeTrue())

		Expect(helpers.IsHTTPServerError(http.StatusOK)).To(BeFalse())
		Expect(helpers.IsHTTPServerError(http.StatusBadRequest)).To(BeFalse())
	})
})

var _ = Describe("ParseJSONError", func() {
	It("should parse JSON error response", func() {
		jsonBody := `{"error":"invalid_request","error_description":"Missing parameter"}`

		errorCode, errorDesc := helpers.ParseJSONError(jsonBody)

		Expect(errorCode).To(Equal("invalid_request"))
		Expect(errorDesc).To(Equal("Missing parameter"))
	})

	It("should handle missing error_description", func() {
		jsonBody := `{"error":"access_denied"}`

		errorCode, errorDesc := helpers.ParseJSONError(jsonBody)

		Expect(errorCode).To(Equal("access_denied"))
		Expect(errorDesc).To(Equal(""))
	})

	It("should handle empty JSON", func() {
		jsonBody := `{}`

		errorCode, errorDesc := helpers.ParseJSONError(jsonBody)

		Expect(errorCode).To(Equal(""))
		Expect(errorDesc).To(Equal(""))
	})
})

// Helper function to parse test URLs
func parseTestURL(urlStr string) (*url.URL, error) {
	return url.Parse(urlStr)
}
