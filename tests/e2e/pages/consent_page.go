// Package pages provides page object models for frontend E2E testing.
// Page objects abstract UI selectors and interactions into high-level methods
// that tests call, making tests more readable and maintainable.
package pages

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Route constants for navigation
const agentDetailPath = "/consent/agent/%s"

// ConsentPage represents the OAuth2 consent flow page where users grant
// OAuth2 scopes to agents. This page is actually the AgentGrantDetailPage
// in the React app, which allows users to delegate services and scopes.
//
// ConsentPage wraps the base Page object for common navigation and utilities.
// All selectors use Playwright's semantic methods (GetByRole, GetByLabel, GetByText)
// to interact with accessible UI elements, never relying on HTML structure or CSS classes.
//
// Usage pattern:
//
//	consentPage := NewConsentPage(page, baseURL)
//	defer consentPage.Close()
//	consentPage.NavigateToAgent(ctx, "agent-id")
//	consentPage.SelectScope(ctx, "user:read")
//	consentPage.SubmitConsent(ctx)
type ConsentPage struct {
	*Page // Embed Page for method forwarding
}

// NewConsentPage creates a new ConsentPage instance.
//
// Parameters:
//   - page: Playwright page instance for this test
//   - baseURL: Base URL of the frontend (e.g., "http://localhost:3000" or "http://localhost:8000/consent")
//
// Returns:
//   - *ConsentPage: Initialized page object
//
// Example:
//
//	page := GetTestPage()
//	consentPage := NewConsentPage(page, GetFrontendURL())
func NewConsentPage(page playwright.Page, baseURL string) *ConsentPage {
	return &ConsentPage{
		Page: NewPage(page, baseURL),
	}
}

// NavigateToAgent navigates to the consent page for a specific agent.
//
// Parameters:
//   - ctx: Context for cancellation
//   - agentID: UUID of the agent
//
// Returns:
//   - error: If navigation fails
//
// The page waits implicitly for load events before returning.
// If the page fails to load, returns a descriptive error.
//
// Example:
//
//	err := consentPage.NavigateToAgent(ctx, "agent-uuid-123")
//	Expect(err).NotTo(HaveOccurred())
func (cp *ConsentPage) NavigateToAgent(ctx context.Context, agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agentID cannot be empty")
	}

	// Navigate to agent detail page
	// Route: /agent/:agentId (defined in React Router App.tsx)
	path := fmt.Sprintf(agentDetailPath, agentID)
	if err := cp.Navigate(ctx, path); err != nil {
		return fmt.Errorf("failed to navigate to agent consent page: %w", err)
	}

	// Wait for agent name heading to ensure page is interactive
	if err := cp.waitForAgentNameHeading(ctx); err != nil {
		return fmt.Errorf("agent name heading not found (page may not have loaded): %w", err)
	}

	return nil
}

// page returns the underlying Playwright page object.
func (cp *ConsentPage) page() playwright.Page {
	return cp.GetPlaywrightPage()
}

// getApproveButton returns the "Approve & Delegate" button locator.
// This private helper prevents duplication across SubmitConsent and IsConsentButtonEnabled.
func (cp *ConsentPage) getApproveButton(ctx context.Context) playwright.Locator {
	return cp.page().GetByRole(
		"button",
		playwright.PageGetByRoleOptions{Name: "Approve & Delegate"},
	)
}

// waitForAgentNameHeading waits for the main h1 heading (agent name) to appear.
// This waits for the page to load the agent detail content (not the header).
// Uses data-testid to avoid ambiguity when multiple h1 elements exist on page.
func (cp *ConsentPage) waitForAgentNameHeading(ctx context.Context) error {
	agentNameHeading := cp.page().GetByTestId("agent-name-heading")
	err := agentNameHeading.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("agent name heading not found (page may not have loaded): %w", err)
	}
	return nil
}

