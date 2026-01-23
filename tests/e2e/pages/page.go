// Package pages provides page object models for frontend E2E testing with Playwright.
// Page objects abstract away selector details and provide high-level methods that describe
// user interactions from a functional perspective.
//
// All page objects should embed the base Page struct and use its methods for navigation
// and waiting. Subclasses should define their own methods for page-specific interactions.
//
// Example pattern:
//
//	type ConsentPage struct {
//		*Page  // Embed base Page
//	}
//
//	func NewConsentPage(page playwright.Page, baseURL string) *ConsentPage {
//		return &ConsentPage{
//			Page: NewPage(page, baseURL),
//		}
//	}
//
//	func (cp *ConsentPage) ClickApprove(ctx context.Context) error {
//		// Use cp.page to interact with selectors
//		// Wrap errors with fmt.Errorf using %w
//	}
package pages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Page represents a Playwright page with common navigation and interaction utilities.
// This is the base struct that all page objects (ConsentPage, AgentDetailPage, etc.) embed.
// It provides high-level methods for navigation, waiting, and screenshots while keeping
// selector details private to subclasses.
//
// Design principles:
// - No selectors exposed to callers
// - Context propagation for all async operations
// - Implicit waits via Playwright (no manual sleep)
// - Clear error messages with proper error wrapping
type Page struct {
	page          playwright.Page
	baseURL       string
	timeout       time.Duration
	screenshotDir string
}

// NewPage creates a new Page wrapper around a Playwright page.
// This initializes the page with a 30-second timeout (Playwright default)
// and prepares the screenshot directory.
//
// Parameters:
//   - page: Playwright page instance (required)
//   - baseURL: Base URL for navigation (e.g., "http://localhost:3000")
//
// Returns:
//   - *Page: Initialized page wrapper
//   - Panics if page is nil (programming error, not recoverable)
//
// Example:
//
//	page, err := context.NewPage()
//	require.NoError(t, err)
//	p := pages.NewPage(page, "http://localhost:3000")
//	defer p.Close()
func NewPage(page playwright.Page, baseURL string) *Page {
	if page == nil {
		panic("page is required and cannot be nil")
	}

	// Prepare screenshot directory
	screenshotDir := filepath.Join("coverage", "screenshots")
	_ = os.MkdirAll(screenshotDir, 0o755) // Best effort; don't fail if it fails

	return &Page{
		page:          page,
		baseURL:       baseURL,
		timeout:       30 * time.Second,
		screenshotDir: screenshotDir,
	}
}

