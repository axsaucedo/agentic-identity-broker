package bootstrap

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWSEmulatorContainerRestoresOriginalEnvironment(t *testing.T) {
	originals := map[string]string{
		"AWS_ENDPOINT_URL":          "https://aws.example.com",
		"AWS_ENDPOINT_URL_KMS":      "https://kms.example.com",
		"AWS_ENDPOINT_URL_DYNAMODB": "https://dynamodb.example.com",
		"AWS_ACCESS_KEY_ID":         "existing-access-key",
		"AWS_SECRET_ACCESS_KEY":     "existing-secret",
		"AWS_DEFAULT_REGION":        "us-west-2",
	}
	for key, value := range originals {
		t.Setenv(key, value)
	}

	emulator := &AWSEmulatorContainer{Endpoint: "http://localhost:4566"}

	require.NoError(t, emulator.SetupEnvironment())
	assert.Equal(t, emulator.Endpoint, os.Getenv("AWS_ENDPOINT_URL"))
	assert.Equal(t, emulator.Endpoint, os.Getenv("AWS_ENDPOINT_URL_KMS"))
	assert.Equal(t, emulator.Endpoint, os.Getenv("AWS_ENDPOINT_URL_DYNAMODB"))
	assert.Equal(t, "test", os.Getenv("AWS_ACCESS_KEY_ID"))
	assert.Equal(t, "test", os.Getenv("AWS_SECRET_ACCESS_KEY"))
	assert.Equal(t, "eu-central-1", os.Getenv("AWS_DEFAULT_REGION"))

	require.NoError(t, emulator.CleanupEnvironment())
	for key, value := range originals {
		assert.Equal(t, value, os.Getenv(key), key)
	}
}

func TestAWSEmulatorContainerSetupEnvironmentReturnsErrorForInvalidEndpoint(t *testing.T) {
	emulator := &AWSEmulatorContainer{Endpoint: "http://localhost\x00:4566"}

	err := emulator.SetupEnvironment()
	require.Error(t, err)
	assert.ErrorContains(t, err, "set AWS_ENDPOINT_URL")
}
