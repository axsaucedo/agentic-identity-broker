package oauth2server

import "github.com/ory/fosite"

const (
	claimEmail       = "email"
	claimDisplayName = "display_name"
)

func setSessionProfile(s *fosite.DefaultSession, email *string, displayName string) {
	extra := s.GetExtraClaims()
	if email != nil {
		extra[claimEmail] = *email
	}
	if displayName != "" {
		extra[claimDisplayName] = displayName
	}
}

func sessionProfile(s fosite.Session) (email *string, displayName string) {
	ecs, ok := s.(fosite.ExtraClaimsSession)
	if !ok {
		return nil, ""
	}
	extra := ecs.GetExtraClaims()
	if v, ok := extra[claimEmail].(string); ok && v != "" {
		email = &v
	}
	if v, ok := extra[claimDisplayName].(string); ok {
		displayName = v
	}
	return email, displayName
}
