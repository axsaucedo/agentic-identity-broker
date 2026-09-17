package support

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
)

const (
	gatewaySigningKeyID       = "test-key"
	defaultSubjectTokenTTL    = time.Hour
	gatewaySigningKeyFileName = "agentgateway-signing-key-*.pem"
)

// SubjectTokenOptions identifies the subject JWT the gateway presents to the Broker.
type SubjectTokenOptions struct {
	Issuer    string
	Audience  string
	Principal string
	AgentID   string
	Lifetime  time.Duration
}

// GatewaySigningKey is ephemeral signing material for one gateway E2E environment.
type GatewaySigningKey struct {
	privateKeyPath string
	privateKeyPEM  string
	publicJWKSet   map[string]interface{}

	closeOnce sync.Once
	closeErr  error
}

// NewGatewaySigningKey creates fresh RSA signing material and writes its private half to a
// temporary file for the agentgateway container mount.
func NewGatewaySigningKey() (*GatewaySigningKey, error) {
	privateKeyPEM, publicKeyPEM, err := helpers.GenerateTestRSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generate gateway signing key: %w", err)
	}

	publicJWKSet, err := helpers.GenerateJWKSFromPublicKey(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("generate gateway public JWKS: %w", err)
	}

	privateKeyFile, err := os.CreateTemp("", gatewaySigningKeyFileName)
	if err != nil {
		return nil, fmt.Errorf("create gateway signing-key file: %w", err)
	}
	privateKeyPath := privateKeyFile.Name()
	removePrivateKeyFile := true
	defer func() {
		if privateKeyFile != nil {
			_ = privateKeyFile.Close()
		}
		if removePrivateKeyFile {
			_ = os.Remove(privateKeyPath)
		}
	}()

	if err := privateKeyFile.Chmod(0o600); err != nil {
		return nil, fmt.Errorf("set gateway signing-key file permissions: %w", err)
	}
	if _, err := privateKeyFile.WriteString(privateKeyPEM); err != nil {
		return nil, fmt.Errorf("write gateway signing-key file: %w", err)
	}
	if err := privateKeyFile.Close(); err != nil {
		return nil, fmt.Errorf("close gateway signing-key file: %w", err)
	}
	privateKeyFile = nil

	removePrivateKeyFile = false
	return &GatewaySigningKey{
		privateKeyPath: privateKeyPath,
		privateKeyPEM:  privateKeyPEM,
		publicJWKSet:   publicJWKSet,
	}, nil
}

// PrivateKeyPath returns the mountable PEM private-key file path.
func (k *GatewaySigningKey) PrivateKeyPath() string {
	return k.privateKeyPath
}

// KeyID returns the key identifier shared by the JWKS and signed JWTs.
func (k *GatewaySigningKey) KeyID() string {
	return gatewaySigningKeyID
}

// PublicJWKSet returns the public JWKS matching the private key mounted in agentgateway.
func (k *GatewaySigningKey) PublicJWKSet() map[string]interface{} {
	return k.publicJWKSet
}

// MintSubjectToken creates a short-lived RS256 subject JWT for the supplied Broker trust tuple.
func (k *GatewaySigningKey) MintSubjectToken(options SubjectTokenOptions) (string, error) {
	lifetime := options.Lifetime
	if lifetime == 0 {
		lifetime = defaultSubjectTokenTTL
	}

	now := time.Now()
	token, err := helpers.SignTestJWT(map[string]interface{}{
		"sub": options.Principal,
		"azp": options.AgentID,
		"iss": options.Issuer,
		"aud": options.Audience,
		"exp": now.Add(lifetime).Unix(),
		"iat": now.Unix(),
	}, k.privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("sign gateway subject token: %w", err)
	}

	return token, nil
}

// Close removes the temporary private-key file. It is safe to call more than once.
func (k *GatewaySigningKey) Close() error {
	k.closeOnce.Do(func() {
		if k.privateKeyPath == "" {
			return
		}
		if err := os.Remove(k.privateKeyPath); err != nil && !os.IsNotExist(err) {
			k.closeErr = fmt.Errorf("remove gateway signing-key file: %w", err)
		}
	})
	return k.closeErr
}