// SelectScope is not applicable in the current UI (Phase 8 - Simplified UI).
//
// In the current design, scopes are displayed as read-only StatusIndicator badges.
// Service delegation happens at the service level (Login/Delegate buttons), not at the scope level.
// Scopes are automatically included based on service requirements.
//
// This method is kept for API compatibility but always returns an error.
//
// Returns:
//   - error: Always returns an error indicating scopes are not selectable
//
// Deprecated: Scopes are determined by service requirements, not user selection.
func (cp *ConsentPage) SelectScope(ctx context.Context, scopeName string) error {
	if scopeName == "" {
		return fmt.Errorf("scopeName cannot be empty")
	}

	// In the current UI (Phase 8), scopes are not selectable.
	// Scopes are automatically included based on service requirements.
	// Use DelegateService() instead to grant access to a service with its required scopes.
	return fmt.Errorf("scope selection is not available in current UI; scopes are determined by service requirements. Use DelegateService() to grant service access")
}

// ClearScope is not applicable in the current UI (Phase 8 - Simplified UI).
//
// In the current design, scopes are displayed as read-only StatusIndicator badges.
// Scopes cannot be individually cleared; instead, revoke access to the entire service using RevokeService().
//
// This method is kept for API compatibility but always returns an error.
//
// Returns:
//   - error: Always returns an error indicating scopes are not modifiable
//
// Deprecated: Use RevokeService() instead to revoke all scopes for a service.
func (cp *ConsentPage) ClearScope(ctx context.Context, scopeName string) error {
	if scopeName == "" {
		return fmt.Errorf("scopeName cannot be empty")
	}

	// In the current UI (Phase 8), scopes are not individually modifiable.
	// Use RevokeService() instead to revoke all scopes for a service.
	return fmt.Errorf("scope clearing is not available in current UI; scopes are determined by service requirements. Use RevokeService() to revoke all scopes for a service")
}

