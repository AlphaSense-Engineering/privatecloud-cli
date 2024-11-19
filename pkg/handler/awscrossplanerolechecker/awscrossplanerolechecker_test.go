// Package awscrossplanerolechecker is the package that contains the check functions for AWS Crossplane role.
package awscrossplanerolechecker

import (
	"testing"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
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

			assert.Equal(t, tc.expected, result, "expected %v, got %v", tc.expected, result)
		})
	}
}
