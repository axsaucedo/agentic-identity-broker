package main

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to create a stack for testing with IRSA parameters.
func createTestStack(t *testing.T, env string, oidcArn, namespace, sa string) (awscdk.Stack, assertions.Template) {
	t.Helper()
	app := awscdk.NewApp(nil)

	stack := NewEncryptionStack(app, "TestStack", &EncryptionStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String("123456789012"),
				Region:  jsii.String("eu-central-1"),
			},
		},
		Environment:           env,
		OIDCProviderArn:       oidcArn,
		K8sNamespace:          namespace,
		K8sServiceAccountName: sa,
	})

	template := assertions.Template_FromStack(stack, nil)
	return stack, template
}

// helper to create a stack with IRSA configuration for production testing.
func createTestStackIRSA(t *testing.T, env string, oidcArn, namespace, sa string) (awscdk.Stack, assertions.Template) {
	t.Helper()
	return createTestStack(t, env, oidcArn, namespace, sa)
}

// --- KMS Key Tests ---

func TestKMSKeyIsSymmetricWithRotation(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::KMS::Key"), map[string]interface{}{
		"KeySpec":           "SYMMETRIC_DEFAULT",
		"KeyUsage":          "ENCRYPT_DECRYPT",
		"EnableKeyRotation": true,
		"Description":       "Agentic Identity Broker - Token Vault KEK (test)",
	})
}

func TestKMSKeyAliasFollowsNamingConvention(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::KMS::Alias"), map[string]interface{}{
		"AliasName": "alias/agentic-identity-broker/test/token-vault-kek",
	})
}

func TestKMSKeyProdRetention(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE", "default", "test-sa")

	// Prod KMS key should have RETAIN deletion policy.
	templateJSON := template.ToJSON()
	resources := findResourcesByType(t, templateJSON, "AWS::KMS::Key")
	require.NotEmpty(t, resources, "expected at least one KMS key")

	for _, res := range resources {
		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Retain", resMap["DeletionPolicy"], "prod KMS key should have Retain deletion policy")
		assert.Equal(t, "Retain", resMap["UpdateReplacePolicy"], "prod KMS key should have Retain update replace policy")
	}
}

func TestKMSKeyNonProdDeletion(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	templateJSON := template.ToJSON()
	resources := findResourcesByType(t, templateJSON, "AWS::KMS::Key")
	require.NotEmpty(t, resources)

	for _, res := range resources {
		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Delete", resMap["DeletionPolicy"], "non-prod KMS key should have Delete deletion policy")
	}
}

// --- DynamoDB Table Tests ---

func TestDynamoDBTableSchema(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	// The AWS Encryption SDK KeyStore requires branch-key-id (S) + type (S).
	// This matches the schema used by the KeyStore client:
	// github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl/awscryptographykeystoresmithygenerated
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"KeySchema": []interface{}{
			map[string]interface{}{
				"AttributeName": "branch-key-id",
				"KeyType":       "HASH",
			},
			map[string]interface{}{
				"AttributeName": "type",
				"KeyType":       "RANGE",
			},
		},
		"AttributeDefinitions": []interface{}{
			map[string]interface{}{
				"AttributeName": "branch-key-id",
				"AttributeType": "S",
			},
			map[string]interface{}{
				"AttributeName": "type",
				"AttributeType": "S",
			},
		},
	})
}

func TestDynamoDBTablePayPerRequest(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"BillingMode": "PAY_PER_REQUEST",
	})
}

func TestDynamoDBTableNameFollowsConvention(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "AgenticIdentityBrokerBranchKeys-test",
	})
}

func TestDynamoDBTableProdDeletionProtection(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE", "default", "test-sa")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"DeletionProtectionEnabled": true,
		"PointInTimeRecoverySpecification": map[string]interface{}{
			"PointInTimeRecoveryEnabled": true,
		},
	})
}