// SubmitConsent finds the "Approve & Delegate" button and clicks it.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - error: If button not found or not clickable
//
// The button may be disabled if mandatory requirements are not met.
//
// Example:
//
//	err := consentPage.SubmitConsent(ctx)
//	Expect(err).NotTo(HaveOccurred())
func (cp *ConsentPage) SubmitConsent(ctx context.Context) error {
	button := cp.getApproveButton(ctx)

	// Check if button exists
	count, err := button.Count()
	if err != nil {
		return fmt.Errorf("failed to count submit button: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("approve & delegate button not found")
	}

	// Check if button is enabled
	enabled, err := button.IsEnabled()
	if err != nil {
		return fmt.Errorf("failed to check button enabled state: %w", err)
	}
	if !enabled {
		return fmt.Errorf("approve & delegate button not enabled (may require mandatory services to be connected)")
	}

	// Click the button
	if err := button.Click(); err != nil {
		return fmt.Errorf("failed to click Approve & Delegate button: %w", err)
	}

	return nil
}

// SetExpiration sets the grant expiration days using the expiration input.
//
// Parameters:
//   - ctx: Context for cancellation
//   - expirationDays: Number of days until expiration
//
// Returns:
//   - error: If expiration input not found or not fillable
//
// Note: This method calculates the expiration date and fills the date input.
// The date format expected is ISO 8601 (YYYY-MM-DD).
//
// Example:
//
//	err := consentPage.SetExpiration(ctx, 30)
//	Expect(err).NotTo(HaveOccurred())
func (cp *ConsentPage) SetExpiration(ctx context.Context, expirationDays int) error {
	if expirationDays <= 0 {
		return fmt.Errorf("expirationDays must be positive")
	}

	// Find expiration input by label
	input := cp.page().GetByLabel("Expiration")

	// Check if element exists
	count, err := input.Count()
	if err != nil {
		return fmt.Errorf("failed to count expiration input: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("expiration input not found")
	}

	// Calculate expiration date
	expirationDate := time.Now().AddDate(0, 0, expirationDays).Format("2006-01-02")

	// Fill the input
	if err := input.Fill(expirationDate); err != nil {
		return fmt.Errorf("failed to set expiration date: %w", err)
	}

	return nil
}

// SetExpirationDate sets the grant expiration to a specific date.
//
// Parameters:
//   - ctx: Context for cancellation
//   - date: Expiration date as time.Time
//
// Returns:
//   - error: If expiration input not found or not fillable
//
// The date is formatted as ISO 8601 (YYYY-MM-DD) for the date input.
//
// Example:
//
//	tomorrow := time.Now().AddDate(0, 0, 1)
//	err := consentPage.SetExpirationDate(ctx, tomorrow)
//	Expect(err).NotTo(HaveOccurred())
func (cp *ConsentPage) SetExpirationDate(ctx context.Context, date time.Time) error {
	// Find expiration input by label
	input := cp.page().GetByLabel("Expiration")

	// Check if element exists
	count, err := input.Count()
	if err != nil {
		return fmt.Errorf("failed to count expiration input: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("expiration input not found")
	}

	// Format date as ISO 8601
	dateStr := date.Format("2006-01-02")

	// Fill the input
	if err := input.Fill(dateStr); err != nil {
		return fmt.Errorf("failed to set expiration date: %w", err)
	}

	return nil
}

// GetAgentName retrieves the agent display name from the page heading.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - string: Agent display name
//   - error: If heading not found
//
// Example:
//
//	name, err := consentPage.GetAgentName(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(name).To(Equal("GitHub Agent"))
func (cp *ConsentPage) GetAgentName(ctx context.Context) (string, error) {
	// Find h1 heading in the main content area using semantic role
	heading := cp.page().GetByRole(
		"heading",
		playwright.PageGetByRoleOptions{Level: playwright.Int(1)},
	).First()

	// Check if element exists
	count, err := heading.Count()
	if err != nil {
		return "", fmt.Errorf("failed to count h1 heading: %w", err)
	}
	if count == 0 {
		return "", fmt.Errorf("agent name heading not found")
	}

	// Get the text content
	text, err := heading.TextContent()
	if err != nil {
		return "", fmt.Errorf("failed to get agent name heading: %w", err)
	}

	if text == "" {
		return "", fmt.Errorf("agent name heading is empty")
	}

	return strings.TrimSpace(text), nil
}

// GetAvailableScopes retrieves all available scopes displayed on the page.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - []string: Slice of scope names (e.g., ["repo", "user"])
//   - error: If no scopes found
//
// Scopes are displayed as StatusIndicator badges in the "Permissions" sections
// of each service card. This method finds all visible scope labels on the page.
//
// Example:
//
//	scopes, err := consentPage.GetAvailableScopes(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(scopes).To(ContainElement("repo"))
func (cp *ConsentPage) GetAvailableScopes(ctx context.Context) ([]string, error) {
	// StatusIndicators use data-testid for semantic identification
	// Pattern: data-testid="scope-indicator-*"
	indicators := cp.page().Locator("[data-testid^='scope-indicator-']")

	// Get count of indicators
	count, err := indicators.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count scope indicators: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("no scopes found on page")
	}

	scopes := []string{}
	seen := make(map[string]bool) // Track unique scopes

	// Iterate through each StatusIndicator
	for i := 0; i < count; i++ {
		indicator := indicators.Nth(i)

		// Get all spans within the indicator
		// The last span contains the scope label (first span is the icon)
		spans := indicator.Locator("span")
		spanCount, _ := spans.Count()

		if spanCount > 0 {
			// Get the last span which contains the label
			lastSpan := spans.Last()
			text, err := lastSpan.TextContent()
			if err == nil && text != "" {
				text = strings.TrimSpace(text)
				// Avoid duplicates and skip non-scope text (like "Permissions")
				if !seen[text] && text != "" && !strings.Contains(strings.ToLower(text), "permissions") {
					scopes = append(scopes, text)
					seen[text] = true
				}
			}
		}
	}

	if len(scopes) == 0 {
		return nil, fmt.Errorf("no scope labels found")
	}

	return scopes, nil
}

// GetSelectedScopes retrieves all currently selected/checked scopes.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - []string: Slice of selected scope names
//   - error: Only if there's an error accessing the page
//
// Returns an empty slice if no scopes are selected (not an error).
//
// Example:
//
//	selected, err := consentPage.GetSelectedScopes(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(selected).To(HaveLen(2))
func (cp *ConsentPage) GetSelectedScopes(ctx context.Context) ([]string, error) {
	// Find all checkboxes
	checkboxes := cp.page().GetByRole("checkbox")

	// Get count
	count, err := checkboxes.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count checkboxes: %w", err)
	}

	selected := []string{}

	// Iterate through each checkbox and check if it's checked
	for i := 0; i < count; i++ {
		checkbox := checkboxes.Nth(i)

		// Check if this checkbox is checked
		isChecked, err := checkbox.IsChecked()
		if err != nil {
			continue
		}

		if isChecked {
			// Get the scope label
			label, err := checkbox.GetAttribute("aria-label")
			if err == nil && label != "" {
				selected = append(selected, label)
			} else {
				// Try to find associated label by id
				id, err := checkbox.GetAttribute("id")
				if err == nil && id != "" {
					labelElem := cp.page().Locator(fmt.Sprintf("label[for='%s']", id))
					if count, _ := labelElem.Count(); count > 0 {
						text, err := labelElem.TextContent()
						if err == nil && text != "" {
							selected = append(selected, strings.TrimSpace(text))
						}
					}
				}
			}
		}
	}

	return selected, nil
}

// IsConsentButtonEnabled checks if the "Approve & Delegate" button is enabled.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - bool: True if button is enabled, false otherwise
//   - error: If button not found
//
// Example:
//
//	enabled, err := consentPage.IsConsentButtonEnabled(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(enabled).To(BeFalse()) // Still loading mandatory requirements
func (cp *ConsentPage) IsConsentButtonEnabled(ctx context.Context) (bool, error) {
	button := cp.getApproveButton(ctx)

	// Check if button exists
	count, err := button.Count()
	if err != nil {
		return false, fmt.Errorf("failed to count submit button: %w", err)
	}
	if count == 0 {
		return false, fmt.Errorf("approve & delegate button not found")
	}

	// Check if enabled
	enabled, err := button.IsEnabled()
	if err != nil {
		return false, fmt.Errorf("failed to check button enabled state: %w", err)
	}

	return enabled, nil
}

// GetSuccessMessage retrieves the success message after form submission.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - string: Success message text
//   - error: If success message not found
//
// Searches for an element with role="status" which contains the success feedback.
// This is typically a toast or status region that appears after successful submission.
//
// Example:
//
//	msg, err := consentPage.GetSuccessMessage(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(msg).To(ContainSubstring("updated successfully"))
func (cp *ConsentPage) GetSuccessMessage(ctx context.Context) (string, error) {
	// Find status region
	status := cp.page().GetByRole("status")

	// Check if it exists
	count, err := status.Count()
	if err != nil || count == 0 {
		return "", fmt.Errorf("success message (status region) not found")
	}

	// Get the text content
	text, err := status.First().TextContent()
	if err != nil {
		return "", fmt.Errorf("failed to get success message text: %w", err)
	}

	return strings.TrimSpace(text), nil
}

// GetErrorMessage retrieves the error message if one is displayed.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - string: Error message text
//   - error: If error region not found or empty
//
// Searches for an element with role="alert" which contains the error feedback.
// Returns empty string and nil if no error is found (not an error condition).
//
// Example:
//
//	msg, err := consentPage.GetErrorMessage(ctx)
//	// err will be non-nil if no error message is displayed
//	if err == nil {
//	  fmt.Printf("Error: %s\n", msg)
//	}
func (cp *ConsentPage) GetErrorMessage(ctx context.Context) (string, error) {
	// Find alert region
	alert := cp.page().GetByRole("alert")

	// Check if it exists
	count, err := alert.Count()
	if err != nil || count == 0 {
		return "", fmt.Errorf("no error message found")
	}

	// Get the text content
	text, err := alert.First().TextContent()
	if err != nil {
		return "", fmt.Errorf("failed to get error message text: %w", err)
	}

	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("error message is empty")
	}

	return strings.TrimSpace(text), nil
}

