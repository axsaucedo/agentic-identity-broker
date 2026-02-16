// Package main is the AWS CDK application entrypoint for the
// Agentic Identity Broker encryption infrastructure.
//
// It provisions KMS, DynamoDB, and IAM resources required by the
// Token Vault's hierarchical keyring envelope encryption.
//
// Usage:
//
//	cdk deploy -c env=dev
//	cdk deploy -c env=staging
//	cdk deploy -c env=prod
//	cdk deploy -c env=prod -c trustPrincipal=arn:aws:iam::123456789012:role/ECSTaskRole
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

	// Read environment from CDK context (default: "dev").
	env := "dev"
	if v := app.Node().TryGetContext(jsii.String("env")); v != nil {
		if s, ok := v.(string); ok {
			env = s
		}
	}

	// Validate environment parameter to prevent typos and unexpected stack creation.
	validEnvs := map[string]bool{"dev": true, "staging": true, "prod": true, "production": true}
	if !validEnvs[env] {
		panic(fmt.Sprintf("Invalid environment '%s'. Must be one of: dev, staging, prod, production", env))
	}

	// Normalize "production" to "prod" for consistency.
	if env == "production" {
		env = "prod"
	}

	// Read optional trust principal ARN for IAM role (generic compute support).
	var trustPrincipal string
	if v := app.Node().TryGetContext(jsii.String("trustPrincipal")); v != nil {
		if s, ok := v.(string); ok {
			trustPrincipal = s
		}
	}

	stackName := "IdentityBrokerEncryption-" + env

	NewEncryptionStack(app, stackName, &EncryptionStackProps{
		StackProps: awscdk.StackProps{
			StackName:   jsii.String(stackName),
			Description: jsii.String("Agentic Identity Broker - Token Vault encryption infrastructure (KMS + DynamoDB + IAM)"),
			Env:         makeEnv(),
			Tags: &map[string]*string{
				"Project":     jsii.String("agentic-identity-broker"),
				"Component":   jsii.String("encryption"),
				"Environment": jsii.String(env),
				"ManagedBy":   jsii.String("aws-cdk"),
			},
		},
		Environment:    env,
		TrustPrincipal: trustPrincipal,
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
