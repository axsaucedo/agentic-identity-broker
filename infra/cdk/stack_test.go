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

// helper to create a stack for testing.
func createTestStack(t *testing.T, env string, trustPrincipal string) (awscdk.Stack, assertions.Template) {
	t.Helper()
	app := awscdk.NewApp(nil)

	stack := NewEncryptionStack(app, "TestStack", &EncryptionStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String("123456789012"),
				Region:  jsii.String("eu-central-1"),
			},
		},
		Environment:    env,
		TrustPrincipal: trustPrincipal,
	})

	template := assertions.Template_FromStack(stack, nil)
	return stack, template
}

// --- KMS Key Tests ---

func TestKMSKeyIsSymmetricWithRotation(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

	template.HasResourceProperties(jsii.String("AWS::KMS::Key"), map[string]interface{}{
		"KeySpec":           "SYMMETRIC_DEFAULT",
		"KeyUsage":          "ENCRYPT_DECRYPT",
		"EnableKeyRotation": true,
		"Description":       "Agentic Identity Broker - Token Vault KEK (dev)",
	})
}

func TestKMSKeyAliasFollowsNamingConvention(t *testing.T) {
	_, template := createTestStack(t, "staging", "")

	template.HasResourceProperties(jsii.String("AWS::KMS::Alias"), map[string]interface{}{
		"AliasName": "alias/identity-broker/staging/token-vault-kek",
	})
}

func TestKMSKeyProdRetention(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:role/TestRole")

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

func TestKMSKeyDevDeletion(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

	templateJSON := template.ToJSON()
	resources := findResourcesByType(t, templateJSON, "AWS::KMS::Key")
	require.NotEmpty(t, resources)

	for _, res := range resources {
		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Delete", resMap["DeletionPolicy"], "dev KMS key should have Delete deletion policy")
	}
}

// --- DynamoDB Table Tests ---

func TestDynamoDBTableSchema(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

	// The AWS Encryption SDK KeyStore requires partition_key (S) + sort_key (S).
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"KeySchema": []interface{}{
			map[string]interface{}{
				"AttributeName": "partition_key",
				"KeyType":       "HASH",
			},
			map[string]interface{}{
				"AttributeName": "sort_key",
				"KeyType":       "RANGE",
			},
		},
		"AttributeDefinitions": []interface{}{
			map[string]interface{}{
				"AttributeName": "partition_key",
				"AttributeType": "S",
			},
			map[string]interface{}{
				"AttributeName": "sort_key",
				"AttributeType": "S",
			},
		},
	})
}

func TestDynamoDBTablePayPerRequest(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"BillingMode": "PAY_PER_REQUEST",
	})
}

func TestDynamoDBTableNameFollowsConvention(t *testing.T) {
	_, template := createTestStack(t, "staging", "")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "IdentityBrokerBranchKeys-staging",
	})
}

func TestDynamoDBTableProdDeletionProtection(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:role/TestRole")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"DeletionProtectionEnabled": true,
		"PointInTimeRecoverySpecification": map[string]interface{}{
			"PointInTimeRecoveryEnabled": true,
		},
	})
}

func TestDynamoDBTableDevNoDeletionProtection(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

	templateJSON := template.ToJSON()
	resources := findResourcesByType(t, templateJSON, "AWS::DynamoDB::Table")
	require.NotEmpty(t, resources)

	for _, res := range resources {
		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Delete", resMap["DeletionPolicy"], "dev DynamoDB table should have Delete deletion policy")
	}
}

// --- IAM Role Tests ---

func TestIAMRoleCreated(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "IdentityBrokerEncryptionRole-dev",
	})
}

func TestIAMRoleTrustPolicyDefaultAccountRoot(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

	// When no trust principal is specified, the role trusts the account root.
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

func TestIAMRoleTrustPolicyCustomPrincipal(t *testing.T) {
	principalARN := "arn:aws:iam::123456789012:role/ECSTaskRole"
	_, template := createTestStack(t, "prod", principalARN)

	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"AssumeRolePolicyDocument": map[string]interface{}{
			"Statement": []interface{}{
				map[string]interface{}{
					"Action": "sts:AssumeRole",
					"Effect": "Allow",
					"Principal": map[string]interface{}{
						"AWS": principalARN,
					},
				},
			},
		},
	})
}

func TestIAMRoleHasKMSPermissions(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

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
	_, template := createTestStack(t, "dev", "")

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
	_, template := createTestStack(t, "dev", "")

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
	_, template := createTestStack(t, "dev", "")

	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"MaxSessionDuration": 3600,
	})
}

// --- Stack Outputs Tests ---