// HasError checks if an error message is currently displayed.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - bool: True if error region found, false otherwise
//   - error: Only if there's a critical error accessing the page
//
// Example:
//
//	hasErr, err := consentPage.HasError(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(hasErr).To(BeTrue())
func (cp *ConsentPage) HasError(ctx context.Context) (bool, error) {
	// Try to find alert region
	count, err := cp.page().GetByRole("alert").Count()
	if err != nil {
		return false, fmt.Errorf("failed to check for error: %w", err)
	}

	return count > 0, nil
}

// GetExpirationDate retrieves the current expiration date value from the input.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - string: Expiration date value (ISO 8601 format YYYY-MM-DD)
//   - error: If expiration input not found
//
// Example:
//
//	date, err := consentPage.GetExpirationDate(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(date).To(Equal("2026-02-15"))
func (cp *ConsentPage) GetExpirationDate(ctx context.Context) (string, error) {
	// Find expiration input by label
	input := cp.page().GetByLabel("Expiration")

	// Check if element exists
	count, err := input.Count()
	if err != nil {
		return "", fmt.Errorf("failed to count expiration input: %w", err)
	}
	if count == 0 {
		return "", fmt.Errorf("expiration input not found")
	}

	// Get the value attribute
	value, err := input.InputValue()
	if err != nil {
		return "", fmt.Errorf("failed to get expiration input value: %w", err)
	}

	return value, nil
}

