// Package main is the AWS CDK application entrypoint for the
// Agentic Identity Broker encryption infrastructure.
//
// It provisions KMS, DynamoDB, and IAM resources required by the
// Token Vault's hierarchical keyring envelope encryption.
//
// Usage:
//
//	cdk synth -c env=test
//	cdk synth -c env=sandbox \
//	  -c k8sNamespace=agentic-identity-broker-sandbox \
//	  -c k8sServiceAccountName=agentic-identity-broker
//	cdk synth -c env=prod \
//	  -c k8sNamespace=agentic-identity-broker \
//	  -c k8sServiceAccountName=agentic-identity-broker
//
// Custom Tags (optional):
//
//	cdk synth -c env=test -c 'customTags={"Application":"TokenVault","Team":"Security","CostCenter":"CC123"}'
package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Read environment from CDK context (default: "test").
	env := "test"
	if v := app.Node().TryGetContext(jsii.String("env")); v != nil {
		if s, ok := v.(string); ok {
			env = s
		}
	}

	// Validate environment parameter to prevent typos and unexpected stack creation.
	validEnvs := map[string]bool{"test": true, "sandbox": true, "prod": true, "production": true}
	if !validEnvs[env] {
		panic(fmt.Sprintf("Invalid environment '%s'. Must be one of: test, sandbox, prod", env))
	}

	// Normalize "production" to "prod" for consistency.
	if env == "production" {
		env = "prod"
	}

	// Read service account parameters for Zalando CDP trust relationship.
	// Defaults are provided per environment; override via context if needed.
	defaultNamespaces := map[string]string{
		"test":    "agentic-identity-broker",
		"sandbox": "agentic-identity-broker-sandbox",
		"prod":    "agentic-identity-broker",
	}
	defaultServiceAccounts := map[string]string{
		"test":    "agentic-identity-broker",
		"sandbox": "agentic-identity-broker",
		"prod":    "agentic-identity-broker",
	}

	k8sNamespace := defaultNamespaces[env]
	if v := app.Node().TryGetContext(jsii.String("k8sNamespace")); v != nil {
		if s, ok := v.(string); ok && s != "" {
			k8sNamespace = s
		}
	}

	k8sServiceAccountName := defaultServiceAccounts[env]
	if v := app.Node().TryGetContext(jsii.String("k8sServiceAccountName")); v != nil {
		if s, ok := v.(string); ok && s != "" {
			k8sServiceAccountName = s
		}
	}

	// Read custom tags from CDK context (optional).
	// Usage: cdk synth -c customTags='{"Application":"TokenVault","Team":"Security"}'
	customTags := make(map[string]string)
	if v := app.Node().TryGetContext(jsii.String("customTags")); v != nil {
		if m, ok := v.(map[string]interface{}); ok {
			for key, val := range m {
				if str, ok := val.(string); ok {
					customTags[key] = str
				}
			}
		}
	}

	stackName := "AgenticIdentityBrokerEncryptionVault-" + env

	NewEncryptionStack(app, stackName, &EncryptionStackProps{
		StackProps: awscdk.StackProps{
			StackName:   jsii.String(stackName),
			Description: jsii.String("Agentic Identity Broker - Token Vault encryption infrastructure (KMS + DynamoDB + IAM)"),
			Env:         makeEnv(),
			Synthesizer: awscdk.NewDefaultStackSynthesizer(&awscdk.DefaultStackSynthesizerProps{
				GenerateBootstrapVersionRule: jsii.Bool(false),
			}),
		},
		Environment:           env,
		K8sNamespace:          k8sNamespace,
		K8sServiceAccountName: k8sServiceAccountName,
		Tags:                  customTags,
	})

	app.Synth(nil)
}

// makeEnv resolves the AWS environment from CDK_DEFAULT_ACCOUNT/CDK_DEFAULT_REGION
// or falls back to eu-central-1 if not set.
func makeEnv() *awscdk.Environment {
	account := os.Getenv("CDK_DEFAULT_ACCOUNT")
	region := os.Getenv("CDK_DEFAULT_REGION")

	if region == "" {
		region = "eu-central-1"
	}

	env := &awscdk.Environment{
		Region: jsii.String(region),
	}

	if account != "" {
		env.Account = jsii.String(account)
	}

	return env
}
