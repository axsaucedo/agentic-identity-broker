package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var waitForEndpointHTTPClient = &http.Client{Timeout: 250 * time.Millisecond}

func waitForEndpoint(t *testing.T, url string) {
	t.Helper()

	require.Eventually(t, func() bool {
		resp, err := waitForEndpointHTTPClient.Get(url)
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode < http.StatusInternalServerError
	}, 2*time.Second, 25*time.Millisecond, "endpoint never became ready: %s", url)
}