func TestDynamoDBTableNonProdNoDeletionProtection(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	templateJSON := template.ToJSON()
	resources := findResourcesByType(t, templateJSON, "AWS::DynamoDB::Table")
	require.NotEmpty(t, resources)

	for _, res := range resources {
		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Delete", resMap["DeletionPolicy"], "non-prod DynamoDB table should have Delete deletion policy")
	}
}

// --- IAM Role Tests ---

func TestIAMRoleCreated(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "AgenticIdentityBrokerEncryptionRole-test",
	})
}

func TestIAMRoleTrustPolicyDefaultAccountRoot(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	// When no OIDC provider is specified, the role trusts the account root.
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"AssumeRolePolicyDocument": map[string]interface{}{
			"Statement": []interface{}{
				map[string]interface{}{
					"Action": "sts:AssumeRole",
					"Effect": "Allow",
					"Principal": map[string]interface{}{
						"AWS": map[string]interface{}{
							"Fn::Join": []interface{}{
								"",
								[]interface{}{
									"arn:",
									map[string]interface{}{"Ref": "AWS::Partition"},
									":iam::123456789012:root",
								},
							},
						},
					},
				},
			},
		},
	})
}

func TestIAMRoleTrustPolicyIRSA(t *testing.T) {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE"
	namespace := "identity-broker"
	sa := "agentic-identity-broker"

	_, template := createTestStack(t, "prod", oidcProviderArn, namespace, sa)

	// Verify federated principal with web identity trust policy
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"AssumeRolePolicyDocument": map[string]interface{}{
			"Statement": []interface{}{
				map[string]interface{}{
					"Action": "sts:AssumeRoleWithWebIdentity",
					"Effect": "Allow",
					"Principal": map[string]interface{}{
						"Federated": oidcProviderArn,
					},
					"Condition": map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:sub": "system:serviceaccount:identity-broker:agentic-identity-broker",
							"oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:aud": "sts.amazonaws.com",
						},
					},
				},
			},
		},
	})
}

func TestIAMRoleHasKMSPermissions(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": assertions.Match_ArrayWith(&[]interface{}{
				assertions.Match_ObjectLike(&map[string]interface{}{
					"Sid":    "KMSHierarchicalKeyringOperations",
					"Effect": "Allow",
					"Action": assertions.Match_ArrayWith(&[]interface{}{
						"kms:Encrypt",
						"kms:Decrypt",
						"kms:GenerateDataKey",
					}),
				}),
			}),
		},
	})
}

func TestIAMRoleHasDynamoDBPermissions(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": assertions.Match_ArrayWith(&[]interface{}{
				assertions.Match_ObjectLike(&map[string]interface{}{
					"Sid":    "DynamoDBBranchKeyOperations",
					"Effect": "Allow",
					"Action": assertions.Match_ArrayWith(&[]interface{}{
						"dynamodb:GetItem",
						"dynamodb:PutItem",
						"dynamodb:Query",
					}),
				}),
			}),
		},
	})
}

func TestIAMRoleHasCreateGrantPermission(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": assertions.Match_ArrayWith(&[]interface{}{
				assertions.Match_ObjectLike(&map[string]interface{}{
					"Sid":    "KMSCreateGrantForKeyring",
					"Effect": "Allow",
					"Action": "kms:CreateGrant",
					"Condition": map[string]interface{}{
						"Bool": map[string]interface{}{
							"kms:GrantIsForAWSResource": "true",
						},
					},
				}),
			}),
		},
	})
}

func TestIAMRoleMaxSessionDurationOneHour(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"MaxSessionDuration": 3600,
	})
}

// --- Stack Outputs Tests ---

func TestStackOutputsExist(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	templateJSON := template.ToJSON()
	outputsRaw := extractSection(t, templateJSON, "Outputs")
	outputs, ok := outputsRaw.(map[string]interface{})
	require.True(t, ok, "Outputs should be a map")

	// Verify expected output keys exist.
	expectedDescriptions := map[string]string{
		"IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN":             "EncryptionKeyARN",
		"IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_TABLE_NAME": "BranchKeyTableName",
		"IDENTITY_BROKER_ENCRYPTION_IAM_ROLE_ARN":                "EncryptionRoleARN",
	}

	for envVar, outputID := range expectedDescriptions {
		found := false
		for _, output := range outputs {
			outMap, ok := output.(map[string]interface{})
			if !ok {
				continue
			}
			desc, _ := outMap["Description"].(string)
			if desc != "" && containsSubstring(desc, envVar) {
				found = true
				break
			}
		}
		assert.True(t, found, "expected output referencing %s (logical ID: %s)", envVar, outputID)
	}
}

