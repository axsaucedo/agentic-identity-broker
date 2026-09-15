package oauth2server

import (
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
)

func TestSessionProfile(t *testing.T) {
	t.Run("round trips email and display name", func(t *testing.T) {
		email := "u@example.com"
		session := &fosite.DefaultSession{}

		setSessionProfile(session, &email, "Jane Doe")

		gotEmail, gotDisplayName := sessionProfile(session)
		assert.Equal(t, &email, gotEmail)
		assert.Equal(t, "Jane Doe", gotDisplayName)
	})

	t.Run("omits nil email", func(t *testing.T) {
		session := &fosite.DefaultSession{}

		setSessionProfile(session, nil, "Jane Doe")

		gotEmail, gotDisplayName := sessionProfile(session)
		assert.Nil(t, gotEmail)
		assert.Equal(t, "Jane Doe", gotDisplayName)
		assert.NotContains(t, session.GetExtraClaims(), claimEmail)
	})

	t.Run("omits empty display name", func(t *testing.T) {
		session := &fosite.DefaultSession{}

		setSessionProfile(session, nil, "")

		gotEmail, gotDisplayName := sessionProfile(session)
		assert.Nil(t, gotEmail)
		assert.Empty(t, gotDisplayName)
		assert.NotContains(t, session.GetExtraClaims(), claimDisplayName)
	})

	t.Run("returns empty profile from bare session", func(t *testing.T) {
		gotEmail, gotDisplayName := sessionProfile(&fosite.DefaultSession{})
		assert.Nil(t, gotEmail)
		assert.Empty(t, gotDisplayName)
	})
}
