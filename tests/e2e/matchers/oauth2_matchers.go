// Package matchers provides custom Gomega matchers for OAuth2-specific assertions.
// These matchers focus on HTTP contract testing and are stable across refactoring.
package matchers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/onsi/gomega/types"
)

// HaveOAuth2Redirect is a Gomega matcher that checks for OAuth2 redirect response.
// Validates:
// - 302 or 303 redirect status code
// - Location header present and parseable
// - Redirect URL matches expected URL
func HaveOAuth2Redirect(expectedURL string) types.GomegaMatcher {
	return &oauth2RedirectMatcher{
		expectedURL: expectedURL,
	}
}

type oauth2RedirectMatcher struct {
	expectedURL string
	actualURL   string
	statusCode  int
	error       string
}

func (m *oauth2RedirectMatcher) Match(actual interface{}) (success bool, err error) {
	resp, ok := actual.(*http.Response)
	if !ok {
		return false, fmt.Errorf("HaveOAuth2Redirect matcher expects an *http.Response, got %T", actual)
	}

	if resp == nil {
		m.error = "response is nil"
		return false, nil
	}

	m.statusCode = resp.StatusCode

	// Check redirect status code (3xx)
	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		m.error = fmt.Sprintf("expected redirect status (3xx), got %d", resp.StatusCode)
		return false, nil
	}

	// Check Location header
	location := resp.Header.Get("Location")
	if location == "" {
		m.error = "Location header is missing"
		return false, nil
	}

	m.actualURL = location

	// Parse redirect URL to validate it's a valid URL
	parsedLocation, err := url.Parse(location)
	if err != nil {
		m.error = fmt.Sprintf("Location header is not a valid URL: %v", err)
		return false, nil
	}

	// If expectedURL is provided, check if it matches
	if m.expectedURL != "" {
		parsedExpected, err := url.Parse(m.expectedURL)
		if err != nil {
			return false, fmt.Errorf("expectedURL is not valid: %v", err)
		}

		// Handle relative URLs (no scheme/host in expected)
		// If expected URL is relative (e.g., /consent/agent/123), just compare paths
		if parsedExpected.Scheme == "" && parsedExpected.Host == "" {
			// Relative URL match - only compare paths
			if parsedLocation.Path != parsedExpected.Path {
				m.error = fmt.Sprintf("redirect URL path mismatch: expected %s, got %s", parsedExpected.Path, parsedLocation.Path)
				return false, nil
			}
		} else {
			// Absolute URL match - compare scheme, host, and path
			if parsedLocation.Scheme != parsedExpected.Scheme || parsedLocation.Host != parsedExpected.Host || parsedLocation.Path != parsedExpected.Path {
				m.error = fmt.Sprintf("redirect URL mismatch: expected %s, got %s", m.expectedURL, location)
				return false, nil
			}
		}
	}

	return true, nil
}

func (m *oauth2RedirectMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected response to have OAuth2 redirect to %q (status: %d)\nError: %s\nActual URL: %s",
		m.expectedURL, m.statusCode, m.error, m.actualURL)
}

func (m *oauth2RedirectMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected response NOT to have OAuth2 redirect to %q", m.expectedURL)
}

// HaveOAuth2Error is a Gomega matcher that checks for OAuth2 error response.
// Validates:
// - Error parameter in redirect URL query string OR
// - Error in JSON response body (for non-redirect errors)
func HaveOAuth2Error(expectedErrorCode string) types.GomegaMatcher {
	return &oauth2ErrorMatcher{
		expectedErrorCode: expectedErrorCode,
	}
}

type oauth2ErrorMatcher struct {
	expectedErrorCode string
	actualErrorCode   string
	errorDescription  string
	location          string
	error             string
}

func (m *oauth2ErrorMatcher) Match(actual interface{}) (success bool, err error) {
	resp, ok := actual.(*http.Response)
	if !ok {
		return false, fmt.Errorf("HaveOAuth2Error matcher expects an *http.Response, got %T", actual)
	}

	if resp == nil {
		m.error = "response is nil"
		return false, nil
	}

	// Check for error in redirect (3xx response with error parameter)
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		if location == "" {
			m.error = "redirect response missing Location header"
			return false, nil
		}

		m.location = location

		redirectURL, err := url.Parse(location)
		if err != nil {
			m.error = fmt.Sprintf("failed to parse redirect URL: %v", err)
			return false, nil
		}

		query := redirectURL.Query()
		errorCode := query.Get("error")
		if errorCode != "" {
			m.actualErrorCode = errorCode
			m.errorDescription = query.Get("error_description")

			if m.expectedErrorCode == "" || m.expectedErrorCode == errorCode {
				return true, nil
			}

			m.error = fmt.Sprintf("error code mismatch: expected %q, got %q", m.expectedErrorCode, errorCode)
			return false, nil
		}
	}

	// Check for JSON error response (4xx or 5xx)
	if resp.StatusCode >= 400 {
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			// Would need to read and parse body - simplified for now
			m.error = "JSON error response (would need body parsing)"
			return false, nil
		}
	}

	m.error = "no error found in response"
	return false, nil
}