func TestStackOutputExportNames(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	templateJSON := template.ToJSON()
	outputsRaw := extractSection(t, templateJSON, "Outputs")
	outputs, ok := outputsRaw.(map[string]interface{})
	require.True(t, ok)

	expectedExports := []string{
		"AgenticIdentityBroker-test-EncryptionKeyARN",
		"AgenticIdentityBroker-test-BranchKeyTableName",
		"AgenticIdentityBroker-test-EncryptionRoleARN",
	}

	var exportNames []string
	for _, output := range outputs {
		outMap, ok := output.(map[string]interface{})
		if !ok {
			continue
		}
		if eMap, ok := outMap["Export"].(map[string]interface{}); ok {
			if name, ok := eMap["Name"].(string); ok {
				exportNames = append(exportNames, name)
			}
		}
	}

	for _, expected := range expectedExports {
		assert.Contains(t, exportNames, expected, "expected export %s", expected)
	}
}

// --- Resource Count Tests ---

func TestResourceCount(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.ResourceCountIs(jsii.String("AWS::KMS::Key"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("AWS::KMS::Alias"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("AWS::DynamoDB::Table"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("AWS::IAM::Role"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("AWS::CloudWatch::Alarm"), jsii.Number(2))
}

// --- Tags Tests ---

func TestRequiredTagsApplied(t *testing.T) {
	// Test that required tags are applied to resources.
	// This validates the tagging policy specified in the REMEDIATION_PLAN.md.
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE", "default", "test-sa")

	// Verify Project tag on DynamoDB table (critical for cost allocation and resource identification)
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "Project",
				"Value": "agentic-identity-broker",
			}),
		}),
	})

	// Verify Component tag on IAM role (critical for operational understanding)
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "Component",
				"Value": "encryption-vault",
			}),
		}),
	})

	// Verify ManagedBy tag on IAM role (shows infrastructure-as-code management)
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "ManagedBy",
				"Value": "aws-cdk",
			}),
		}),
	})

	// Verify environment is embedded in resource names (alternative to Environment tag)
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "AgenticIdentityBrokerBranchKeys-prod",
	})

	template.HasResourceProperties(jsii.String("AWS::KMS::Alias"), map[string]interface{}{
		"AliasName": "alias/agentic-identity-broker/prod/token-vault-kek",
	})

	t.Log("Required tagging policy validated: Project, Component, ManagedBy tags present; environment embedded in resource names")
}

func TestAllResourcesTagged(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	// DynamoDB table should be tagged (CDK reliably propagates tags to DynamoDB).
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "Environment",
				"Value": "test",
			}),
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "Project",
				"Value": "agentic-identity-broker",
			}),
		}),
	})

	// IAM role should be tagged.
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "Component",
				"Value": "encryption-vault",
			}),
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "ManagedBy",
				"Value": "aws-cdk",
			}),
		}),
	})
}

// --- CloudWatch Dashboard Tests ---

func TestDashboardCreated(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE", "default", "test-sa")

	template.ResourceCountIs(jsii.String("AWS::CloudWatch::Dashboard"), jsii.Number(1))
}

func TestDashboardNameFollowsConvention(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::CloudWatch::Dashboard"), map[string]interface{}{
		"DashboardName": "AgenticIdentityBroker-Encryption-test",
	})
}

// --- CloudWatch Alarms Tests ---

