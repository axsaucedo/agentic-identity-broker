package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuth2Flavor_Validate(t *testing.T) {
	tests := []struct {
		name    string
		flavor  OAuth2Flavor
		wantErr bool
	}{
		{"standard is valid", OAuth2FlavorStandard, false},
		{"google is valid", OAuth2FlavorGoogle, false},
		{"github is valid", OAuth2FlavorGitHub, false},
		{"empty string is invalid", OAuth2Flavor(""), true},
		{"unknown value is invalid", OAuth2Flavor("azure"), true},
		{"case-sensitive GitHub", OAuth2Flavor("GitHub"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flavor.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "github")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOAuth2Flavor_ScopeSeparator(t *testing.T) {
	tests := []struct {
		name   string
		flavor OAuth2Flavor
		want   string
	}{
		{"standard uses space", OAuth2FlavorStandard, " "},
		{"google uses space", OAuth2FlavorGoogle, " "},
		{"github uses comma", OAuth2FlavorGitHub, ","},
		{"empty uses space", OAuth2Flavor(""), " "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.flavor.ScopeSeparator())
		})
	}
}

func TestInferOAuth2Flavor(t *testing.T) {
	tests := []struct {
		name          string
		tokenEndpoint string
		want          OAuth2Flavor
	}{
		{"github.com token endpoint", "https://github.com/login/oauth/access_token", OAuth2FlavorGitHub},
		{"GitHub.com case insensitive", "https://GitHub.com/login/oauth/access_token", OAuth2FlavorGitHub},
		{"empty endpoint defaults to standard", "", OAuth2FlavorStandard},
		{"non-github endpoint defaults to standard", "https://auth.example.com/token", OAuth2FlavorStandard},
		{"google endpoint defaults to standard", "https://oauth2.googleapis.com/token", OAuth2FlavorStandard},
		{"different registrable domain defaults to standard", "https://notgithub.com/token", OAuth2FlavorStandard},
		{"github.com in path does not match", "https://example.com/github.com/token", OAuth2FlavorStandard},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, InferOAuth2Flavor(tt.tokenEndpoint))
		})
	}
}