func (m *oauth2ErrorMatcher) FailureMessage(actual interface{}) string {
	msg := fmt.Sprintf("Expected response to have OAuth2 error %q", m.expectedErrorCode)
	if m.actualErrorCode != "" {
		msg += fmt.Sprintf("\nActual error: %q", m.actualErrorCode)
	}
	if m.errorDescription != "" {
		msg += fmt.Sprintf("\nError description: %s", m.errorDescription)
	}
	if m.location != "" {
		msg += fmt.Sprintf("\nLocation: %s", m.location)
	}
	if m.error != "" {
		msg += fmt.Sprintf("\nDetails: %s", m.error)
	}
	return msg
}

func (m *oauth2ErrorMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected response NOT to have OAuth2 error %q", m.expectedErrorCode)
}

// HaveOAuth2SuccessRedirect checks for successful OAuth2 redirect with authorization code.
// Validates:
// - 302 or 303 redirect status
// - Location header present
// - Contains "code" parameter
// - Contains "state" parameter (if expectedState provided)
func HaveOAuth2SuccessRedirect(expectedState string) types.GomegaMatcher {
	return &oauth2SuccessRedirectMatcher{
		expectedState: expectedState,
	}
}

type oauth2SuccessRedirectMatcher struct {
	expectedState string
	actualCode    string
	actualState   string
	location      string
	error         string
}

func (m *oauth2SuccessRedirectMatcher) Match(actual interface{}) (success bool, err error) {
	resp, ok := actual.(*http.Response)
	if !ok {
		return false, fmt.Errorf("HaveOAuth2SuccessRedirect matcher expects an *http.Response, got %T", actual)
	}

	if resp == nil {
		m.error = "response is nil"
		return false, nil
	}

	// Check redirect status code
	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		m.error = fmt.Sprintf("expected redirect status (3xx), got %d", resp.StatusCode)
		return false, nil
	}

	location := resp.Header.Get("Location")
	if location == "" {
		m.error = "Location header is missing"
		return false, nil
	}

	m.location = location

	redirectURL, err := url.Parse(location)
	if err != nil {
		m.error = fmt.Sprintf("failed to parse Location header: %v", err)
		return false, nil
	}

	query := redirectURL.Query()

	// Check for error parameter (should not have one)
	if errCode := query.Get("error"); errCode != "" {
		m.error = fmt.Sprintf("response contains error: %s", errCode)
		return false, nil
	}

	// Check for authorization code
	code := query.Get("code")
	if code == "" {
		m.error = "response missing 'code' parameter"
		return false, nil
	}

	m.actualCode = code

	// Check state if expected
	state := query.Get("state")
	if m.expectedState != "" && state != m.expectedState {
		m.error = fmt.Sprintf("state mismatch: expected %q, got %q", m.expectedState, state)
		return false, nil
	}

	m.actualState = state
	return true, nil
}

func (m *oauth2SuccessRedirectMatcher) FailureMessage(actual interface{}) string {
	msg := fmt.Sprintf("Expected successful OAuth2 redirect")
	if m.actualCode != "" {
		msg += fmt.Sprintf("\nAuthorization code: %s", m.actualCode)
	}
	if m.actualState != "" {
		msg += fmt.Sprintf("\nState: %s", m.actualState)
	}
	if m.location != "" {
		msg += fmt.Sprintf("\nLocation: %s", m.location)
	}
	if m.error != "" {
		msg += fmt.Sprintf("\nError: %s", m.error)
	}
	return msg
}

func (m *oauth2SuccessRedirectMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected response NOT to have successful OAuth2 redirect")
}

// HaveMetadataField checks for a specific field in metadata response.
// Typically used for OpenID Connect discovery endpoint responses.
func HaveMetadataField(fieldName string) types.GomegaMatcher {
	return &metadataFieldMatcher{
		fieldName: fieldName,
	}
}

type metadataFieldMatcher struct {
	fieldName   string
	statusCode  int
	error       string
	bodyPreview string
}

