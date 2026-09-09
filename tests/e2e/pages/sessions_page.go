package pages

import (
	"context"
	"fmt"
	"strings"

	"github.com/mxschmitt/playwright-go"
)

const thirdPartySessionsRoute = "/sessions"

type SessionsPage struct {
	*Page
}

func NewSessionsPage(page playwright.Page, baseURL string) *SessionsPage {
	return &SessionsPage{Page: NewPage(page, baseURL)}
}

func (sp *SessionsPage) NavigateToSessions(ctx context.Context) error {
	if err := sp.Navigate(ctx, thirdPartySessionsRoute); err != nil {
		return err
	}
	return sp.waitForPageLoad(ctx)
}

func (sp *SessionsPage) IsRefreshButtonVisible(ctx context.Context) (bool, error) {
	button := sp.refreshButtonLocator()
	count, err := button.Count()
	if err != nil {
		return false, fmt.Errorf("failed to count refresh buttons: %w", err)
	}
	if count == 0 {
		return false, nil
	}
	visible, err := button.First().IsVisible()
	if err != nil {
		return false, fmt.Errorf("failed to determine refresh button visibility: %w", err)
	}
	return visible, nil
}

func (sp *SessionsPage) ClickRefreshButton(ctx context.Context) error {
	button := sp.refreshButtonLocator().First()
	if err := button.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("refresh button did not become visible: %w", err)
	}
	if err := button.Click(); err != nil {
		return fmt.Errorf("failed to click refresh button: %w", err)
	}
	return nil
}

func (sp *SessionsPage) WaitForSuccessMessage(ctx context.Context, message string, timeoutMs int) error {
	if timeoutMs <= 0 {
		timeoutMs = 5000
	}

	status := sp.page().GetByRole("status").First()
	if err := status.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(float64(timeoutMs)),
	}); err != nil {
		return fmt.Errorf("success status did not appear: %w", err)
	}

	text, err := status.TextContent()
	if err != nil {
		return fmt.Errorf("failed to read success status text: %w", err)
	}
	if !strings.Contains(text, message) {
		return fmt.Errorf("success status %q did not contain %q", text, message)
	}
	return nil
}

func (sp *SessionsPage) page() playwright.Page {
	return sp.GetPlaywrightPage()
}

func (sp *SessionsPage) waitForPageLoad(ctx context.Context) error {
	heading := sp.page().GetByRole(
		"heading",
		playwright.PageGetByRoleOptions{Name: "Third-Party Sessions", Level: playwright.Int(2)},
	).First()
	if err := heading.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("third-party sessions heading did not appear: %w", err)
	}
	return nil
}

func (sp *SessionsPage) refreshButtonLocator() playwright.Locator {
	return sp.page().GetByRole(
		"button",
		playwright.PageGetByRoleOptions{Name: "Refresh"},
	)
}
