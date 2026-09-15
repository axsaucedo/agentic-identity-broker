package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testOIDCProviderArn          = "arn:aws:iam::123456789012:oidc-provider/oidc.eks.eu-central-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE"
	testOIDCSubjectKey           = "oidc.eks.eu-central-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:sub"
	defaultServiceAccountSubject = "system:serviceaccount:agentic-identity-broker:agentic-identity-broker"
)

// helper to create a stack for testing.
// An empty subject defaults to the canonical test service account.
func createTestStack(t *testing.T, env, subject string) (awscdk.Stack, assertions.Template) {
	t.Helper()
	app := awscdk.NewApp(nil)

	stack := NewEncryptionStack(app, "TestStack", newTestStackProps(env, subject))

	template := assertions.Template_FromStack(stack, nil)
	return stack, template
}

// --- KMS Key Tests ---

func TestKMSKeyIsSymmetricWithRotation(t *testing.T) {
	_, template := createTestStack(t, "test", "")

	template.HasResourceProperties(jsii.String("AWS::KMS::Key"), map[string]interface{}{
		"KeySpec":           "SYMMETRIC_DEFAULT",
		"KeyUsage":          "ENCRYPT_DECRYPT",
		"EnableKeyRotation": true,
		"Description":       "Agentic Identity Broker - Token Vault KEK (test)",
	})
}

func TestKMSKeyAliasFollowsNamingConvention(t *testing.T) {
	_, template := createTestStack(t, "test", "")

	template.HasResourceProperties(jsii.String("AWS::KMS::Alias"), map[string]interface{}{
		"AliasName": "alias/agentic-identity-broker/test/token-vault-kek",
	})
}

func TestKMSKeyProdRetention(t *testing.T) {
	_, template := createTestStack(t, "prod", "system:serviceaccount:default:test-sa")

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
	_, template := createTestStack(t, "test", "")

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
	_, template := createTestStack(t, "test", "")

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
	_, template := createTestStack(t, "test", "")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"BillingMode": "PAY_PER_REQUEST",
	})
}

func TestDynamoDBTableNameFollowsConvention(t *testing.T) {
	_, template := createTestStack(t, "test", "")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "AgenticIdentityBrokerBranchKeys-test",
	})
}

func TestDynamoDBTableProdDeletionProtection(t *testing.T) {
	_, template := createTestStack(t, "prod", "system:serviceaccount:default:test-sa")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"DeletionProtectionEnabled": true,
		"PointInTimeRecoverySpecification": map[string]interface{}{
			"PointInTimeRecoveryEnabled": true,
		},
	})
}

func TestDynamoDBTableNonProdNoDeletionProtection(t *testing.T) {
	_, template := createTestStack(t, "test", "")

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
	_, template := createTestStack(t, "test", "")

	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "AgenticIdentityBrokerEncryptionRole-test",
	})
}

func TestIAMRoleTrustPolicyOIDCForTest(t *testing.T) {
	_, template := createTestStack(t, "test", "")

	templateJSON := template.ToJSON()
	roles := findResourcesByType(t, templateJSON, "AWS::IAM::Role")
	require.NotEmpty(t, roles)

	for _, role := range roles {
		stmt := extractTrustStatement(t, role)
		principal, ok := stmt["Principal"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, testOIDCProviderArn, principal["Federated"])
		assert.Equal(t, "sts:AssumeRoleWithWebIdentity", stmt["Action"])
	}
}

func TestIAMRoleTrustPolicyOIDC(t *testing.T) {
	subject := "system:serviceaccount:identity-broker:agentic-identity-broker"

	_, template := createTestStack(t, "prod", subject)

	templateJSON := template.ToJSON()
	roles := findResourcesByType(t, templateJSON, "AWS::IAM::Role")
	require.NotEmpty(t, roles)

	for _, role := range roles {
		stmt := extractTrustStatement(t, role)
		principal, ok := stmt["Principal"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, testOIDCProviderArn, principal["Federated"])
		condition, ok := stmt["Condition"].(map[string]interface{})
		require.True(t, ok)
		stringEquals, ok := condition["StringEquals"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, subject, stringEquals[testOIDCSubjectKey])
	}
}

func TestIAMRoleHasKMSPermissions(t *testing.T) {
	_, template := createTestStack(t, "test", "")

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
	_, template := createTestStack(t, "test", "")

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
	_, template := createTestStack(t, "test", "")

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
	_, template := createTestStack(t, "test", "")

	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"MaxSessionDuration": 3600,
	})
}

// --- Stack Outputs Tests ---

