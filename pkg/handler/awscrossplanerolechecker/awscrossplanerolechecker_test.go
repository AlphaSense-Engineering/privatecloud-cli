// Package awscrossplanerolechecker is the package that contains the check functions for AWS Crossplane role.
package awscrossplanerolechecker

import (
	"testing"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/envconfig"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

// setupAWSCrossplaneRoleCheckerTest is a function that sets up a awsCrossplaneRoleChecker for testing.
func setupAWSCrossplaneRoleCheckerTest() *AWSCrossplaneRoleChecker {
	return &AWSCrossplaneRoleChecker{
		envConfig: &envconfig.EnvConfig{
			Spec: envconfig.Spec{
				ClusterName: "test",
				CloudSpec: envconfig.CloudSpec{
					AWS: &envconfig.AWSSpec{
						AccountID: "1234567890",
						OIDCURL:   "oidc.eks.us-west-2.amazonaws.com/id/1234567890",
					},
				},
			},
		},
	}
}

// Test_validatePolicyDocument tests the validatePolicyDocument function.
//
// nolint:funlen
func Test_validatePolicyDocument(t *testing.T) {
	testCases := []struct {
		name             string
		document         rolePolicyDocument
		expectedDocument rolePolicyDocument
		expected         bool
	}{
		{
			name: "Valid Assume Role Policy Document",
			document: rolePolicyDocument{
				Version: aws.String("2012-10-17"),
				Statement: []*rolePolicyStatement{
					{
						Effect: aws.String("Allow"),
						Principal: &rolePolicyPrincipal{
							Federated: aws.String("arn:aws:iam::1234567890:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/1234567890"),
						},
						Action: &[]*string{
							aws.String("sts:AssumeRoleWithWebIdentity"),
						},
						Condition: &rolePolicyCondition{
							StringLike: &map[string]*string{
								"oidc.eks.us-west-2.amazonaws.com/id/1234567890:sub": aws.String("system:serviceaccount:crossplane:aws-*"),
							},
						},
					},
				},
			},
			expectedDocument: constExpectedAssumeRolePolicyDocument,
			expected:         true,
		},
		{
			name: "Invalid Assume Role Policy Document",
			document: rolePolicyDocument{
				Version: aws.String("2012-10-17"),
				Statement: []*rolePolicyStatement{
					{
						Effect: aws.String("Allow"),
						Principal: &rolePolicyPrincipal{
							Federated: aws.String("arn:aws:iam::0987654321:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/0987654321"),
						},
						Action: &[]*string{
							aws.String("sts:AssumeRoleWithWebIdentity"),
						},
						Condition: &rolePolicyCondition{
							StringLike: &map[string]*string{
								"oidc.eks.us-east-1.amazonaws.com/id/0987654321:sub": aws.String("system:serviceaccount:crossplane:aws-*"),
							},
						},
					},
				},
			},
			expectedDocument: constExpectedAssumeRolePolicyDocument,
			expected:         false,
		},
		{
			name: "Valid Boundary Policy Document",
			document: rolePolicyDocument{
				Version: aws.String("2012-10-17"),
				Statement: []*rolePolicyStatement{
					{
						Effect: aws.String("Allow"),
						NotAction: &[]*string{
							aws.String("support:*"),
							aws.String("organizations:*"),
							aws.String("iam:Upload*"),
							aws.String("iam:Update*"),
							aws.String("iam:Untag*"),
							aws.String("iam:Tag*"),
							aws.String("iam:Set*"),
							aws.String("iam:Resync*"),
							aws.String("iam:Reset*"),
							aws.String("iam:Remove*"),
							aws.String("iam:Put*"),
							aws.String("iam:PassRole"),
							aws.String("iam:ListVirtualMFA*"),
							aws.String("iam:ListMFA*"),
							aws.String("iam:GetOrganizationsAccessReport"),
							aws.String("iam:GetAccountAuthorizationDetails"),
							aws.String("iam:Generate*"),
							aws.String("iam:Enable*"),
							aws.String("iam:Detach*"),
							aws.String("iam:Delete*"),
							aws.String("iam:Deactivate*"),
							aws.String("iam:Create*"),
							aws.String("iam:Change*"),
							aws.String("iam:Attach*"),
							aws.String("iam:Add*"),
							aws.String("cloudtrail:DeleteTrail"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowAllActionsApartFromListed"),
					},
				},
			},
			expectedDocument: constExpectedPolicyDocuments["boundary"],
			expected:         true,
		},
		{
			name: "Valid Boundary Policy Document with Extra Actions",
			document: rolePolicyDocument{
				Version: aws.String("2012-10-17"),
				Statement: []*rolePolicyStatement{
					{
						Effect: aws.String("Allow"),
						NotAction: &[]*string{
							aws.String("support:*"),
							aws.String("organizations:*"),
							aws.String("iam:Upload*"),
							aws.String("iam:Update*"),
							aws.String("iam:Untag*"),
							aws.String("iam:Tag*"),
							aws.String("iam:Set*"),
							aws.String("iam:Resync*"),
							aws.String("iam:Reset*"),
							aws.String("iam:Remove*"),
							aws.String("iam:Put*"),
							aws.String("iam:PassRole"),
							aws.String("iam:ListVirtualMFA*"),
							aws.String("iam:ListMFA*"),
							aws.String("iam:GetOrganizationsAccessReport"),
							aws.String("iam:GetAccountAuthorizationDetails"),
							aws.String("iam:Generate*"),
							aws.String("iam:Enable*"),
							aws.String("iam:Detach*"),
							aws.String("iam:Delete*"),
							aws.String("iam:Deactivate*"),
							aws.String("iam:Create*"),
							aws.String("iam:Change*"),
							aws.String("iam:Attach*"),
							aws.String("iam:Add*"),
							aws.String("cloudtrail:DeleteTrail"),
							aws.String("cloudtrail:Get*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowAllActionsApartFromListed"),
					},
				},
			},
			expectedDocument: constExpectedPolicyDocuments["boundary"],
			expected:         true,
		},
		{
			name: "Valid Policy Document",
			document: rolePolicyDocument{
				Version: aws.String("2012-10-17"),
				Statement: []*rolePolicyStatement{
					{
						Effect: aws.String("Deny"),
						Action: &[]*string{
							aws.String("iam:Update*"),
							aws.String("iam:Put*"),
							aws.String("iam:DetachRolePolicy"),
							aws.String("iam:DeleteRolePolicy"),
							aws.String("iam:AttachRolePolicy"),
						},
						Resource: aws.String("arn:aws:iam::1234567890:role/web-identity/test/crossplane-provider-test"),
						SID:      aws.String("DenyAlteringOwnRole"),
					},
					{
						Effect: aws.String("Deny"),
						Action: &[]*string{
							aws.String("iam:SetDefaultPolicyVersion"),
							aws.String("iam:DeletePolicyVersion"),
							aws.String("iam:DeletePolicy"),
							aws.String("iam:CreatePolicyVersion"),
						},
						Resource: aws.String("arn:aws:iam::1234567890:policy/web-identity/test/crossplane-provider-test-boundary"),
						SID:      aws.String("DenyAlteringPermissionsBoundary"),
					},
					{
						Effect: aws.String("Deny"),
						Action: &[]*string{
							aws.String("iam:DeleteRolePermissionsBoundary"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("DenyDeletingAnyPermissionsBoundary"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("sts:GetCallerIdentity"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowCallSTSToGetCurrentIdentity"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("iam:PutRolePolicy"),
							aws.String("iam:PutRolePermissionsBoundary"),
							aws.String("iam:DetachRolePolicy"),
							aws.String("iam:DeleteRolePolicy"),
							aws.String("iam:CreateRole"),
							aws.String("iam:AttachRolePolicy"),
						},
						Resource: aws.String("arn:aws:iam::1234567890:role/web-identity/test/crossplane/*"),
						SID:      aws.String("EnforcePermissionBoundaryOnSpecificIAMActions"),
						Condition: &rolePolicyCondition{
							StringEquals: &map[string]*string{
								"iam:PermissionsBoundary": aws.String("arn:aws:iam::1234567890:policy/web-identity/test/crossplane-provider-test-boundary"),
							},
						},
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("iam:List*"),
							aws.String("iam:GetRole*"),
							aws.String("iam:GetPolicy*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowReadnListAllIAMRolesAndPolicies"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("iam:UpdateRoleDescription"),
							aws.String("iam:UpdateRole"),
							aws.String("iam:UpdateAssumeRolePolicy"),
							aws.String("iam:UntagRole"),
							aws.String("iam:TagRole"),
							aws.String("iam:ListAttachedRolePolicies"),
							aws.String("iam:DeleteRole"),
						},
						Resource: aws.String("arn:aws:iam::1234567890:role/web-identity/test/crossplane/*"),
						SID:      aws.String("AllowCertainIAMActionsWithManagedRoles"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("iam:UntagPolicy"),
							aws.String("iam:TagPolicy"),
							aws.String("iam:DeletePolicy*"),
							aws.String("iam:CreatePolicy*"),
						},
						Resource: aws.String("arn:aws:iam::1234567890:policy/web-identity/test/crossplane/*"),
						SID:      aws.String("AllowCertainIAMActionsWithManagedPolicies"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("s3:ReplicateDelete"),
							aws.String("s3:PutStorageLensConfiguration"),
							aws.String("s3:PutReplicationConfiguration"),
							aws.String("s3:PutLifecycleConfiguration"),
							aws.String("s3:PutIntelligentTieringConfiguration"),
							aws.String("s3:PutEncryptionConfiguration"),
							aws.String("s3:PutBucket*"),
							aws.String("s3:PutAccelerateConfiguration"),
							aws.String("s3:List*"),
							aws.String("s3:Get*"),
							aws.String("s3:DeleteStorageLensConfiguration"),
							aws.String("s3:CreateBucket"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowS3BucketCreation"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("dynamodb:UpdateTimeToLive"),
							aws.String("dynamodb:UpdateTable"),
							aws.String("dynamodb:UpdateGlobalTableSettings"),
							aws.String("dynamodb:UpdateGlobalTable"),
							aws.String("dynamodb:UpdateContinuousBackups"),
							aws.String("dynamodb:UntagResource"),
							aws.String("dynamodb:TagResource"),
							aws.String("dynamodb:ListTagsOfResource"),
							aws.String("dynamodb:ListTables"),
							aws.String("dynamodb:ListStreams"),
							aws.String("dynamodb:ListImports"),
							aws.String("dynamodb:ListGlobalTables"),
							aws.String("dynamodb:ListExports"),
							aws.String("dynamodb:ListContributorInsights"),
							aws.String("dynamodb:ListBackups"),
							aws.String("dynamodb:DescribeTimeToLive"),
							aws.String("dynamodb:DescribeTable"),
							aws.String("dynamodb:DescribeContinuousBackups"),
							aws.String("dynamodb:DeleteTable"),
							aws.String("dynamodb:CreateTable"),
							aws.String("dynamodb:CreateGlobalTable"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowDynamoDB"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("sns:UntagResource"),
							aws.String("sns:Unsubscribe"),
							aws.String("sns:TagResource"),
							aws.String("sns:Subscribe"),
							aws.String("sns:SetTopicAttributes"),
							aws.String("sns:SetSubscriptionAttributes"),
							aws.String("sns:SetEndpointAttributes"),
							aws.String("sns:ListTopics"),
							aws.String("sns:ListTagsForResource"),
							aws.String("sns:ListSubscriptionsByTopic"),
							aws.String("sns:ListSubscriptions"),
							aws.String("sns:ListSMSSandboxPhoneNumbers"),
							aws.String("sns:ListPlatformApplications"),
							aws.String("sns:ListOriginationNumbers"),
							aws.String("sns:ListEndpointsByPlatformApplication"),
							aws.String("sns:GetTopicAttributes"),
							aws.String("sns:GetSubscriptionAttributes"),
							aws.String("sns:GetEndpointAttributes"),
							aws.String("sns:DeleteTopic"),
							aws.String("sns:DeleteEndpoint"),
							aws.String("sns:CreateTopic"),
							aws.String("sns:ConfirmSubscription"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowSNS"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("sqs:UntagQueue"),
							aws.String("sqs:TagQueue"),
							aws.String("sqs:SetQueueAttributes"),
							aws.String("sqs:ReceiveMessage"),
							aws.String("sqs:ListQueues"),
							aws.String("sqs:ListQueueTags"),
							aws.String("sqs:ListDeadLetterSourceQueues"),
							aws.String("sqs:GetQueueUrl"),
							aws.String("sqs:GetQueueAttributes"),
							aws.String("sqs:DeleteQueue"),
							aws.String("sqs:CreateQueue"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowSQS"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("rds:Create*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowTagRestrictedDBCreate"),
						Condition: &rolePolicyCondition{
							StringEquals: &map[string]*string{
								"aws:RequestTag/crossplane-managed": aws.String("true"),
							},
						},
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("rds:Describe*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowAllDBRead"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("rds:RemoveTagsFromResource"),
							aws.String("rds:ListTagsForResource"),
							aws.String("rds:AddTagsToResource"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowTags"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("rds:Modify*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowCrossplaneToModify"),
						Condition: &rolePolicyCondition{
							StringEquals: &map[string]*string{
								"aws:ResourceTag/crossplane-managed": aws.String("true"),
							},
						},
					},
				},
			},
			expectedDocument: constExpectedPolicyDocuments["policy"],
			expected:         true,
		},
		{
			name: "Valid Policy Document with Extra Actions and Conditions",
			document: rolePolicyDocument{
				Version: aws.String("2012-10-17"),
				Statement: []*rolePolicyStatement{
					{
						Effect: aws.String("Deny"),
						Action: &[]*string{
							aws.String("iam:Update*"),
							aws.String("iam:Put*"),
							aws.String("iam:DetachRolePolicy"),
							aws.String("iam:DeleteRolePolicy"),
							aws.String("iam:AttachRolePolicy"),
							aws.String("iam:CreateRole"),
						},
						Resource: aws.String("arn:aws:iam::1234567890:role/web-identity/test/crossplane-provider-test"),
						SID:      aws.String("DenyAlteringOwnRole"),
					},
					{
						Effect: aws.String("Deny"),
						Action: &[]*string{
							aws.String("iam:SetDefaultPolicyVersion"),
							aws.String("iam:DeletePolicyVersion"),
							aws.String("iam:DeletePolicy"),
							aws.String("iam:CreatePolicyVersion"),
							aws.String("iam:CreatePolicy"),
						},
						Resource: aws.String("arn:aws:iam::1234567890:policy/web-identity/test/crossplane-provider-test-boundary"),
						SID:      aws.String("DenyAlteringPermissionsBoundary"),
					},
					{
						Effect: aws.String("Deny"),
						Action: &[]*string{
							aws.String("iam:DeleteRolePermissionsBoundary"),
							aws.String("iam:DeletePolicy"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("DenyDeletingAnyPermissionsBoundary"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("sts:GetCallerIdentity"),
							aws.String("iam:Get*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowCallSTSToGetCurrentIdentity"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("iam:PutRolePolicy"),
							aws.String("iam:PutRolePermissionsBoundary"),
							aws.String("iam:DetachRolePolicy"),
							aws.String("iam:DeleteRolePolicy"),
							aws.String("iam:CreateRole"),
							aws.String("iam:AttachRolePolicy"),
							aws.String("iam:CreatePolicy"),
						},
						Resource: aws.String("arn:aws:iam::1234567890:role/web-identity/test/crossplane/*"),
						SID:      aws.String("EnforcePermissionBoundaryOnSpecificIAMActions"),
						Condition: &rolePolicyCondition{
							StringEquals: &map[string]*string{
								"iam:PermissionsBoundary": aws.String(
									"arn:aws:iam::1234567890:policy/web-identity/test/crossplane-provider-test-boundary",
								),
								"iam:ResourceTag/crossplane-managed": aws.String("true"),
							},
						},
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("iam:List*"),
							aws.String("iam:GetRole*"),
							aws.String("iam:GetPolicy*"),
							aws.String("iam:GetPolicyVersion"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowReadnListAllIAMRolesAndPolicies"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("iam:UpdateRoleDescription"),
							aws.String("iam:UpdateRole"),
							aws.String("iam:UpdateAssumeRolePolicy"),
							aws.String("iam:UntagRole"),
							aws.String("iam:TagRole"),
							aws.String("iam:ListAttachedRolePolicies"),
							aws.String("iam:DeleteRole"),
							aws.String("iam:UpdateRolePermissionsBoundary"),
						},
						Resource: aws.String("arn:aws:iam::1234567890:role/web-identity/test/crossplane/*"),
						SID:      aws.String("AllowCertainIAMActionsWithManagedRoles"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("iam:UntagPolicy"),
							aws.String("iam:TagPolicy"),
							aws.String("iam:DeletePolicy*"),
							aws.String("iam:CreatePolicy*"),
							aws.String("iam:GetPolicyVersion"),
						},
						Resource: aws.String("arn:aws:iam::1234567890:policy/web-identity/test/crossplane/*"),
						SID:      aws.String("AllowCertainIAMActionsWithManagedPolicies"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("s3:ReplicateDelete"),
							aws.String("s3:PutStorageLensConfiguration"),
							aws.String("s3:PutReplicationConfiguration"),
							aws.String("s3:PutLifecycleConfiguration"),
							aws.String("s3:PutIntelligentTieringConfiguration"),
							aws.String("s3:PutEncryptionConfiguration"),
							aws.String("s3:PutBucket*"),
							aws.String("s3:PutAccelerateConfiguration"),
							aws.String("s3:List*"),
							aws.String("s3:Get*"),
							aws.String("s3:DeleteStorageLensConfiguration"),
							aws.String("s3:CreateBucket"),
							aws.String("s3:GetBucketPolicy"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowS3BucketCreation"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("dynamodb:UpdateTimeToLive"),
							aws.String("dynamodb:UpdateTable"),
							aws.String("dynamodb:UpdateGlobalTableSettings"),
							aws.String("dynamodb:UpdateGlobalTable"),
							aws.String("dynamodb:UpdateContinuousBackups"),
							aws.String("dynamodb:UntagResource"),
							aws.String("dynamodb:TagResource"),
							aws.String("dynamodb:ListTagsOfResource"),
							aws.String("dynamodb:ListTables"),
							aws.String("dynamodb:ListStreams"),
							aws.String("dynamodb:ListImports"),
							aws.String("dynamodb:ListGlobalTables"),
							aws.String("dynamodb:ListExports"),
							aws.String("dynamodb:ListContributorInsights"),
							aws.String("dynamodb:ListBackups"),
							aws.String("dynamodb:DescribeTimeToLive"),
							aws.String("dynamodb:DescribeTable"),
							aws.String("dynamodb:DescribeContinuousBackups"),
							aws.String("dynamodb:DeleteTable"),
							aws.String("dynamodb:CreateTable"),
							aws.String("dynamodb:CreateGlobalTable"),
							aws.String("dynamodb:Get*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowDynamoDB"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("sns:UntagResource"),
							aws.String("sns:Unsubscribe"),
							aws.String("sns:TagResource"),
							aws.String("sns:Subscribe"),
							aws.String("sns:SetTopicAttributes"),
							aws.String("sns:SetSubscriptionAttributes"),
							aws.String("sns:SetEndpointAttributes"),
							aws.String("sns:ListTopics"),
							aws.String("sns:ListTagsForResource"),
							aws.String("sns:ListSubscriptionsByTopic"),
							aws.String("sns:ListSubscriptions"),
							aws.String("sns:ListSMSSandboxPhoneNumbers"),
							aws.String("sns:ListPlatformApplications"),
							aws.String("sns:ListOriginationNumbers"),
							aws.String("sns:ListEndpointsByPlatformApplication"),
							aws.String("sns:GetTopicAttributes"),
							aws.String("sns:GetSubscriptionAttributes"),
							aws.String("sns:GetEndpointAttributes"),
							aws.String("sns:DeleteTopic"),
							aws.String("sns:DeleteEndpoint"),
							aws.String("sns:CreateTopic"),
							aws.String("sns:ConfirmSubscription"),
							aws.String("sns:Get*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowSNS"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("sqs:UntagQueue"),
							aws.String("sqs:TagQueue"),
							aws.String("sqs:SetQueueAttributes"),
							aws.String("sqs:ReceiveMessage"),
							aws.String("sqs:ListQueues"),
							aws.String("sqs:ListQueueTags"),
							aws.String("sqs:ListDeadLetterSourceQueues"),
							aws.String("sqs:GetQueueUrl"),
							aws.String("sqs:GetQueueAttributes"),
							aws.String("sqs:DeleteQueue"),
							aws.String("sqs:CreateQueue"),
							aws.String("sqs:Get*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowSQS"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("rds:Create*"),
							aws.String("rds:Get*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowTagRestrictedDBCreate"),
						Condition: &rolePolicyCondition{
							StringEquals: &map[string]*string{
								"aws:RequestTag/crossplane-managed": aws.String("true"),
								"rds:DatabaseClass":                 aws.String("db.t3.small"),
							},
						},
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("rds:Describe*"),
							aws.String("rds:Get*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowAllDBRead"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("rds:RemoveTagsFromResource"),
							aws.String("rds:ListTagsForResource"),
							aws.String("rds:AddTagsToResource"),
							aws.String("rds:Get*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowTags"),
					},
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("rds:Modify*"),
							aws.String("rds:Get*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowCrossplaneToModify"),
						Condition: &rolePolicyCondition{
							StringEquals: &map[string]*string{
								"aws:ResourceTag/crossplane-managed": aws.String("true"),
								"rds:DatabaseClass":                  aws.String("db.t3.small"),
							},
						},
					},
				},
			},
			expectedDocument: constExpectedPolicyDocuments["policy"],
			expected:         true,
		},
		{
			name: "Invalid Policy Document",
			document: rolePolicyDocument{
				Version: aws.String("2012-10-17"),
				Statement: []*rolePolicyStatement{
					{
						Effect: aws.String("Allow"),
						Action: &[]*string{
							aws.String("iam:Update*"),
						},
						Resource: aws.String("*"),
						SID:      aws.String("AllowAllActionsApartFromListed"),
					},
				},
			},
			expectedDocument: constExpectedPolicyDocuments["boundary"],
			expected:         false,
		},
		{
			name: "Valid Redis Policy Document",
			document: rolePolicyDocument{
				Version: aws.String("2012-10-17"),
				Statement: []*rolePolicyStatement{
					{
						Action: &[]*string{
							aws.String("ec2:DescribeSecurityGroups"),
							aws.String("ec2:DescribeSecurityGroupRules"),
							aws.String("ec2:ModifySecurityGroupRules"),
							aws.String("ec2:CreateSecurityGroup"),
							aws.String("ec2:DeleteSecurityGroup"),
							aws.String("ec2:AuthorizeSecurityGroupIngress"),
							aws.String("ec2:AuthorizeSecurityGroupEgress"),
							aws.String("ec2:RevokeSecurityGroupIngress"),
							aws.String("ec2:RevokeSecurityGroupEgress"),
							aws.String("ec2:CreateTags"),
							aws.String("ec2:DeleteTags"),
							aws.String("ec2:DescribeTags"),
							aws.String("ec2:DescribeVpcs"),
							aws.String("ec2:DescribeSubnets"),
							aws.String("ec2:DescribeAvailabilityZones"),
							aws.String("ec2:DescribeNetworkInterfaces"),
							aws.String("ec2:DescribeRouteTables"),
							aws.String("ec2:DescribeVpcEndpoints"),
						},
						Effect:   aws.String("Allow"),
						Resource: aws.String("*"),
						SID:      aws.String("AllowEC2ForRedisInfrastructure"),
					},
					{
						Action: &[]*string{
							aws.String("elasticache:CreateUser"),
							aws.String("elasticache:DescribeReservedCacheNodes"),
							aws.String("elasticache:DescribeReservedCacheNodesOfferings"),
							aws.String("elasticache:DescribeEvents"),
							aws.String("elasticache:IncreaseReplicaCount"),
							aws.String("elasticache:DescribeCacheParameterGroups"),
							aws.String("elasticache:DecreaseReplicaCount"),
							aws.String("elasticache:DescribeEngineDefaultParameters"),
							aws.String("elasticache:CreateGlobalReplicationGroup"),
							aws.String("elasticache:ModifyReplicationGroup"),
							aws.String("elasticache:CreateCacheCluster"),
							aws.String("elasticache:DeleteCacheSubnetGroup"),
							aws.String("elasticache:DescribeServiceUpdates"),
							aws.String("elasticache:DescribeReplicationGroups"),
							aws.String("elasticache:ModifyUserGroup"),
							aws.String("elasticache:DeleteUser"),
							aws.String("elasticache:RemoveTagsFromResource"),
							aws.String("elasticache:DeleteUserGroup"),
							aws.String("elasticache:DeleteCacheCluster"),
							aws.String("elasticache:AddTagsToResource"),
							aws.String("elasticache:ModifyCacheParameterGroup"),
							aws.String("elasticache:DescribeGlobalReplicationGroups"),
							aws.String("elasticache:DescribeUsers"),
							aws.String("elasticache:DescribeCacheClusters"),
							aws.String("elasticache:ListTagsForResource"),
							aws.String("elasticache:CreateReplicationGroup"),
							aws.String("elasticache:AuthorizeCacheSecurityGroupIngress"),
							aws.String("elasticache:DeleteCacheSecurityGroup"),
							aws.String("elasticache:DescribeCacheEngineVersions"),
							aws.String("elasticache:DescribeCacheSubnetGroups"),
							aws.String("elasticache:CreateCacheSubnetGroup"),
							aws.String("elasticache:DescribeSnapshots"),
							aws.String("elasticache:CreateCacheParameterGroup"),
							aws.String("elasticache:DeleteCacheParameterGroup"),
							aws.String("elasticache:DescribeUserGroups"),
							aws.String("elasticache:DisassociateGlobalReplicationGroup"),
							aws.String("elasticache:CreateCacheSecurityGroup"),
							aws.String("elasticache:DescribeCacheParameters"),
							aws.String("elasticache:CreateUserGroup"),
							aws.String("elasticache:DescribeUpdateActions"),
							aws.String("elasticache:ModifyUser"),
							aws.String("elasticache:DeleteGlobalReplicationGroup"),
							aws.String("elasticache:ResetCacheParameterGroup"),
							aws.String("elasticache:DeleteReplicationGroup"),
							aws.String("elasticache:ListAllowedNodeTypeModifications"),
							aws.String("elasticache:ModifyCacheCluster"),
							aws.String("elasticache:ModifyGlobalReplicationGroup"),
							aws.String("elasticache:DescribeCacheSecurityGroups"),
							aws.String("elasticache:ModifyReplicationGroupShardConfiguration"),
							aws.String("elasticache:ModifyCacheSubnetGroup"),
						},
						Effect:   aws.String("Allow"),
						Resource: aws.String("*"),
						SID:      aws.String("AllowElasticache"),
					},
					{
						Action: &[]*string{
							aws.String("iam:CreateServiceLinkedRole"),
						},
						Effect:   aws.String("Allow"),
						Resource: aws.String("*"),
						SID:      aws.String("AllowIAMForServiceLinkedRoles"),
					},
				},
			},
			expectedDocument: constExpectedPolicyDocuments["redis"],
			expected:         true,
		},
		{
			name:             "Empty Policy Document",
			document:         rolePolicyDocument{},
			expectedDocument: constExpectedPolicyDocuments["boundary"],
			expected:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := setupAWSCrossplaneRoleCheckerTest()

			result := c.validatePolicyDocument(tc.document, tc.expectedDocument)

			resultBool := len(result) == 0

			assert.Equal(t, tc.expected, resultBool, "expected %v, got %v (%#v)", tc.expected, resultBool, result)
		})
	}
}