// DelegateService clicks the Delegate or Login button for a service.
//
// Parameters:
//   - ctx: Context for cancellation
//   - serviceDisplayName: Display name of the service (e.g., "GitHub")
//
// Returns:
//   - error: If service button not found or not clickable
//
// Finds the service by its display name heading (semantic role="heading" level 3)
// then locates and clicks the action button (Login or Delegate).
// Uses data-testid for reliable button identification without brittle XPath selectors.
//
// Example:
//
//	err := consentPage.DelegateService(ctx, "GitHub")
//	Expect(err).NotTo(HaveOccurred())
func (cp *ConsentPage) DelegateService(ctx context.Context, serviceDisplayName string) error {
	if serviceDisplayName == "" {
		return fmt.Errorf("serviceDisplayName cannot be empty")
	}

	// Step 1: Find the service heading (semantic)
	serviceHeading := cp.page().GetByRole(
		"heading",
		playwright.PageGetByRoleOptions{Name: serviceDisplayName, Level: playwright.Int(3)},
	)

	count, err := serviceHeading.Count()
	if err != nil {
		return fmt.Errorf("failed to find service %q: %w", serviceDisplayName, err)
	}
	if count == 0 {
		return fmt.Errorf("service %q not found on page", serviceDisplayName)
	}

	// Step 2: Find the article parent (semantic role wrapper added in React)
	// Use aria-label attribute for semantic, accessible selector
	serviceArticle := cp.page().Locator(
		fmt.Sprintf(`article[aria-label="Service: %s"]`, serviceDisplayName),
	).First()

	articleCount, err := serviceArticle.Count()
	if err != nil || articleCount == 0 {
		return fmt.Errorf("service article wrapper not found for %q", serviceDisplayName)
	}

	// Step 3: Find action button within the service article (not page-wide)
	// Try Login button first (for services not yet connected)
	actionBtn := serviceArticle.Locator("[data-testid='service-login-button']")
	if btnCount, _ := actionBtn.Count(); btnCount > 0 {
		if err := actionBtn.Click(); err != nil {
			return fmt.Errorf("failed to click login button for service %q: %w", serviceDisplayName, err)
		}
		return nil
	}

	// Try Delegate button (for connected services not yet delegated)
	actionBtn = serviceArticle.Locator("[data-testid='service-delegate-button']")
	if btnCount, _ := actionBtn.Count(); btnCount > 0 {
		if err := actionBtn.Click(); err != nil {
			return fmt.Errorf("failed to click delegate button for service %q: %w", serviceDisplayName, err)
		}
		return nil
	}

	return fmt.Errorf("action button (Login or Delegate) not found for service %q", serviceDisplayName)
}