func TestStackOutputsExist(t *testing.T) {
	_, template := createTestStack(t, "test", "")

	templateJSON := template.ToJSON()
	outputsRaw := extractSection(t, templateJSON, "Outputs")
	outputs, ok := outputsRaw.(map[string]interface{})
	require.True(t, ok, "Outputs should be a map")

	// Verify expected output keys exist.
	expectedDescriptions := map[string]string{
		"IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN":             "EncryptionKeyARN",
		"IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_TABLE_NAME": "BranchKeyTableName",
	}

	for envVar, outputID := range expectedDescriptions {
		found := false
		for _, output := range outputs {
			outMap, ok := output.(map[string]interface{})
			if !ok {
				continue
			}
			desc, _ := outMap["Description"].(string)
			if desc != "" && strings.Contains(desc, envVar) {
				found = true
				break
			}
		}
		assert.True(t, found, "expected output referencing %s (logical ID: %s)", envVar, outputID)
	}
}

func TestStackOutputExportNames(t *testing.T) {
	_, template := createTestStack(t, "test", "")

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
	_, template := createTestStack(t, "test", "")

	template.ResourceCountIs(jsii.String("AWS::KMS::Key"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("AWS::KMS::Alias"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("AWS::DynamoDB::Table"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("AWS::IAM::Role"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("AWS::CloudWatch::Alarm"), jsii.Number(2))
}

// --- Tags Tests ---

func TestRequiredTagsApplied(t *testing.T) {
	// Test that required tags are applied to resources.
	_, template := createTestStack(t, "prod", "system:serviceaccount:default:test-sa")

	// Verify application tag on DynamoDB table (critical for cost allocation and resource identification)
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "application",
				"Value": "agentic-identity-broker",
			}),
		}),
	})

	// Verify component tag on IAM role (critical for operational understanding)
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "component",
				"Value": "encryption-vault",
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

}

func TestAllResourcesTagged(t *testing.T) {
	_, template := createTestStack(t, "test", "")

	// DynamoDB table should be tagged (CDK reliably propagates tags to DynamoDB).
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "application",
				"Value": "agentic-identity-broker",
			}),
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "environment",
				"Value": "test",
			}),
		}),
	})

	// IAM role should be tagged.
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "component",
				"Value": "encryption-vault",
			}),
		}),
	})
}

func TestEnvironmentTagCannotBeOverriddenByCustomTags(t *testing.T) {
	app := awscdk.NewApp(nil)
	props := newTestStackProps("prod", "system:serviceaccount:ns:sa")
	props.Tags = map[string]string{"environment": "staging"}
	stack := NewEncryptionStack(app, "TestStack", props)
	template := assertions.Template_FromStack(stack, nil)

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"Tags": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"Key":   "environment",
				"Value": "prod",
			}),
		}),
	})
}

// --- CloudWatch Dashboard Tests ---

func TestDashboardCreated(t *testing.T) {
	_, template := createTestStack(t, "prod", "system:serviceaccount:default:test-sa")

	template.ResourceCountIs(jsii.String("AWS::CloudWatch::Dashboard"), jsii.Number(1))
}

func TestDashboardNameFollowsConvention(t *testing.T) {
	_, template := createTestStack(t, "test", "")

	template.HasResourceProperties(jsii.String("AWS::CloudWatch::Dashboard"), map[string]interface{}{
		"DashboardName": "AgenticIdentityBroker-Encryption-test",
	})
}

func TestDashboardKMSWidgetUsesThrottleCount(t *testing.T) {
	_, template := createTestStack(t, "test", "")

	templateJSON := template.ToJSON()
	dashboards := findResourcesByType(t, templateJSON, "AWS::CloudWatch::Dashboard")
	require.NotEmpty(t, dashboards)

	for _, d := range dashboards {
		dMap, ok := d.(map[string]interface{})
		require.True(t, ok)
		props, ok := dMap["Properties"].(map[string]interface{})
		require.True(t, ok)
		bodyBytes, err := json.Marshal(props["DashboardBody"])
		require.NoError(t, err)
		body := string(bodyBytes)
		assert.Contains(t, body, "ThrottleCount", "KMS throttle widget must use ThrottleCount metric")
		assert.NotContains(t, body, "UserErrorCount", "KMS widget must not use UserErrorCount (belongs to alarms only)")
	}
}

// --- CloudWatch Alarms Tests ---

