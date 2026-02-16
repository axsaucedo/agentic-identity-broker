package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Constants for encryption infrastructure configuration.
const (
	// IAM role session duration (hours).
	MaxSessionDurationHours = 1

	// KMS key pending deletion window for production (days, maximum for safety).
	ProdDeletionWindowDays = 30

	// KMS key pending deletion window for non-production (days, minimum for fast cleanup).
	DevDeletionWindowDays = 7

	// CloudWatch metrics period (minutes).
	MetricsPeriodMinutes = 5

	// CloudWatch alarm thresholds.
	KMSThrottleThreshold   = 10 // throttled requests over 2 evaluation periods
	KMSThrottleEvalPeriods = 2  // evaluation periods for throttle alarm
	KMSErrorThreshold      = 5  // error count over 1 evaluation period
	KMSErrorEvalPeriods    = 1  // evaluation periods for error alarm
)

// EncryptionStackProps extends StackProps with encryption-specific parameters.
type EncryptionStackProps struct {
	awscdk.StackProps

	// Environment is the deployment environment: "dev", "staging", or "prod".
	Environment string

	// TrustPrincipal is the IAM principal ARN allowed to assume the encryption role.
	// Examples:
	//   - "arn:aws:iam::123456789012:role/ECSTaskRole"  (ECS)
	//   - "arn:aws:iam::123456789012:root"              (same-account)
	//   - OIDC provider ARN for EKS IRSA
	//
	// When empty, defaults to same-account root principal (for dev/test).
	TrustPrincipal string
}