// RevokeService clicks the Revoke button for a service.
//
// Parameters:
//   - ctx: Context for cancellation
//   - serviceDisplayName: Display name of the service (e.g., "GitHub")
//
// Returns:
//   - error: If service or Revoke button not found
//
// Example:
//
//	err := consentPage.RevokeService(ctx, "GitHub")
//	Expect(err).NotTo(HaveOccurred())
func (cp *ConsentPage) RevokeService(ctx context.Context, serviceDisplayName string) error {
	if serviceDisplayName == "" {
		return fmt.Errorf("serviceDisplayName cannot be empty")
	}

	// Find service heading (semantic)
	serviceHeading := cp.page().GetByRole(
		"heading",
		playwright.PageGetByRoleOptions{Name: serviceDisplayName, Level: playwright.Int(3)},
	)

	count, err := serviceHeading.Count()
	if err != nil {
		return fmt.Errorf("failed to find service %q: %w", serviceDisplayName, err)
	}
	if count == 0 {
		return fmt.Errorf("service %q not found on page", serviceDisplayName)
	}

	// Find the article parent (semantic role wrapper)
	// Use aria-label attribute for semantic, accessible selector
	serviceArticle := cp.page().Locator(
		fmt.Sprintf(`article[aria-label="Service: %s"]`, serviceDisplayName),
	).First()

	if articleCount, _ := serviceArticle.Count(); articleCount == 0 {
		return fmt.Errorf("service article wrapper not found for %q", serviceDisplayName)
	}

	// Find Revoke button within service article
	revokeBtn := serviceArticle.GetByRole(
		"button",
		playwright.LocatorGetByRoleOptions{Name: "Revoke"},
	).First()

	btnCount, err := revokeBtn.Count()
	if err != nil || btnCount == 0 {
		return fmt.Errorf("revoke button not found for service %q", serviceDisplayName)
	}

	if err := revokeBtn.Click(); err != nil {
		return fmt.Errorf("failed to click revoke button for service %q: %w", serviceDisplayName, err)
	}

	return nil
}

// GetServiceCount returns the number of services displayed on the page.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - int: Number of services
//   - error: If failed to count services
//
// Example:
//
//	count, err := consentPage.GetServiceCount(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(count).To(Equal(3))
func (cp *ConsentPage) GetServiceCount(ctx context.Context) (int, error) {
	// Count service headings (level 3)
	headings := cp.page().GetByRole(
		"heading",
		playwright.PageGetByRoleOptions{Level: playwright.Int(3)},
	)

	count, err := headings.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count service headings: %w", err)
	}

	// Exclude the main title (Services) heading
	if count > 0 {
		count-- // Subtract 1 for the "Services" section heading
	}

	return count, nil
}

// GetServiceNames returns the display names of all services on the page.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - []string: Slice of service display names
//   - error: If failed to retrieve service names
//
// Example:
//
//	names, err := consentPage.GetServiceNames(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(names).To(Equal([]string{"GitHub", "Google Cloud"}))
func (cp *ConsentPage) GetServiceNames(ctx context.Context) ([]string, error) {
	// Count service headings (level 3)
	headings := cp.page().GetByRole(
		"heading",
		playwright.PageGetByRoleOptions{Level: playwright.Int(3)},
	)

	count, err := headings.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count service headings: %w", err)
	}

	names := []string{}

	// Iterate through headings starting from index 1 (skip the main "Services" heading)
	for i := 1; i < count; i++ {
		heading := headings.Nth(i)
		text, err := heading.TextContent()
		if err != nil {
			continue
		}
		names = append(names, strings.TrimSpace(text))
	}

	return names, nil
}

// GetValidationErrorCount returns the number of validation errors displayed.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - int: Number of validation errors
//   - error: If failed to count errors
//
// Example:
//
//	count, err := consentPage.GetValidationErrorCount(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(count).To(Equal(1))
func (cp *ConsentPage) GetValidationErrorCount(ctx context.Context) (int, error) {
	// Find error container using semantic role="alert"
	errorAlert := cp.page().GetByRole("alert").First()

	// Check if error container exists
	count, err := cp.page().GetByRole("alert").Count()
	if err != nil {
		return 0, fmt.Errorf("failed to check for error container: %w", err)
	}

	if count == 0 {
		return 0, nil
	}

	// Count list items in error container
	errorItems := errorAlert.GetByRole("listitem")
	itemCount, err := errorItems.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count error items: %w", err)
	}

	return itemCount, nil
}

