package gateway_test

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/gateway/support"
)

const directGatewayGuideLink = "token-exchange-gateway-direct"

var directGatewayGuideLinkPattern = regexp.MustCompile(`\[[^\]]+\]\([^)]*token-exchange-gateway-direct(?:\.md)?[^)]*\)`)

var _ = Describe("Agentgateway Native Token Exchange", func() {
	Context("Deploy the Alternative Safely", func() {
		var (
			directReference  []byte
			extProcReference []byte
			extProcGuide     string
			environment      *nativeGatewayEnvironment
		)

		BeforeEach(func() {
			var err error
			directReference, err = support.LoadAgentgatewayDirectReference()
			Expect(err).NotTo(HaveOccurred())

			extProcReference = readGatewayRepositoryFile("mocks", "agentgateway", "config.yaml")
			extProcGuide = string(readGatewayRepositoryFile("docs", "guides", "token-exchange-gateway.md"))
			environment = startNativeGatewayEnvironment(nativeGatewayEnvironmentOptions{
				directReferenceYAML: directReference,
			})
		})

		// US3-S1 from specs/045-agentgateway-token-exchange/spec.md
		It("ships the ExtProc and direct configurations as alternatives, never combined", func() {
			directReferenceText := string(directReference)
			extProcReferenceText := string(extProcReference)

			Expect(directReferenceText).To(ContainSubstring("backendAuth:"))
			Expect(directReferenceText).To(ContainSubstring("oauthTokenExchange:"))
			Expect(directReferenceText).NotTo(ContainSubstring("extProc:"))
			Expect(extProcReferenceText).To(ContainSubstring("extProc:"))
			Expect(extProcReferenceText).NotTo(ContainSubstring("oauthTokenExchange:"))
			Expect(string(environment.gateway.RenderedConfig)).NotTo(ContainSubstring("extProc:"))

			directGuide := readDirectGatewayGuideAfterLink(extProcGuide)
			Expect(directGuide).To(MatchRegexp(`(?i)alternative`))
			Expect(directGuide).To(ContainSubstring("privateKeyJwt"))
			Expect(directGuide).To(MatchRegexp(`(?is)(do not|must not|never|exclude)[^.]{0,160}clientSecretBasic`))
			Expect(directGuide).To(MatchRegexp(`(?is)(do not|must not|never|exclude)[^.]{0,160}clientSecretPost`))
		})

		// US3-S2 from specs/045-agentgateway-token-exchange/spec.md
		It("completes an exchange from the reference configuration without any embedded secret", func() {
			inboundBearer := environment.mintSubjectToken("", "")
			directReferenceText := string(directReference)
			renderedConfigText := string(environment.gateway.RenderedConfig)

			Expect(directReferenceText).To(ContainSubstring("backendAuth:"))
			Expect(directReferenceText).To(ContainSubstring("oauthTokenExchange:"))
			Expect(directReferenceText).To(ContainSubstring("path: /oauth2/token"))
			Expect(directReferenceText).To(ContainSubstring("method: privateKeyJwt"))
			Expect(directReferenceText).To(ContainSubstring("signingKey:\n                  file:"))
			Expect(directReferenceText).NotTo(ContainSubstring("clientSecret"))
			Expect(directReferenceText).NotTo(ContainSubstring("-----BEGIN"))
			Expect(renderedConfigText).NotTo(ContainSubstring("clientSecret"))
			Expect(renderedConfigText).NotTo(ContainSubstring("-----BEGIN"))

			result, err := environment.gateway.CallWhoAmI(environment.ctx, inboundBearer)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(Equal("Bearer " + inboundBearer))
			Expect(lastBackendAuthorization(environment.backend)).To(HavePrefix("Bearer "))
			Expect(lastBackendAuthorization(environment.backend)).NotTo(Equal("Bearer " + inboundBearer))
		})

		// US3-S3 from specs/045-agentgateway-token-exchange/spec.md
		It("verifies a direct deployment through the documented steps", func() {
			inboundBearer := environment.mintSubjectToken("", "")

			directGuide := readDirectGatewayGuideAfterLink(extProcGuide)
			Expect(directGuide).To(MatchRegexp(`(?i)verif`))
			Expect(directGuide).To(MatchRegexp(`(?i)authorization`))
			Expect(directGuide).To(ContainSubstring("ExtProc"))
			Expect(string(directReference)).NotTo(ContainSubstring("extProc:"))
			Expect(string(environment.gateway.RenderedConfig)).NotTo(ContainSubstring("extProc:"))
			Expect(string(environment.gateway.RenderedConfig)).NotTo(ContainSubstring("extproc-token-exchange"))

			result, err := environment.gateway.CallWhoAmI(environment.ctx, inboundBearer)
			Expect(err).NotTo(HaveOccurred())
			Expect(environment.backend.RequestCount()).To(BeNumerically(">", 0))
			Expect(result).To(Equal(lastBackendAuthorization(environment.backend)))
			Expect(environment.backend.AuthorizationHeaders()).NotTo(ContainElement("Bearer " + inboundBearer))
			Expect(lastBackendAuthorization(environment.backend)).NotTo(Equal("Bearer " + inboundBearer))
			Expect(environment.gateway.ExtProc.ConnectionCount()).To(Equal(0))
		})
	})
})

func readDirectGatewayGuideAfterLink(extProcGuide string) string {
	if !directGatewayGuideLinkPattern.MatchString(extProcGuide) {
		Expect(extProcGuide).NotTo(MatchRegexp(directGatewayGuideLinkPattern.String()))
		Expect(extProcGuide).To(MatchRegexp(directGatewayGuideLinkPattern.String()),
			"the existing ExtProc guide must link the direct-route alternative before its guide is read")
		return ""
	}

	return string(readGatewayRepositoryFile("docs", "guides", directGatewayGuideLink+".md"))
}

func readGatewayRepositoryFile(elements ...string) []byte {
	path, err := gatewayRepositoryPath(elements...)
	Expect(err).NotTo(HaveOccurred())

	contents, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return contents
}

func gatewayRepositoryPath(elements ...string) (string, error) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("locate gateway reference configuration test source")
	}

	parts := append([]string{filepath.Dir(sourceFile), "..", "..", ".."}, elements...)
	return filepath.Join(parts...), nil
}