func TestKMSThrottleAlarmCreated(t *testing.T) {
	_, template := createTestStack(t, "prod", "system:serviceaccount:default:test-sa")

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
	_, template := createTestStack(t, "prod", "system:serviceaccount:default:test-sa")

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
	_, template := createTestStack(t, "test", "")

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
		if strings.Contains(desc, "KMS throttling") {
			foundThrottleAlarm = true
		}
		if strings.Contains(desc, "KMS errors") {
			foundErrorAlarm = true
		}
	}

	assert.True(t, foundThrottleAlarm, "expected KMS throttle alarm output")
	assert.True(t, foundErrorAlarm, "expected KMS error alarm output")
}

// --- Production Security Tests ---

func TestProductionRequiresServiceAccountSubject(t *testing.T) {
	app := awscdk.NewApp(nil)
	props := newTestStackProps("prod", defaultServiceAccountSubject)
	props.ServiceAccountSubject = ""

	assert.Panics(t, func() {
		NewEncryptionStack(app, "test", props)
	})
}

func TestProductionWithServiceAccountSucceeds(t *testing.T) {
	app := awscdk.NewApp(nil)
	stack := NewEncryptionStack(app, "test", newTestStackProps("prod", defaultServiceAccountSubject))
	assert.NotNil(t, stack)
}

func TestNonProductionRequiresServiceAccountSubject(t *testing.T) {
	app := awscdk.NewApp(nil)
	props := newTestStackProps("test", defaultServiceAccountSubject)
	props.ServiceAccountSubject = ""

	assert.Panics(t, func() {
		NewEncryptionStack(app, "test", props)
	})
}

func TestRequiresOIDCTrustParameters(t *testing.T) {
	t.Run("missing provider arn", func(t *testing.T) {
		app := awscdk.NewApp(nil)
		props := newTestStackProps("prod", defaultServiceAccountSubject)
		props.OIDCProviderArn = ""

		assert.Panics(t, func() {
			NewEncryptionStack(app, "test", props)
		})
	})

	t.Run("missing subject key", func(t *testing.T) {
		app := awscdk.NewApp(nil)
		props := newTestStackProps("prod", defaultServiceAccountSubject)
		props.OIDCSubjectKey = ""

		assert.Panics(t, func() {
			NewEncryptionStack(app, "test", props)
		})
	})
}

// --- OIDC Trust Policy Tests ---

func TestOIDCTrustPolicyServiceAccount(t *testing.T) {
	subject := "system:serviceaccount:my-namespace:my-service-account"

	_, template := createTestStack(t, "prod", subject)

	templateJSON := template.ToJSON()
	roles := findResourcesByType(t, templateJSON, "AWS::IAM::Role")
	require.NotEmpty(t, roles)

	for _, role := range roles {
		stmt := extractTrustStatement(t, role)
		condition, ok := stmt["Condition"].(map[string]interface{})
		require.True(t, ok)
		stringEquals, ok := condition["StringEquals"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, subject, stringEquals[testOIDCSubjectKey])
	}
}

func TestServiceAccountOutputs(t *testing.T) {
	_, template := createTestStack(t, "prod", "system:serviceaccount:identity-broker:agentic-identity-broker")

	templateJSON := template.ToJSON()
	outputsRaw := extractSection(t, templateJSON, "Outputs")
	outputs, ok := outputsRaw.(map[string]interface{})
	require.True(t, ok, "Outputs should be a map")

	var foundRoleName bool
	for _, output := range outputs {
		outMap, ok := output.(map[string]interface{})
		if !ok {
			continue
		}
		if desc, _ := outMap["Description"].(string); strings.Contains(desc, "iam.amazonaws.com/role annotation") {
			foundRoleName = true
		}
	}

	assert.True(t, foundRoleName, "expected IamRoleName output")
}

// --- Template Metadata Tests ---

func TestTemplateMetadataHeader(t *testing.T) {
	// CI/CD pipeline requires Metadata.StackName and Metadata.Tags.application
	// in the synthesized CloudFormation template to identify deployment manifests.
	_, template := createTestStack(t, "test", "")

	templateJSON := template.ToJSON()
	metadataRaw := extractSection(t, templateJSON, "Metadata")
	metadata, ok := metadataRaw.(map[string]interface{})
	require.True(t, ok, "template Metadata section should be present")

	stackName, ok := metadata["StackName"].(string)
	require.True(t, ok, "Metadata.StackName should be a string")
	assert.Equal(t, "TestStack", stackName, "Metadata.StackName should match the stack logical ID")

	tags, ok := metadata["Tags"].(map[string]interface{})
	require.True(t, ok, "Metadata.Tags should be a map")

	application, ok := tags["application"].(string)
	require.True(t, ok, "Metadata.Tags.application should be a string")
	assert.Equal(t, "agentic-identity-broker", application)
}

// --- Environment Parameterization Tests ---