// GetValidationErrors returns all validation error messages displayed.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - []string: Slice of error message texts
//   - error: If failed to retrieve error messages
//
// Example:
//
//	errors, err := consentPage.GetValidationErrors(ctx)
//	Expect(err).NotTo(HaveOccurred())
//	Expect(errors).To(ContainElement(ContainSubstring("required")))
func (cp *ConsentPage) GetValidationErrors(ctx context.Context) ([]string, error) {
	// Find error container using semantic role="alert"
	errorAlert := cp.page().GetByRole("alert").First()

	// Check if error container exists
	count, err := cp.page().GetByRole("alert").Count()
	if err != nil {
		return nil, fmt.Errorf("failed to check for error container: %w", err)
	}

	if count == 0 {
		return []string{}, nil
	}

	// Get all error items
	errorItems := errorAlert.GetByRole("listitem")
	itemCount, err := errorItems.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count error items: %w", err)
	}

	errors := []string{}
	for i := 0; i < itemCount; i++ {
		item := errorItems.Nth(i)
		text, err := item.TextContent()
		if err != nil {
			continue
		}
		errors = append(errors, strings.TrimSpace(text))
	}

	return errors, nil
}

// IsMandatoryServiceConnected checks if a mandatory service shows as connected/delegated.
//
// Parameters:
//   - ctx: Context for cancellation
//   - serviceDisplayName: Display name of the service
//
// Returns:
//   - bool: True if service is connected/delegated
//   - error: If service not found
//
// Example:
//
//	connected, err := consentPage.IsMandatoryServiceConnected(ctx, "GitHub")
//	Expect(err).NotTo(HaveOccurred())
//	Expect(connected).To(BeTrue())
func (cp *ConsentPage) IsMandatoryServiceConnected(ctx context.Context, serviceDisplayName string) (bool, error) {
	if serviceDisplayName == "" {
		return false, fmt.Errorf("serviceDisplayName cannot be empty")
	}

	// Find service heading (semantic)
	serviceHeading := cp.page().GetByRole(
		"heading",
		playwright.PageGetByRoleOptions{Name: serviceDisplayName, Level: playwright.Int(3)},
	)

	count, err := serviceHeading.Count()
	if err != nil {
		return false, fmt.Errorf("failed to find service %q: %w", serviceDisplayName, err)
	}
	if count == 0 {
		return false, fmt.Errorf("service %q not found", serviceDisplayName)
	}

	// Find the article parent (semantic role wrapper)
	// Use aria-label attribute for semantic, accessible selector
	serviceArticle := cp.page().Locator(
		fmt.Sprintf(`article[aria-label="Service: %s"]`, serviceDisplayName),
	).First()

	if articleCount, _ := serviceArticle.Count(); articleCount == 0 {
		return false, fmt.Errorf("service article wrapper not found for %q", serviceDisplayName)
	}

	// Check for Revoke button (indicates service is delegated)
	revokeBtn := serviceArticle.GetByRole("button", playwright.LocatorGetByRoleOptions{Name: "Revoke"})
	btnCount, _ := revokeBtn.Count()

	return btnCount > 0, nil
}

// WaitForServiceToAppear waits for a specific service to appear on the page.
//
// Parameters:
//   - ctx: Context for cancellation
//   - serviceDisplayName: Display name of the service to wait for
//   - timeout: Maximum time to wait (in milliseconds)
//
// Returns:
//   - error: If service doesn't appear within timeout
//
// Example:
//
//	err := consentPage.WaitForServiceToAppear(ctx, "GitHub", 5000)
//	Expect(err).NotTo(HaveOccurred())
func (cp *ConsentPage) WaitForServiceToAppear(ctx context.Context, serviceDisplayName string, timeout int) error {
	if serviceDisplayName == "" {
		return fmt.Errorf("serviceDisplayName cannot be empty")
	}

	if timeout <= 0 {
		timeout = 5000 // Default to 5 seconds
	}

	// Wait for service heading using Playwright's built-in wait mechanism
	heading := cp.page().GetByRole(
		"heading",
		playwright.PageGetByRoleOptions{
			Name:  serviceDisplayName,
			Level: playwright.Int(3),
		},
	)
	err := heading.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(float64(timeout)),
	})
	if err != nil {
		return fmt.Errorf("service %q did not appear within %dms: %w", serviceDisplayName, timeout, err)
	}
	return nil
}
