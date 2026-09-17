package gateway_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGatewayNativeTokenExchange(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Agentgateway Native Token Exchange")
}