func TestEnvironmentParameterizationTest(t *testing.T) {
	_, template := createTestStack(t, "test", "")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "AgenticIdentityBrokerBranchKeys-test",
	})
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "AgenticIdentityBrokerEncryptionRole-test",
	})
}

func TestEnvironmentParameterizationProd(t *testing.T) {
	_, template := createTestStack(t, "prod", "system:serviceaccount:default:test-sa")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "AgenticIdentityBrokerBranchKeys-prod",
	})
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "AgenticIdentityBrokerEncryptionRole-prod",
	})
}

func TestEnvironmentParameterizationProductionNormalized(t *testing.T) {
	subject := "system:serviceaccount:default:test-sa"
	originalProps := newTestStackProps("production", subject)
	app := awscdk.NewApp(nil)
	stack := NewEncryptionStack(app, "TestStack", originalProps)
	template := assertions.Template_FromStack(stack, nil)

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "AgenticIdentityBrokerBranchKeys-prod",
	})
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "AgenticIdentityBrokerEncryptionRole-prod",
	})
	template.HasResourceProperties(jsii.String("AWS::KMS::Alias"), map[string]interface{}{
		"AliasName": "alias/agentic-identity-broker/prod/token-vault-kek",
	})
	assert.Equal(t, "production", originalProps.Environment, "NewEncryptionStack must not mutate caller's Environment field")
}

func TestEnvironmentParameterizationSandbox(t *testing.T) {
	_, template := createTestStack(t, "sandbox", "system:serviceaccount:agentic-identity-broker-sandbox:agentic-identity-broker")

	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "AgenticIdentityBrokerBranchKeys-sandbox",
	})
	template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
		"RoleName": "AgenticIdentityBrokerEncryptionRole-sandbox",
	})
	template.HasResourceProperties(jsii.String("AWS::KMS::Alias"), map[string]interface{}{
		"AliasName": "alias/agentic-identity-broker/sandbox/token-vault-kek",
	})
}

func TestSandboxHasNonProdRetentionSettings(t *testing.T) {
	_, template := createTestStack(t, "sandbox", "system:serviceaccount:agentic-identity-broker-sandbox:agentic-identity-broker")

	templateJSON := template.ToJSON()

	// KMS key should use Delete policy (not Retain).
	kmsResources := findResourcesByType(t, templateJSON, "AWS::KMS::Key")
	require.NotEmpty(t, kmsResources)
	for _, res := range kmsResources {
		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Delete", resMap["DeletionPolicy"], "sandbox KMS key should have Delete deletion policy")
	}

	// DynamoDB should have no deletion protection and no PITR.
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"DeletionProtectionEnabled": false,
		"PointInTimeRecoverySpecification": map[string]interface{}{
			"PointInTimeRecoveryEnabled": false,
		},
	})
}

func TestSandboxTrustPolicy(t *testing.T) {
	subject := "system:serviceaccount:agentic-identity-broker-sandbox:agentic-identity-broker"

	_, template := createTestStack(t, "sandbox", subject)

	templateJSON := template.ToJSON()
	roles := findResourcesByType(t, templateJSON, "AWS::IAM::Role")
	require.NotEmpty(t, roles)

	for _, role := range roles {
		stmt := extractTrustStatement(t, role)
		condition, ok := stmt["Condition"].(map[string]interface{})
		require.True(t, ok)
		stringEquals, ok := condition["StringEquals"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, subject, stringEquals[testOIDCSubjectKey])
	}
}

// --- Helpers ---
func newTestStackProps(env, subject string) *EncryptionStackProps {
	if subject == "" {
		subject = defaultServiceAccountSubject
	}

	return &EncryptionStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String("123456789012"),
				Region:  jsii.String("eu-central-1"),
			},
		},
		Environment:           env,
		ServiceAccountSubject: subject,
		OIDCProviderArn:       testOIDCProviderArn,
		OIDCSubjectKey:        testOIDCSubjectKey,
	}
}

// extractTrustStatement returns Statement[0] from the AssumeRolePolicyDocument of a role resource map.
func extractTrustStatement(t *testing.T, role interface{}) map[string]interface{} {
	t.Helper()
	roleMap, ok := role.(map[string]interface{})
	require.True(t, ok)
	props, ok := roleMap["Properties"].(map[string]interface{})
	require.True(t, ok)
	trustDoc, ok := props["AssumeRolePolicyDocument"].(map[string]interface{})
	require.True(t, ok)
	stmts, ok := trustDoc["Statement"].([]interface{})
	require.True(t, ok)
	require.Len(t, stmts, 1)
	stmt, ok := stmts[0].(map[string]interface{})
	require.True(t, ok)
	return stmt
}

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