func TestKMSThrottleAlarmCreated(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE", "default", "test-sa")

	template.HasResourceProperties(jsii.String("AWS::CloudWatch::Alarm"), map[string]interface{}{
		"AlarmName":          "AgenticIdentityBroker-Encryption-prod-KMS-Throttle",
		"AlarmDescription":   "KMS key throttling detected - may indicate insufficient quota",
		"Threshold":          float64(10),
		"EvaluationPeriods":  float64(2),
		"ComparisonOperator": "GreaterThanThreshold",
		"TreatMissingData":   "notBreaching",
	})
}

func TestKMSErrorAlarmCreated(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE", "default", "test-sa")

	template.HasResourceProperties(jsii.String("AWS::CloudWatch::Alarm"), map[string]interface{}{
		"AlarmName":          "AgenticIdentityBroker-Encryption-prod-KMS-Errors",
		"AlarmDescription":   "KMS API errors detected - check key policy and permissions",
		"Threshold":          float64(5),
		"EvaluationPeriods":  float64(1),
		"ComparisonOperator": "GreaterThanThreshold",
		"TreatMissingData":   "notBreaching",
	})
}

func TestKMSAlarmsOutputsExist(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	templateJSON := template.ToJSON()
	outputsRaw := extractSection(t, templateJSON, "Outputs")
	outputs, ok := outputsRaw.(map[string]interface{})
	require.True(t, ok, "Outputs should be a map")

	// Verify alarm outputs exist
	var foundThrottleAlarm, foundErrorAlarm bool
	for _, output := range outputs {
		outMap, ok := output.(map[string]interface{})
		if !ok {
			continue
		}
		desc, _ := outMap["Description"].(string)
		if containsSubstring(desc, "KMS throttling") {
			foundThrottleAlarm = true
		}
		if containsSubstring(desc, "KMS errors") {
			foundErrorAlarm = true
		}
	}

	assert.True(t, foundThrottleAlarm, "expected KMS throttle alarm output")
	assert.True(t, foundErrorAlarm, "expected KMS error alarm output")
}

// --- Production Security Tests ---

func TestProductionRequiresOIDCProviderArn(t *testing.T) {
	// Should panic when env=prod and no oidcProviderArn
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for production without oidcProviderArn")
		}
	}()

	app := awscdk.NewApp(nil)
	NewEncryptionStack(app, "test", &EncryptionStackProps{
		Environment:           "prod",
		OIDCProviderArn:       "", // Missing!
		K8sNamespace:          "default",
		K8sServiceAccountName: "test-sa",
	})
}

func TestProductionRequiresK8sNamespace(t *testing.T) {
	// Should panic when env=prod and no k8sNamespace
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for production without k8sNamespace")
		}
	}()

	app := awscdk.NewApp(nil)
	NewEncryptionStack(app, "test", &EncryptionStackProps{
		Environment:           "prod",
		OIDCProviderArn:       "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE",
		K8sNamespace:          "", // Missing!
		K8sServiceAccountName: "test-sa",
	})
}

func TestProductionRequiresK8sServiceAccountName(t *testing.T) {
	// Should panic when env=prod and no k8sServiceAccountName
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for production without k8sServiceAccountName")
		}
	}()

	app := awscdk.NewApp(nil)
	NewEncryptionStack(app, "test", &EncryptionStackProps{
		Environment:           "prod",
		OIDCProviderArn:       "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE",
		K8sNamespace:          "default",
		K8sServiceAccountName: "", // Missing!
	})
}

func TestProductionWithIRSAParametersSucceeds(t *testing.T) {
	// Should NOT panic when env=prod WITH all IRSA parameters
	app := awscdk.NewApp(nil)
	stack := NewEncryptionStack(app, "test", &EncryptionStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String("123456789012"),
				Region:  jsii.String("eu-central-1"),
			},
		},
		Environment:           "prod",
		OIDCProviderArn:       "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE",
		K8sNamespace:          "identity-broker",
		K8sServiceAccountName: "agentic-identity-broker",
	})
	assert.NotNil(t, stack)
}

