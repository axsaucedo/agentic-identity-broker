package support

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
)

var defaultTransportMu sync.Mutex

// HTTPSJWKSServer serves a gateway public JWK set through a TLS-protected endpoint.
type HTTPSJWKSServer struct {
	server      *httptest.Server
	certificate *x509.Certificate
	systemRoots *x509.CertPool

	requestCount atomic.Int64
	closeOnce    sync.Once
}

// NewHTTPSJWKSServer starts a TLS server that serves publicJWKSet at its URL.
func NewHTTPSJWKSServer(publicJWKSet map[string]interface{}) (*HTTPSJWKSServer, error) {
	encodedJWKSet, err := json.Marshal(publicJWKSet)
	if err != nil {
		return nil, fmt.Errorf("marshal gateway public JWK set: %w", err)
	}

	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("load system certificate pool: %w", err)
	}
	if systemRoots == nil {
		systemRoots = x509.NewCertPool()
	}

	fixture := &HTTPSJWKSServer{systemRoots: systemRoots}
	fixture.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fixture.requestCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(encodedJWKSet)
	}))
	fixture.certificate = fixture.server.Certificate()
	if fixture.certificate == nil {
		fixture.server.Close()
		return nil, fmt.Errorf("read HTTPS JWKS fixture certificate")
	}

	return fixture, nil
}

// URL returns the TLS-protected public JWK set endpoint.
func (s *HTTPSJWKSServer) URL() string {
	if s == nil || s.server == nil {
		return ""
	}
	return s.server.URL
}

// RequestCount returns the number of requests received by the fixture.
func (s *HTTPSJWKSServer) RequestCount() int {
	if s == nil {
		return 0
	}
	return int(s.requestCount.Load())
}

// InstallDefaultTransport adds this fixture's certificate to a clone of the
// default transport's roots. The returned cleanup restores the exact original
// default transport.
func (s *HTTPSJWKSServer) InstallDefaultTransport() func() {
	defaultTransportMu.Lock()

	original := http.DefaultTransport
	transport, ok := original.(*http.Transport)
	if !ok {
		defaultTransportMu.Unlock()
		panic("HTTPSJWKSServer: http.DefaultTransport is not an *http.Transport")
	}

	cloned := transport.Clone()
	tlsConfig := cloned.TLSClientConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	} else {
		tlsConfig = tlsConfig.Clone()
	}

	if tlsConfig.RootCAs == nil {
		tlsConfig.RootCAs = s.systemRoots.Clone()
	} else {
		tlsConfig.RootCAs = tlsConfig.RootCAs.Clone()
	}
	tlsConfig.RootCAs.AddCert(s.certificate)
	tlsConfig.InsecureSkipVerify = false
	cloned.TLSClientConfig = tlsConfig
	http.DefaultTransport = cloned

	var restoreOnce sync.Once
	return func() {
		restoreOnce.Do(func() {
			cloned.CloseIdleConnections()
			http.DefaultTransport = original
			defaultTransportMu.Unlock()
		})
	}
}

// Close stops the HTTPS JWKS fixture.
func (s *HTTPSJWKSServer) Close() {
	if s == nil || s.server == nil {
		return
	}
	s.closeOnce.Do(s.server.Close)
}