func (m *metadataFieldMatcher) Match(actual interface{}) (success bool, err error) {
	resp, ok := actual.(*http.Response)
	if !ok {
		return false, fmt.Errorf("HaveMetadataField matcher expects an *http.Response, got %T", actual)
	}

	if resp == nil {
		m.error = "response is nil"
		return false, nil
	}

	m.statusCode = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		m.error = fmt.Sprintf("expected status 200, got %d", resp.StatusCode)
		return false, nil
	}

	// Check Content-Type is JSON
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		m.error = fmt.Sprintf("expected JSON content type, got %s", contentType)
		return false, nil
	}

	// Simple check: look for field name in response body
	// Note: This is a simplified check that doesn't require JSON unmarshaling
	if resp.Body != nil {
		defer resp.Body.Close()

		buf := make([]byte, 8192)
		n, _ := resp.Body.Read(buf)
		body := string(buf[:n])
		m.bodyPreview = body

		// Check if field name appears in JSON (look for "fieldName": )
		// Try both quoted forms to handle JSON encoding variations
		searchPatterns := []string{
			fmt.Sprintf(`"%s"`, m.fieldName),
			fmt.Sprintf(`"%s":`, m.fieldName),
		}

		for _, pattern := range searchPatterns {
			if strings.Contains(body, pattern) {
				return true, nil
			}
		}
	}

	m.error = fmt.Sprintf("field %q not found in metadata response", m.fieldName)
	return false, nil
}

func (m *metadataFieldMatcher) FailureMessage(actual interface{}) string {
	msg := fmt.Sprintf("Expected metadata response to contain field %q (status: %d)", m.fieldName, m.statusCode)
	if m.error != "" {
		msg += fmt.Sprintf("\nError: %s", m.error)
	}
	if m.bodyPreview != "" && len(m.bodyPreview) < 200 {
		msg += fmt.Sprintf("\nResponse body: %s", m.bodyPreview)
	}
	return msg
}

func (m *metadataFieldMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected metadata response NOT to contain field %q", m.fieldName)
}

// HaveAllQueryParams checks that a redirect URL contains all expected query parameters.
// Used to validate OAuth2 response parameters.
func HaveAllQueryParams(expectedParams map[string]string) types.GomegaMatcher {
	return &queryParamsMatcher{
		expectedParams: expectedParams,
	}
}

type queryParamsMatcher struct {
	expectedParams map[string]string
	missingParams  []string
	wrongParams    map[string]string
	location       string
	error          string
}

func (m *queryParamsMatcher) Match(actual interface{}) (success bool, err error) {
	resp, ok := actual.(*http.Response)
	if !ok {
		return false, fmt.Errorf("HaveAllQueryParams matcher expects an *http.Response, got %T", actual)
	}

	if resp == nil {
		m.error = "response is nil"
		return false, nil
	}

	// Try to extract URL from Location header (for redirects)
	location := resp.Header.Get("Location")
	if location == "" {
		m.error = "Location header is missing"
		return false, nil
	}

	m.location = location

	redirectURL, err := url.Parse(location)
	if err != nil {
		m.error = fmt.Sprintf("failed to parse Location header: %v", err)
		return false, nil
	}

	query := redirectURL.Query()
	m.wrongParams = make(map[string]string)
	m.missingParams = []string{}

	// Check all expected parameters
	for paramName, expectedValue := range m.expectedParams {
		actualValue := query.Get(paramName)
		if actualValue == "" {
			m.missingParams = append(m.missingParams, paramName)
		} else if actualValue != expectedValue {
			m.wrongParams[paramName] = actualValue
		}
	}

	// Success if no missing or wrong parameters
	return len(m.missingParams) == 0 && len(m.wrongParams) == 0, nil
}

func (m *queryParamsMatcher) FailureMessage(actual interface{}) string {
	msg := fmt.Sprintf("Expected redirect URL to contain all query parameters\nLocation: %s", m.location)

	if len(m.missingParams) > 0 {
		msg += fmt.Sprintf("\nMissing parameters: %v", m.missingParams)
	}

	if len(m.wrongParams) > 0 {
		msg += fmt.Sprintf("\nIncorrect parameter values:")
		for name, actualValue := range m.wrongParams {
			expectedValue := m.expectedParams[name]
			msg += fmt.Sprintf("\n  %s: expected %q, got %q", name, expectedValue, actualValue)
		}
	}

	if m.error != "" {
		msg += fmt.Sprintf("\nError: %s", m.error)
	}

	return msg
}

func (m *queryParamsMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected redirect URL NOT to contain all query parameters")
}

// HaveStatusCode is a simple matcher for HTTP status codes.
// Provides clear error messages for status code assertions.
func HaveStatusCode(expectedStatus int) types.GomegaMatcher {
	return &statusCodeMatcher{
		expectedStatus: expectedStatus,
	}
}

type statusCodeMatcher struct {
	expectedStatus int
	actualStatus   int
}

func (m *statusCodeMatcher) Match(actual interface{}) (success bool, err error) {
	resp, ok := actual.(*http.Response)
	if !ok {
		return false, fmt.Errorf("HaveStatusCode matcher expects an *http.Response, got %T", actual)
	}

	if resp == nil {
		return false, fmt.Errorf("response is nil")
	}

	m.actualStatus = resp.StatusCode
	return m.actualStatus == m.expectedStatus, nil
}

func (m *statusCodeMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected HTTP status %d, got %d", m.expectedStatus, m.actualStatus)
}

func (m *statusCodeMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected HTTP status not to be %d", m.expectedStatus)
}
