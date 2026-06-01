package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ProvisionSigningKey calls the admin API to generate an initial signing key and
// immediately promotes it so it's active (bypassing the grace period). Safe to
// call in BeforeEach whenever local token issuance is needed.
func ProvisionSigningKey(adminBaseURL string) error {
	resp, err := http.Post(adminBaseURL+"/api/oauth2-server/signing-keys", "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to provision signing key: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("provision signing key returned %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("failed to decode signing key response: %w", err)
	}
	kid, ok := body["kid"].(string)
	if !ok || kid == "" {
		return fmt.Errorf("signing key response missing kid")
	}

	// PUT /current to reset activates_at to now, bypassing the grace period.
	req, err := http.NewRequest(http.MethodPut,
		adminBaseURL+"/api/oauth2-server/signing-keys/"+kid+"/current",
		strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("failed to build promote request: %w", err)
	}
	promResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to promote signing key: %w", err)
	}
	defer func() { _ = promResp.Body.Close() }()
	if promResp.StatusCode != http.StatusOK {
		return fmt.Errorf("promote signing key returned %d", promResp.StatusCode)
	}
	return nil
}
