package oauth2server

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslateFositeError(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.NoError(t, translateFositeError(nil))
	})

	t.Run("non-fosite error returned unchanged", func(t *testing.T) {
		plain := errors.New("plain error")
		assert.Equal(t, plain, translateFositeError(plain))
	})

	t.Run("ErrInvalidRedirectURI returned unchanged", func(t *testing.T) {
		wrapped := fmt.Errorf("wrapping: %w", ErrInvalidRedirectURI)
		assert.ErrorIs(t, translateFositeError(wrapped), ErrInvalidRedirectURI)
	})

	t.Run("normal fosite error translates fields", func(t *testing.T) {
		fe := &fosite.RFC6749Error{
			ErrorField:       "invalid_client",
			DescriptionField: "client not found",
			CodeField:        http.StatusUnauthorized,
		}
		got := translateFositeError(fe)
		var rfc *RFC6749Error
		require.ErrorAs(t, got, &rfc)
		assert.Equal(t, "invalid_client", rfc.Code())
		assert.Equal(t, "client not found", rfc.Description())
		assert.Equal(t, http.StatusUnauthorized, rfc.HTTPStatus())
	})

	t.Run("zero CodeField falls back to 500", func(t *testing.T) {
		fe := &fosite.RFC6749Error{
			ErrorField:       "some_error",
			DescriptionField: "something went wrong",
			CodeField:        0,
		}
		got := translateFositeError(fe)
		var rfc *RFC6749Error
		require.ErrorAs(t, got, &rfc)
		assert.Equal(t, http.StatusInternalServerError, rfc.HTTPStatus())
	})

	t.Run("negative CodeField falls back to 500", func(t *testing.T) {
		fe := &fosite.RFC6749Error{
			ErrorField: "some_error",
			CodeField:  -1,
		}
		got := translateFositeError(fe)
		var rfc *RFC6749Error
		require.ErrorAs(t, got, &rfc)
		assert.Equal(t, http.StatusInternalServerError, rfc.HTTPStatus())
	})

	t.Run("empty ErrorField falls back to server_error", func(t *testing.T) {
		fe := &fosite.RFC6749Error{
			ErrorField: "",
			CodeField:  http.StatusInternalServerError,
		}
		got := translateFositeError(fe)
		var rfc *RFC6749Error
		require.ErrorAs(t, got, &rfc)
		assert.Equal(t, "server_error", rfc.Code())
	})

	t.Run("domain sentinel wired for known fosite errors", func(t *testing.T) {
		cases := []struct {
			fositeErr *fosite.RFC6749Error
			sentinel  error
		}{
			{fosite.ErrInvalidClient, ErrInvalidClient},
			{fosite.ErrInvalidGrant, ErrInvalidGrant},
			{fosite.ErrInvalidRequest, ErrInvalidRequest},
			{fosite.ErrInvalidScope, ErrInvalidScope},
			{fosite.ErrServerError, ErrServerError},
			{fosite.ErrUnsupportedResponseType, ErrUnsupportedResponseType},
		}
		for _, tc := range cases {
			got := translateFositeError(tc.fositeErr)
			assert.ErrorIs(t, got, tc.sentinel, "expected %v for fosite error %v", tc.sentinel, tc.fositeErr)
		}
	})
}
