// Package bootstrap provides test infrastructure for E2E testing.
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/mxschmitt/playwright-go"
)

// PlaywrightHelper manages Playwright browser lifecycle for frontend E2E tests.
// It provides isolated browser contexts for test isolation and ensures the frontend
// is accessible before running tests.
//
// Usage pattern:
//
//	helper, err := NewPlaywrightHelper(logger)
//	require.NoError(t, err)
//	defer helper.Close(ctx)
//
//	// Initialize browser (typically in BeforeSuite)
//	err = helper.InitBrowser(ctx)
//	require.NoError(t, err)
//
//	// Create fresh context per test (typically in BeforeEach)
//	context, err := helper.CreateContext(ctx)
//	require.NoError(t, err)
//	defer context.Close()
type PlaywrightHelper struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	baseURL string
	logger  *slog.Logger
}

// NewPlaywrightHelper creates a new Playwright helper instance.
// Browser initialization is deferred to InitBrowser() to allow independent
// error handling and logging at startup.
//
// Parameters:
//   - logger: Structured logger for operations
//
// Returns:
//   - *PlaywrightHelper: Initialized helper (browser not yet running)
//   - error: If logger is nil
//
// Example:
//
//	helper, err := NewPlaywrightHelper(logger)
//	require.NoError(t, err)
func NewPlaywrightHelper(logger *slog.Logger) (*PlaywrightHelper, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &PlaywrightHelper{
		logger: logger,
	}, nil
}

// InitBrowser initializes the Playwright browser (called from BeforeSuite).
// This creates the Playwright instance and launches a Chromium browser.
//
// Parameters:
//   - ctx: Context for cancellation (typically context.Background in tests)
//
// Returns:
//   - error: If Playwright initialization fails or browser launch fails
//
// Error handling:
//   - If Playwright binary not found, suggests installation command
//   - Includes version info in log for debugging
//
// Example:
//
//	err := helper.InitBrowser(context.Background())
//	require.NoError(t, err)
func (ph *PlaywrightHelper) InitBrowser(ctx context.Context) error {
	// Run Playwright (installs browsers if needed)
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to initialize Playwright: %w. Try: playwright install", err)
	}
	ph.pw = pw

	// Determine headless mode from environment
	headless := true // default
	if h := os.Getenv("HEADLESS"); h != "" && h != "true" {
		headless = false
	}

	// Launch Chromium browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		_ = pw.Stop()
		return fmt.Errorf("failed to launch Chromium browser: %w", err)
	}
	ph.browser = browser

	// Determine frontend URL from environment
	ph.baseURL = ph.GetBaseURL()

	ph.logger.Info(
		"Playwright browser initialized",
		"headless", headless,
		"frontend_mode", os.Getenv("E2E_FRONTEND_MODE"),
		"frontend_url", ph.baseURL,
	)

	// Verify frontend is accessible before running tests
	if err := ph.VerifyFrontendAccessible(ctx); err != nil {
		_ = ph.Close(ctx)
		return err
	}

	return nil
}

// CreateContext creates a fresh isolated browser context for test isolation.
// Each test should get its own context to ensure complete isolation.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - playwright.BrowserContext: Fresh context for this test
//   - error: If context creation fails
//
// Isolation guarantees:
//   - Each context has separate cookies, storage, cache
//   - Tests can run in parallel without interference
//   - Context should be closed after test (using defer)
//
// Example:
//
//	context, err := helper.CreateContext(context.Background())
//	require.NoError(t, err)
//	defer context.Close()
//
//	page, err := context.NewPage()
//	require.NoError(t, err)
//	defer page.Close()
func (ph *PlaywrightHelper) CreateContext(ctx context.Context) (playwright.BrowserContext, error) {
	if ph.browser == nil {
		return nil, fmt.Errorf("browser not initialized, call InitBrowser first")
	}

	context, err := ph.browser.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}

	return context, nil
}