func TestNonProductionAllowsEmptyIRSAParameters(t *testing.T) {
	// Should NOT panic for test environment without IRSA parameters
	app := awscdk.NewApp(nil)
	stack := NewEncryptionStack(app, "test", &EncryptionStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String("123456789012"),
				Region:  jsii.String("eu-central-1"),
			},
		},
		Environment:           "test",
		OIDCProviderArn:       "",
		K8sNamespace:          "",
		K8sServiceAccountName: "",
	})
	assert.NotNil(t, stack)
}

// --- IRSA-Specific Tests ---

func TestIRSAConditions(t *testing.T) {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/TESTID12345"
	namespace := "my-namespace"
	sa := "my-service-account"

	_, template := createTestStack(t, "prod", oidcProviderArn, namespace, sa)

	// Verify IRSA conditions are correctly formatted
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"AssumeRolePolicyDocument": map[string]interface{}{
			"Statement": []interface{}{
				map[string]interface{}{
					"Action": "sts:AssumeRoleWithWebIdentity",
					"Effect": "Allow",
					"Principal": map[string]interface{}{
						"Federated": oidcProviderArn,
					},
					"Condition": map[string]interface{}{
						"StringEquals": map[string]interface{}{
							"oidc.eks.us-west-2.amazonaws.com/id/TESTID12345:sub": "system:serviceaccount:my-namespace:my-service-account",
							"oidc.eks.us-west-2.amazonaws.com/id/TESTID12345:aud": "sts.amazonaws.com",
						},
					},
				},
			},
		},
	})
}

func TestIRSAOutputs(t *testing.T) {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE"
	namespace := "identity-broker"
	sa := "agentic-identity-broker"

	_, template := createTestStack(t, "prod", oidcProviderArn, namespace, sa)

	// Verify IRSA-specific outputs exist
	templateJSON := template.ToJSON()
	outputsRaw := extractSection(t, templateJSON, "Outputs")
	outputs, ok := outputsRaw.(map[string]interface{})
	require.True(t, ok, "Outputs should be a map")

	// Check for IRSA outputs
	var foundNamespace, foundSAName, foundFullName, foundRoleName bool
	for _, output := range outputs {
		outMap, ok := output.(map[string]interface{})
		if !ok {
			continue
		}
		desc, _ := outMap["Description"].(string)
		if containsSubstring(desc, "namespace for service account") {
			foundNamespace = true
		}
		if containsSubstring(desc, "service account name") && !containsSubstring(desc, "Full") {
			foundSAName = true
		}
		if containsSubstring(desc, "Full service account reference") {
			foundFullName = true
		}
		if containsSubstring(desc, "iam.amazonaws.com/role annotation") {
			foundRoleName = true
		}
	}

	assert.True(t, foundNamespace, "expected ServiceAccountNamespace output")
	assert.True(t, foundSAName, "expected ServiceAccountName output")
	assert.True(t, foundFullName, "expected ServiceAccountFullName output")
	assert.True(t, foundRoleName, "expected IamRoleName output")
}

