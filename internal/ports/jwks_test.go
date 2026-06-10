package ports

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJWKSPort_EmbedsJWKSHealthPort(t *testing.T) {
	jwksPortType := reflect.TypeOf((*JWKSPort)(nil)).Elem()
	jwksHealthPortType := reflect.TypeOf((*JWKSHealthPort)(nil)).Elem()

	assert.True(t, jwksPortType.Implements(jwksHealthPortType))
}