// NewEncryptionStack creates the Token Vault encryption infrastructure stack.
//
// Resources provisioned:
//   - AWS KMS symmetric CMK (KEK for the hierarchical keyring)
//   - DynamoDB table (branch key cache for the hierarchical keyring)
//   - IAM role with least-privilege KMS + DynamoDB permissions
//
// Stack outputs are named to map directly to the application's IDENTITY_BROKER_*
// environment variables.
func NewEncryptionStack(scope constructs.Construct, id string, props *EncryptionStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	isProd := props.Environment == "prod" || props.Environment == "production"

	// ─── Production Trust Principal Enforcement ─────────────────────────
	//
	// Production deployments MUST specify an explicit trust principal to prevent
	// the default account-root principal from allowing any IAM user/role in the
	// account to assume the encryption role.
	if isProd && props.TrustPrincipal == "" {
		panic("ERROR: Production deployments require explicit trustPrincipal.\n" +
			"Usage: cdk deploy -c env=prod -c trustPrincipal=arn:aws:iam::ACCOUNT:role/ROLE_NAME\n" +
			"Example: cdk deploy -c env=prod -c trustPrincipal=arn:aws:iam::123456789012:role/ECSTaskRole")
	}

	// ─── Tags ───────────────────────────────────────────────────────────
	//
	// Apply standard tags to all taggable resources in this stack.
	awscdk.Tags_Of(stack).Add(jsii.String("Project"), jsii.String("agentic-identity-broker"), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("Component"), jsii.String("encryption"), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("Environment"), jsii.String(props.Environment), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("ManagedBy"), jsii.String("aws-cdk"), nil)

	// ─── KMS Key (KEK) ──────────────────────────────────────────────────
	//
	// Symmetric CMK used as the Key Encryption Key in the hierarchical keyring.
	// The AWS Encryption SDK wraps per-operation DEKs with branch keys that are
	// themselves protected by this CMK.
	kmsKey := awskms.NewKey(stack, jsii.String("EncryptionKEK"), &awskms.KeyProps{
		Description:       jsii.String(fmt.Sprintf("Agentic Identity Broker - Token Vault KEK (%s)", props.Environment)),
		KeySpec:           awskms.KeySpec_SYMMETRIC_DEFAULT,
		KeyUsage:          awskms.KeyUsage_ENCRYPT_DECRYPT,
		EnableKeyRotation: jsii.Bool(true),
		Alias:             jsii.String(fmt.Sprintf("alias/identity-broker/%s/token-vault-kek", props.Environment)),
		// Prod: RETAIN on stack deletion to prevent accidental data loss.
		// Non-prod: DESTROY for clean teardown.
		RemovalPolicy: removalPolicy(isProd),
		// 30-day pending deletion window for production (max safety).
		PendingWindow: pendingWindow(isProd),
	})

	// ─── DynamoDB Table (Branch Key Cache) ──────────────────────────────
	//
	// The AWS Encryption SDK Hierarchical Keyring requires a DynamoDB table
	// with the specific schema: partition_key (S) + sort_key (S).
	// See: https://docs.aws.amazon.com/encryption-sdk/latest/developer-guide/use-hierarchical-keyring.html
	tableName := fmt.Sprintf("IdentityBrokerBranchKeys-%s", props.Environment)
	branchKeyTable := awsdynamodb.NewTable(stack, jsii.String("BranchKeyTable"), &awsdynamodb.TableProps{
		TableName: jsii.String(tableName),
		// Schema required by the AWS Encryption SDK KeyStore.
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("partition_key"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("sort_key"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode: awsdynamodb.BillingMode_PAY_PER_REQUEST,
		Encryption:  awsdynamodb.TableEncryption_AWS_MANAGED,
		PointInTimeRecoverySpecification: &awsdynamodb.PointInTimeRecoverySpecification{
			PointInTimeRecoveryEnabled: jsii.Bool(isProd),
		},
		DeletionProtection: jsii.Bool(isProd),
		RemovalPolicy:      removalPolicy(isProd),
	})

	// ─── CloudWatch Alarms (KMS Monitoring) ─────────────────────────────
	//
	// Monitor KMS key health and throttling to ensure encryption operations
	// remain performant and catch quota issues early.

	// KMS Throttle Alarm - alerts when KMS API throttling is detected.
	// Throttling indicates approaching or exceeding KMS request quota.
	kmsThrottleMetric := awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Namespace:  jsii.String("AWS/KMS"),
		MetricName: jsii.String("UserErrorCount"),
		DimensionsMap: &map[string]*string{
			"KeyId": kmsKey.KeyId(),
		},
		Statistic: jsii.String("Sum"),
		Period:    awscdk.Duration_Minutes(jsii.Number(MetricsPeriodMinutes)),
	})

	kmsThrottleAlarm := awscloudwatch.NewAlarm(stack, jsii.String("KMSThrottleAlarm"), &awscloudwatch.AlarmProps{
		Metric:             kmsThrottleMetric,
		Threshold:          jsii.Number(KMSThrottleThreshold),
		EvaluationPeriods:  jsii.Number(KMSThrottleEvalPeriods),
		AlarmName:          jsii.String(fmt.Sprintf("IdentityBroker-Encryption-%s-KMS-Throttle", props.Environment)),
		AlarmDescription:   jsii.String("KMS key throttling detected - may indicate insufficient quota"),
		TreatMissingData:   awscloudwatch.TreatMissingData_NOT_BREACHING,
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_THRESHOLD,
	})

	// KMS User Errors Alarm - alerts on KMS API errors (access denied, key disabled, etc.).
	// Helps catch permission issues, key state problems, or misconfiguration.
	kmsErrorMetric := awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Namespace:  jsii.String("AWS/KMS"),
		MetricName: jsii.String("UserErrorCount"),
		DimensionsMap: &map[string]*string{
			"KeyId": kmsKey.KeyId(),
		},
		Statistic: jsii.String("Sum"),
		Period:    awscdk.Duration_Minutes(jsii.Number(MetricsPeriodMinutes)),
	})

	kmsErrorAlarm := awscloudwatch.NewAlarm(stack, jsii.String("KMSErrorAlarm"), &awscloudwatch.AlarmProps{
		Metric:             kmsErrorMetric,
		Threshold:          jsii.Number(KMSErrorThreshold),
		EvaluationPeriods:  jsii.Number(KMSErrorEvalPeriods),
		AlarmName:          jsii.String(fmt.Sprintf("IdentityBroker-Encryption-%s-KMS-Errors", props.Environment)),
		AlarmDescription:   jsii.String("KMS API errors detected - check key policy and permissions"),
		TreatMissingData:   awscloudwatch.TreatMissingData_NOT_BREACHING,
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_THRESHOLD,
	})

	// ─── IAM Role (Encryption Operations) ───────────────────────────────
	//
	// Dedicated role with least-privilege access to KMS and DynamoDB.
	// The trust principal is configurable to support any compute platform.
	trustPrincipal := buildTrustPrincipal(stack, props.TrustPrincipal)

	encryptionRole := awsiam.NewRole(stack, jsii.String("EncryptionRole"), &awsiam.RoleProps{
		RoleName:           jsii.String(fmt.Sprintf("IdentityBrokerEncryptionRole-%s", props.Environment)),
		Description:        jsii.String("IAM role for Agentic Identity Broker encryption operations (KMS + DynamoDB)"),
		AssumedBy:          trustPrincipal,
		MaxSessionDuration: awscdk.Duration_Hours(jsii.Number(MaxSessionDurationHours)),
	})

	// KMS permissions: encrypt, decrypt, generate/re-encrypt data keys.
	// These are the minimum permissions required by the hierarchical keyring.
	encryptionRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("KMSHierarchicalKeyringOperations"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"kms:Encrypt",
			"kms:Decrypt",
			"kms:GenerateDataKey",
			"kms:GenerateDataKeyWithoutPlaintext",
			"kms:ReEncryptFrom",
			"kms:ReEncryptTo",
			"kms:DescribeKey",
		),
		Resources: &[]*string{kmsKey.KeyArn()},
	}))

	// DynamoDB permissions: the KeyStore needs read/write access for branch key CRUD.
	encryptionRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("DynamoDBBranchKeyOperations"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"dynamodb:GetItem",
			"dynamodb:PutItem",
			"dynamodb:Query",
			"dynamodb:UpdateItem",
			"dynamodb:DeleteItem",
			"dynamodb:DescribeTable",
		),
		Resources: &[]*string{branchKeyTable.TableArn()},
	}))

	// KMS:CreateGrant permission — required by the hierarchical keyring
	// to delegate key operations to DynamoDB for branch key wrapping.
	encryptionRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("KMSCreateGrantForKeyring"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"kms:CreateGrant",
		),
		Resources: &[]*string{kmsKey.KeyArn()},
		Conditions: &map[string]interface{}{
			"Bool": map[string]*string{
				"kms:GrantIsForAWSResource": jsii.String("true"),
			},
		},
	}))

	// ─── Stack Outputs ──────────────────────────────────────────────────
	//
	// These map directly to the IDENTITY_BROKER_* environment variables
	// consumed by the application's EncryptionConfig.
	//
	// To inject into the app:
	//   export IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN=$(aws cloudformation describe-stacks \
	//     --stack-name <stack> --query "Stacks[0].Outputs[?OutputKey=='EncryptionKeyARN'].OutputValue" --output text)

	awscdk.NewCfnOutput(stack, jsii.String("EncryptionKeyARN"), &awscdk.CfnOutputProps{
		Value:       kmsKey.KeyArn(),
		Description: jsii.String("KMS CMK ARN → IDENTITY_BROKER_ENCRYPTION_AWS_KMS_KEY_ARN"),
		ExportName:  jsii.String(fmt.Sprintf("IdentityBroker-%s-EncryptionKeyARN", props.Environment)),
	})

	awscdk.NewCfnOutput(stack, jsii.String("EncryptionKeyAlias"), &awscdk.CfnOutputProps{
		Value:       jsii.String(fmt.Sprintf("alias/identity-broker/%s/token-vault-kek", props.Environment)),
		Description: jsii.String("KMS CMK Alias for human-readable reference"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("BranchKeyTableName"), &awscdk.CfnOutputProps{
		Value:       branchKeyTable.TableName(),
		Description: jsii.String("DynamoDB table name → IDENTITY_BROKER_ENCRYPTION_AWS_KMS_DYNAMODB_TABLE_NAME"),
		ExportName:  jsii.String(fmt.Sprintf("IdentityBroker-%s-BranchKeyTableName", props.Environment)),
	})

	awscdk.NewCfnOutput(stack, jsii.String("BranchKeyTableARN"), &awscdk.CfnOutputProps{
		Value:       branchKeyTable.TableArn(),
		Description: jsii.String("DynamoDB table ARN (for IAM policy reference)"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("EncryptionRoleARN"), &awscdk.CfnOutputProps{
		Value:       encryptionRole.RoleArn(),
		Description: jsii.String("IAM role ARN → IDENTITY_BROKER_ENCRYPTION_IAM_ROLE_ARN"),
		ExportName:  jsii.String(fmt.Sprintf("IdentityBroker-%s-EncryptionRoleARN", props.Environment)),
	})

	awscdk.NewCfnOutput(stack, jsii.String("KMSThrottleAlarmArn"), &awscdk.CfnOutputProps{
		Value:       kmsThrottleAlarm.AlarmArn(),
		Description: jsii.String("CloudWatch Alarm ARN for KMS throttling (subscribe SNS topic for alerts)"),
		ExportName:  jsii.String(fmt.Sprintf("IdentityBroker-%s-KMSThrottleAlarmArn", props.Environment)),
	})

	awscdk.NewCfnOutput(stack, jsii.String("KMSErrorAlarmArn"), &awscdk.CfnOutputProps{
		Value:       kmsErrorAlarm.AlarmArn(),
		Description: jsii.String("CloudWatch Alarm ARN for KMS errors (subscribe SNS topic for alerts)"),
		ExportName:  jsii.String(fmt.Sprintf("IdentityBroker-%s-KMSErrorAlarmArn", props.Environment)),
	})

	// ─── CloudWatch Dashboard ───────────────────────────────────────────
	//
	// Create operational dashboard for encryption infrastructure monitoring.
	createDashboard(stack, kmsKey, branchKeyTable, props)

	return stack
}

// buildTrustPrincipal constructs the IAM trust principal for the encryption role.
// If trustPrincipalARN is empty, uses the same-account root principal.
func buildTrustPrincipal(stack awscdk.Stack, trustPrincipalARN string) awsiam.IPrincipal {
	if trustPrincipalARN == "" {
		// Default: allow any principal in the same AWS account to assume the role.
		// In production, always specify a concrete trust principal.
		return awsiam.NewAccountRootPrincipal()
	}
	return awsiam.NewArnPrincipal(jsii.String(trustPrincipalARN))
}

// removalPolicy returns RETAIN for production, DESTROY for non-production.
func removalPolicy(isProd bool) awscdk.RemovalPolicy {
	if isProd {
		return awscdk.RemovalPolicy_RETAIN
	}
	return awscdk.RemovalPolicy_DESTROY
}

// pendingWindow returns the KMS key pending deletion window.
// Production: 30 days (maximum, for safety). Non-production: 7 days (minimum, for fast cleanup).
func pendingWindow(isProd bool) awscdk.Duration {
	if isProd {
		return awscdk.Duration_Days(jsii.Number(ProdDeletionWindowDays))
	}
	return awscdk.Duration_Days(jsii.Number(DevDeletionWindowDays))
}

// createDashboard creates a CloudWatch dashboard for encryption infrastructure monitoring.
// Displays KMS API metrics and DynamoDB branch key cache utilization for operational visibility.
func createDashboard(stack awscdk.Stack, kmsKey awskms.IKey, table awsdynamodb.ITable, props *EncryptionStackProps) awscloudwatch.Dashboard {
	dashboard := awscloudwatch.NewDashboard(stack, jsii.String("EncryptionDashboard"), &awscloudwatch.DashboardProps{
		DashboardName: jsii.String(fmt.Sprintf("IdentityBroker-Encryption-%s", props.Environment)),
	})

	// KMS API Calls Metric - tracks encryption/decryption operations.
	kmsApiCallsMetric := awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Namespace:  jsii.String("AWS/KMS"),
		MetricName: jsii.String("ApiCalls"),
		DimensionsMap: &map[string]*string{
			"KeyId": kmsKey.KeyId(),
		},
		Statistic: jsii.String("Sum"),
		Period:    awscdk.Duration_Minutes(jsii.Number(MetricsPeriodMinutes)),
	})

	// KMS Throttled Requests Metric - monitors throttling issues.
	kmsThrottleMetric := awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Namespace:  jsii.String("AWS/KMS"),
		MetricName: jsii.String("UserErrorCount"),
		DimensionsMap: &map[string]*string{
			"KeyId": kmsKey.KeyId(),
		},
		Statistic: jsii.String("Sum"),
		Period:    awscdk.Duration_Minutes(jsii.Number(MetricsPeriodMinutes)),
	})

	// DynamoDB Read Capacity Metric - tracks branch key cache reads.
	dynamoReadMetric := awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Namespace:  jsii.String("AWS/DynamoDB"),
		MetricName: jsii.String("ConsumedReadCapacityUnits"),
		DimensionsMap: &map[string]*string{
			"TableName": table.TableName(),
		},
		Statistic: jsii.String("Sum"),
		Period:    awscdk.Duration_Minutes(jsii.Number(MetricsPeriodMinutes)),
	})

	// DynamoDB Write Capacity Metric - tracks branch key cache writes.
	dynamoWriteMetric := awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Namespace:  jsii.String("AWS/DynamoDB"),
		MetricName: jsii.String("ConsumedWriteCapacityUnits"),
		DimensionsMap: &map[string]*string{
			"TableName": table.TableName(),
		},
		Statistic: jsii.String("Sum"),
		Period:    awscdk.Duration_Minutes(jsii.Number(MetricsPeriodMinutes)),
	})

	// Add widgets to dashboard - two columns layout.
	dashboard.AddWidgets(
		awscloudwatch.NewGraphWidget(&awscloudwatch.GraphWidgetProps{
			Title: jsii.String("KMS API Operations"),
			Left: &[]awscloudwatch.IMetric{
				kmsApiCallsMetric,
				kmsThrottleMetric,
			},
			Width: jsii.Number(12),
		}),
		awscloudwatch.NewGraphWidget(&awscloudwatch.GraphWidgetProps{
			Title: jsii.String("DynamoDB Branch Key Cache"),
			Left: &[]awscloudwatch.IMetric{
				dynamoReadMetric,
				dynamoWriteMetric,
			},
			Width: jsii.Number(12),
		}),
	)

	return dashboard
}