func TestIRSAConditionKeysFormatting(t *testing.T) {
	// Test various OIDC provider ARN formats to ensure condition keys are extracted correctly
	testCases := []struct {
		name            string
		oidcArn         string
		expectedHost    string
		namespace       string
		sa              string
		expectedSubject string
	}{
		{
			name:            "us-east-1 EKS OIDC provider",
			oidcArn:         "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLE1",
			expectedHost:    "oidc.eks.us-east-1.amazonaws.com/id/EXAMPLE1",
			namespace:       "default",
			sa:              "test-sa",
			expectedSubject: "system:serviceaccount:default:test-sa",
		},
		{
			name:            "eu-west-1 EKS OIDC provider",
			oidcArn:         "arn:aws:iam::987654321098:oidc-provider/oidc.eks.eu-west-1.amazonaws.com/id/EXAMPLE2",
			expectedHost:    "oidc.eks.eu-west-1.amazonaws.com/id/EXAMPLE2",
			namespace:       "production",
			sa:              "identity-broker",
			expectedSubject: "system:serviceaccount:production:identity-broker",
		},
		{
			name:            "ap-southeast-2 EKS OIDC provider",
			oidcArn:         "arn:aws:iam::111111111111:oidc-provider/oidc.eks.ap-southeast-2.amazonaws.com/id/EXAMPLE3",
			expectedHost:    "oidc.eks.ap-southeast-2.amazonaws.com/id/EXAMPLE3",
			namespace:       "kube-system",
			sa:              "aws-load-balancer-controller",
			expectedSubject: "system:serviceaccount:kube-system:aws-load-balancer-controller",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, template := createTestStack(t, "prod", tc.oidcArn, tc.namespace, tc.sa)

			// Verify the condition keys are formatted correctly
			template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
				"AssumeRolePolicyDocument": map[string]interface{}{
					"Statement": []interface{}{
						map[string]interface{}{
							"Action": "sts:AssumeRoleWithWebIdentity",
							"Effect": "Allow",
							"Principal": map[string]interface{}{
								"Federated": tc.oidcArn,
							},
							"Condition": map[string]interface{}{
								"StringEquals": map[string]interface{}{
									tc.expectedHost + ":sub": tc.expectedSubject,
									tc.expectedHost + ":aud": "sts.amazonaws.com",
								},
							},
						},
					},
				},
			})
		})
	}
}

// --- Template Metadata Tests ---

func TestTemplateMetadataHeader(t *testing.T) {
	// CI/CD tooling requires Metadata.Application, Environment, and StackName
	// in the synthesized CloudFormation template to validate deployment manifests.
	_, template := createTestStack(t, "test", "", "", "")

	templateJSON := template.ToJSON()
	metadataRaw := extractSection(t, templateJSON, "Metadata")
	metadata, ok := metadataRaw.(map[string]interface{})
	require.True(t, ok, "template Metadata section should be present")

	application, ok := metadata["Application"].(string)
	require.True(t, ok, "Metadata.Application should be a string")
	assert.Equal(t, "agentic-identity-broker", application)

	environment, ok := metadata["Environment"].(string)
	require.True(t, ok, "Metadata.Environment should be a string")
	assert.Equal(t, "test", environment)

	stackName, ok := metadata["StackName"].(string)
	require.True(t, ok, "Metadata.StackName should be a string")
	assert.Equal(t, "TestStack", stackName, "Metadata.StackName should match the stack logical ID")
}

// --- Environment Parameterization Tests ---

func TestEnvironmentParameterizationTest(t *testing.T) {
	_, template := createTestStack(t, "test", "", "", "")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "AgenticIdentityBrokerBranchKeys-test",
	})
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "AgenticIdentityBrokerEncryptionRole-test",
	})
}

func TestEnvironmentParameterizationProd(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE", "default", "test-sa")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "AgenticIdentityBrokerBranchKeys-prod",
	})
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "AgenticIdentityBrokerEncryptionRole-prod",
	})
}

// --- Helpers ---

// findResourcesByType extracts all CloudFormation resources of a given type.
func findResourcesByType(t *testing.T, templateJSON interface{}, resourceType string) []interface{} {
	t.Helper()

	resources := extractSection(t, templateJSON, "Resources")
	resMap, ok := resources.(map[string]interface{})
	require.True(t, ok)

	var result []interface{}
	for _, res := range resMap {
		r, ok := res.(map[string]interface{})
		if !ok {
			continue
		}
		if r["Type"] == resourceType {
			result = append(result, res)
		}
	}
	return result
}

// extractSection extracts a top-level section from the template JSON.
func extractSection(t *testing.T, templateJSON interface{}, section string) interface{} {
	t.Helper()

	// templateJSON from CDK assertions is already a map.
	m, ok := templateJSON.(*map[string]interface{})
	if ok {
		return (*m)[section]
	}

	// Try as direct map.
	if dm, ok := templateJSON.(map[string]interface{}); ok {
		return dm[section]
	}

	// Last resort: marshal/unmarshal.
	b, err := json.Marshal(templateJSON)
	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(b, &result)
	require.NoError(t, err)
	return result[section]
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