// GetBaseURL returns the frontend base URL based on E2E_FRONTEND_MODE environment variable.
// This allows tests to run against both development (Vite) and built frontend.
//
// Returns:
//   - string: Frontend base URL
//
// Mode selection:
//   - E2E_FRONTEND_MODE="dev": Returns http://localhost:3000 (Vite dev server)
//   - E2E_FRONTEND_MODE="built" or unset: Returns http://localhost:8000/consent (built frontend)
//
// Example:
//
//	url := helper.GetBaseURL()
//	// In dev mode: "http://localhost:3000"
//	// In built mode: "http://localhost:8000/consent"
func (ph *PlaywrightHelper) GetBaseURL() string {
	mode := os.Getenv("E2E_FRONTEND_MODE")

	if mode == "dev" {
		return "http://localhost:3000/"
	}

	// Default to built frontend served from Go backend
	return "http://localhost:8000/"
}

// VerifyFrontendAccessible verifies the frontend is accessible with retries.
// This runs before tests to fail fast if frontend isn't ready.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - error: If frontend unreachable after all retries
//
// Retry strategy:
//   - 3 attempts with 1 second delay between attempts
//   - Clear error message showing URL, mode, and troubleshooting steps
//   - Logs successful verification
//
// Example:
//
//	err := helper.VerifyFrontendAccessible(context.Background())
//	require.NoError(t, err)
func (ph *PlaywrightHelper) VerifyFrontendAccessible(ctx context.Context) error {
	if ph.browser == nil {
		return fmt.Errorf("browser not initialized, call InitBrowser first")
	}

	const (
		maxRetries = 3
		retryDelay = 1 * time.Second
	)

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Create temporary context for verification
		tempCtx, err := ph.browser.NewContext()
		if err != nil {
			lastErr = fmt.Errorf("failed to create temp context: %w", err)
			continue
		}

		// Create temporary page
		page, err := tempCtx.NewPage()
		if err != nil {
			_ = tempCtx.Close()
			lastErr = fmt.Errorf("failed to create temp page: %w", err)
			continue
		}

		// Try to navigate to frontend
		_, err = page.Goto(ph.baseURL)

		// Clean up temporary page and context
		_ = page.Close()
		_ = tempCtx.Close()

		// If successful, log and return
		if err == nil {
			ph.logger.Info(
				"Frontend verified accessible",
				"url", ph.baseURL,
				"attempt", attempt,
			)
			return nil
		}

		// Save error for final report
		lastErr = err

		// If not the last attempt, wait before retry
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	// All retries exhausted - build informative error message
	mode := os.Getenv("E2E_FRONTEND_MODE")
	if mode == "" {
		mode = "built (default)"
	}

	return fmt.Errorf(
		"frontend not accessible after %d attempts: %w\n"+
			"URL: %s\n"+
			"Mode: %s\n"+
			"Troubleshooting:\n"+
			"  - Dev mode: Run 'just web-dev' in another terminal\n"+
			"  - Built mode: Run 'just web-build && just run' or 'just build-all && just run'\n"+
			"  - Check port 3000 (dev) or 8000 (built) is not in use\n"+
			"  - Verify Go backend is running: curl http://localhost:8000/health",
		maxRetries,
		lastErr,
		ph.baseURL,
		mode,
	)
}

// Close gracefully closes browser and stops Playwright.
// Safe to call multiple times (idempotent).
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - error: If closing fails (rare)
//
// Cleanup:
//   - Closes browser instance
//   - Stops Playwright process
//   - Cleans up resources
//
// Example:
//
//	defer helper.Close(context.Background())
func (ph *PlaywrightHelper) Close(ctx context.Context) error {
	// Close browser if it's running
	if ph.browser != nil {
		if err := ph.browser.Close(); err != nil {
			ph.logger.Error("failed to close browser", "error", err)
		}
	}

	// Stop Playwright if it's running
	if ph.pw != nil {
		if err := ph.pw.Stop(); err != nil {
			ph.logger.Error("failed to stop Playwright", "error", err)
		}
	}

	return nil
}