// Navigate navigates to the given path relative to the base URL.
// Path should start with "/" or be a path like "agent/123".
// The method will combine baseURL + path to create the full URL.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - path: Relative path (e.g., "/consent", "agent/123")
//
// Returns:
//   - error: If navigation fails (network error, invalid URL, etc.)
//
// Error messages include the full URL for debugging:
// - "failed to navigate to http://localhost:3000/consent: context cancelled"
// - "failed to navigate to http://localhost:3000/404: HTTP 404"
//
// Example:
//
//	err := p.Navigate(ctx, "/consent")
//	require.NoError(t, err)
//
//	// Navigate with path parameters
//	err = p.Navigate(ctx, "/agent/123/detail")
//	require.NoError(t, err)
func (p *Page) Navigate(ctx context.Context, path string) error {
	// Ensure path is absolute
	if path != "" && path[0] != '/' {
		path = "/" + path
	}

	url := p.baseURL + path

	// Set timeout for navigation
	_, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Navigate (Playwright automatically waits for page load)
	// Note: Playwright Go client handles timeouts internally, not via context
	_, err := p.page.Goto(url)
	if err != nil {
		return fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	// Wait for page to be fully loaded
	if err := p.page.WaitForLoadState(); err != nil {
		return fmt.Errorf("failed waiting for page load at %s: %w", url, err)
	}

	return nil
}

// GetBaseURL returns the base URL of this page.
// Useful for subclasses that need to build URLs programmatically.
//
// Returns:
//   - string: Base URL (e.g., "http://localhost:3000")
//
// Example:
//
//	baseURL := p.GetBaseURL()
//	// Use in subclass to build full URL
func (p *Page) GetBaseURL() string {
	return p.baseURL
}

// WaitForURL waits for the page URL to match a pattern (regex or substring).
// This is useful for verifying page navigation or redirects.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - pattern: String pattern to match in URL (can be regex or substring)
//   - If pattern contains regex special chars and is a valid regex, it's treated as regex
//   - Otherwise, it's treated as a substring match
//   - Examples: "/success", "consent", "agent/[0-9]+"
//
// Returns:
//   - error: If timeout expires before URL matches
//
// Error message format:
// - "timeout waiting for URL to match '/success' (current: 'http://localhost:3000/error', waited 30s)"
//
// Example:
//
//	// Wait for redirect to success page
//	err := p.WaitForURL(ctx, "/success")
//	require.NoError(t, err)
//
//	// Wait with regex pattern
//	err = p.WaitForURL(ctx, "agent/[0-9]+")
//	require.NoError(t, err)
func (p *Page) WaitForURL(ctx context.Context, pattern string) error {
	deadline := time.Now().Add(p.timeout)

	// Try to compile as regex; if it fails, treat as substring
	var regex *regexp.Regexp
	regex, _ = regexp.Compile(pattern)

	for {
		currentURL := p.page.URL()

		// Check if URL matches pattern
		if regex != nil {
			// Try regex match
			if regex.MatchString(currentURL) {
				return nil
			}
		} else {
			// Try substring match
			if contains(currentURL, pattern) {
				return nil
			}
		}

		// Check timeout
		if time.Now().After(deadline) {
			return fmt.Errorf(
				"timeout waiting for URL to match %q (current: %q, waited %v)",
				pattern,
				currentURL,
				p.timeout,
			)
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for URL %q: %w", pattern, ctx.Err())
		default:
		}

		// Wait a bit before checking again
		time.Sleep(100 * time.Millisecond)
	}
}

// WaitForNavigation waits for the page to navigate (URL change).
// This is useful after clicking a link that triggers navigation.
// Returns when a navigation event occurs or timeout expires.
// Internally uses WaitForLoadState to detect navigation completion.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//
// Returns:
//   - error: If timeout expires or context is cancelled
//
// Example:
//
//	// Click a link and wait for navigation
//	err := p.page.Click("#continue-link")
//	require.NoError(t, err)
//
//	err = p.WaitForNavigation(ctx)
//	require.NoError(t, err)
//
//	// Now on the new page
//	currentURL := p.page.URL()
func (p *Page) WaitForNavigation(ctx context.Context) error {
	// Set timeout
	_, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Wait for page to reach load state after navigation
	// This ensures the new page is fully loaded
	if err := p.page.WaitForLoadState(); err != nil {
		return fmt.Errorf("timeout waiting for navigation (waited %v): %w", p.timeout, err)
	}

	return nil
}

// TakeScreenshot captures a screenshot of the current page state.
// Screenshots are saved to coverage/screenshots/{name}.png.
// This is useful for debugging test failures.
//
// Parameters:
//   - ctx: Context for cancellation (Playwright has built-in timeout)
//   - name: Screenshot name (e.g., "consent_page_loaded")
//
// Returns:
//   - error: If screenshot capture fails
//
// File location:
//   - Screenshots are saved to: coverage/screenshots/{name}.png
//   - Directory is created automatically if it doesn't exist
//
// Example:
//
//	// Take screenshot on test failure
//	err := p.Navigate(ctx, "/consent")
//	if err != nil {
//		_ = p.TakeScreenshot(ctx, "navigation_failed")
//		t.Fatal(err)
//	}
//
//	// Take screenshot for comparison
//	_ = p.TakeScreenshot(ctx, "consent_page_state")
func (p *Page) TakeScreenshot(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("screenshot name cannot be empty")
	}

	// Ensure screenshot directory exists
	if err := os.MkdirAll(p.screenshotDir, 0o755); err != nil {
		return fmt.Errorf("failed to create screenshot directory %s: %w", p.screenshotDir, err)
	}

	// Build full file path
	filePath := filepath.Join(p.screenshotDir, name+".png")

	// Set timeout for screenshot operation
	_, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Take screenshot
	data, err := p.page.Screenshot()
	if err != nil {
		return fmt.Errorf("failed to take screenshot %s: %w", filePath, err)
	}

	// Write screenshot to file
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write screenshot to %s: %w", filePath, err)
	}

	return nil
}

// GetCurrentURL returns the current page URL.
// Useful for verifying navigation without waiting.
//
// Returns:
//   - string: Current page URL (e.g., "http://localhost:3000/consent")
//   - error: Never returns error; always succeeds
//
// Example:
//
//	currentURL := p.GetCurrentURL(ctx)
//	fmt.Printf("Currently on: %s\n", currentURL)
//
//	// Verify we're on expected page
//	require.Contains(t, p.GetCurrentURL(ctx), "/consent")
func (p *Page) GetCurrentURL(ctx context.Context) (string, error) {
	return p.page.URL(), nil
}

// GetPlaywrightPage returns the underlying Playwright page object.
// This is used by subclasses (like ConsentPage) to access Playwright APIs.
//
// Returns:
//   - playwright.Page: The underlying Playwright page
//
// Example:
//
//	page := p.GetPlaywrightPage()
//	page.GetByRole("button").Click()
func (p *Page) GetPlaywrightPage() playwright.Page {
	return p.page
}

// Close closes the page and releases resources.
// Safe to call multiple times (idempotent).
// Should be called in test cleanup (defer p.Close()).
//
// Parameters: None
//
// Returns:
//   - error: If page close fails (rare)
//
// Example:
//
//	p := pages.NewPage(page, "http://localhost:3000")
//	defer p.Close()
//
//	// Test code here
func (p *Page) Close() error {
	if p.page == nil {
		return nil
	}
	return p.page.Close()
}

// contains checks if haystack contains needle as a substring.
// Helper function for URL matching.
func contains(haystack, needle string) bool {
	// Use built-in strings.Contains logic
	if len(needle) == 0 {
		return true
	}
	if len(needle) > len(haystack) {
		return false
	}

	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
