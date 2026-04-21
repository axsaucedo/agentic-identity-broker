package oauth2server

import (
	"github.com/ory/fosite"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// brokerClient wraps an Agent + ClientCredential as a fosite.Client.
// Fosite types are contained within this package and never leak.
type brokerClient struct {
	agent      *storage.Agent
	credential *storage.ClientCredential
}

var _ fosite.Client = (*brokerClient)(nil)

func (c *brokerClient) GetID() string {
	return c.agent.ID.String()
}

func (c *brokerClient) GetHashedSecret() []byte {
	return []byte(c.credential.SecretHash)
}

func (c *brokerClient) GetRedirectURIs() []string {
	return c.agent.RedirectURIs
}

func (c *brokerClient) GetGrantTypes() fosite.Arguments {
	return fosite.Arguments{"authorization_code", "client_credentials"}
}

func (c *brokerClient) GetResponseTypes() fosite.Arguments {
	return fosite.Arguments{"code"}
}

func (c *brokerClient) GetScopes() fosite.Arguments {
	if len(c.agent.AllowedScopes) > 0 {
		return fosite.Arguments(c.agent.AllowedScopes)
	}
	return fosite.Arguments{} // Empty means unrestricted
}

func (c *brokerClient) IsPublic() bool {
	return false // Broker clients are always confidential
}

func (c *brokerClient) GetAudience() fosite.Arguments {
	return fosite.Arguments{}
}