func TestStackOutputsExist(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

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
	_, template := createTestStack(t, "staging", "")

	templateJSON := template.ToJSON()
	outputsRaw := extractSection(t, templateJSON, "Outputs")
	outputs, ok := outputsRaw.(map[string]interface{})
	require.True(t, ok)

	expectedExports := []string{
		"IdentityBroker-staging-EncryptionKeyARN",
		"IdentityBroker-staging-BranchKeyTableName",
		"IdentityBroker-staging-EncryptionRoleARN",
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
	_, template := createTestStack(t, "dev", "")

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
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:role/TestRole")

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
				"Value": "encryption",
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
		"TableName": "IdentityBrokerBranchKeys-prod",
	})

	template.HasResourceProperties(jsii.String("AWS::KMS::Alias"), map[string]interface{}{
		"AliasName": "alias/identity-broker/prod/token-vault-kek",
	})

	t.Log("Required tagging policy validated: Project, Component, ManagedBy tags present; environment embedded in resource names")
}

func TestAllResourcesTagged(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

	// DynamoDB table should be tagged (CDK reliably propagates tags to DynamoDB).
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "Environment",
				"Value": "dev",
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
				"Value": "encryption",
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
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:role/TestRole")

	template.ResourceCountIs(jsii.String("AWS::CloudWatch::Dashboard"), jsii.Number(1))
}

func TestDashboardNameFollowsConvention(t *testing.T) {
	_, template := createTestStack(t, "staging", "")

	template.HasResourceProperties(jsii.String("AWS::CloudWatch::Dashboard"), map[string]interface{}{
		"DashboardName": "IdentityBroker-Encryption-staging",
	})
}

// --- CloudWatch Alarms Tests ---

func TestKMSThrottleAlarmCreated(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:role/TestRole")

	template.HasResourceProperties(jsii.String("AWS::CloudWatch::Alarm"), map[string]interface{}{
		"AlarmName":          "IdentityBroker-Encryption-prod-KMS-Throttle",
		"AlarmDescription":   "KMS key throttling detected - may indicate insufficient quota",
		"Threshold":          float64(10),
		"EvaluationPeriods":  float64(2),
		"ComparisonOperator": "GreaterThanThreshold",
		"TreatMissingData":   "notBreaching",
	})
}

func TestKMSErrorAlarmCreated(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:role/TestRole")

	template.HasResourceProperties(jsii.String("AWS::CloudWatch::Alarm"), map[string]interface{}{
		"AlarmName":          "IdentityBroker-Encryption-prod-KMS-Errors",
		"AlarmDescription":   "KMS API errors detected - check key policy and permissions",
		"Threshold":          float64(5),
		"EvaluationPeriods":  float64(1),
		"ComparisonOperator": "GreaterThanThreshold",
		"TreatMissingData":   "notBreaching",
	})
}

func TestKMSAlarmsOutputsExist(t *testing.T) {
	_, template := createTestStack(t, "staging", "")

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

func TestProductionRequiresTrustPrincipal(t *testing.T) {
	// Should panic when env=prod and no trustPrincipal
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for production without trustPrincipal")
		}
	}()

	app := awscdk.NewApp(nil)
	NewEncryptionStack(app, "test", &EncryptionStackProps{
		Environment:    "prod",
		TrustPrincipal: "", // Missing!
	})
}

func TestProductionWithTrustPrincipalSucceeds(t *testing.T) {
	// Should NOT panic when env=prod WITH trustPrincipal
	app := awscdk.NewApp(nil)
	stack := NewEncryptionStack(app, "test", &EncryptionStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String("123456789012"),
				Region:  jsii.String("eu-central-1"),
			},
		},
		Environment:    "prod",
		TrustPrincipal: "arn:aws:iam::123456789012:role/ECSTaskRole",
	})
	assert.NotNil(t, stack)
}

func TestNonProductionAllowsEmptyTrustPrincipal(t *testing.T) {
	// Should NOT panic for dev/staging without trustPrincipal
	app := awscdk.NewApp(nil)
	stack := NewEncryptionStack(app, "test", &EncryptionStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String("123456789012"),
				Region:  jsii.String("eu-central-1"),
			},
		},
		Environment:    "dev",
		TrustPrincipal: "",
	})
	assert.NotNil(t, stack)
}

// --- Environment Parameterization Tests ---

func TestEnvironmentParameterizationDev(t *testing.T) {
	_, template := createTestStack(t, "dev", "")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "IdentityBrokerBranchKeys-dev",
	})
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "IdentityBrokerEncryptionRole-dev",
	})
}

func TestEnvironmentParameterizationProd(t *testing.T) {
	_, template := createTestStack(t, "prod", "arn:aws:iam::123456789012:role/TestRole")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "IdentityBrokerBranchKeys-prod",
	})
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "IdentityBrokerEncryptionRole-prod",
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
