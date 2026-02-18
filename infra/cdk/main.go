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
//	cdk deploy -c env=prod \
//	  -c oidcProviderArn=arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLEID \
//	  -c k8sNamespace=default \
//	  -c k8sServiceAccountName=agentic-identity-broker
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

	// Read IRSA parameters for Kubernetes IAM Roles for Service Accounts.
	var oidcProviderArn string
	if v := app.Node().TryGetContext(jsii.String("oidcProviderArn")); v != nil {
		if s, ok := v.(string); ok {
			oidcProviderArn = s
		}
	}

	var k8sNamespace string
	if v := app.Node().TryGetContext(jsii.String("k8sNamespace")); v != nil {
		if s, ok := v.(string); ok {
			k8sNamespace = s
		}
	}

	var k8sServiceAccountName string
	if v := app.Node().TryGetContext(jsii.String("k8sServiceAccountName")); v != nil {
		if s, ok := v.(string); ok {
			k8sServiceAccountName = s
		}
	}

	// Validate IRSA parameters for production deployments.
	isProd := env == "prod" || env == "production"
	if isProd {
		if oidcProviderArn == "" {
			panic("ERROR: Production deployments require oidcProviderArn.\n" +
				"Usage: cdk deploy -c env=prod -c oidcProviderArn=arn:aws:iam::ACCOUNT:oidc-provider/oidc.eks.REGION.amazonaws.com/id/ID\n" +
				"Example: cdk deploy -c env=prod -c oidcProviderArn=arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE")
		}
		if k8sNamespace == "" {
			panic("ERROR: Production deployments require k8sNamespace.\n" +
				"Usage: cdk deploy -c env=prod -c k8sNamespace=default\n" +
				"Example: cdk deploy -c env=prod -c k8sNamespace=identity-broker")
		}
		if k8sServiceAccountName == "" {
			panic("ERROR: Production deployments require k8sServiceAccountName.\n" +
				"Usage: cdk deploy -c env=prod -c k8sServiceAccountName=agentic-identity-broker\n" +
				"Example: cdk deploy -c env=prod -c k8sServiceAccountName=identity-broker-sa")
		}
	}

	stackName := "AgenticIdentityBrokerEncryption-" + env

	NewEncryptionStack(app, stackName, &EncryptionStackProps{
		StackProps: awscdk.StackProps{
			StackName:   jsii.String(stackName),
			Description: jsii.String("Agentic Identity Broker - Token Vault encryption infrastructure (KMS + DynamoDB + IAM) with IRSA support"),
			Env:         makeEnv(),
			Synthesizer: awscdk.NewDefaultStackSynthesizer(&awscdk.DefaultStackSynthesizerProps{
				GenerateBootstrapVersionRule: jsii.Bool(false),
			}),
			Tags: &map[string]*string{
				"Project":     jsii.String("agentic-identity-broker"),
				"Component":   jsii.String("encryption"),
				"Environment": jsii.String(env),
				"ManagedBy":   jsii.String("aws-cdk"),
			},
		},
		Environment:           env,
		OIDCProviderArn:       oidcProviderArn,
		K8sNamespace:          k8sNamespace,
		K8sServiceAccountName: k8sServiceAccountName,
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
