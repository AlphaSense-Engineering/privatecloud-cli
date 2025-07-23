// Package awscrossplanerolechecker is the package that contains the check functions for AWS Crossplane role.
package awscrossplanerolechecker

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/cloud/awscloudutil"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/envconfig"
	pkgerrors "github.com/AlphaSense-Engineering/privatecloud-cli/pkg/errors"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/charmbracelet/log"
	"github.com/r3labs/diff/v3"
)

var (
	// errNoAssumeRolePolicyDocument is an error that occurs when the healthcheck fails to find the assume role policy document.
	errNoAssumeRolePolicyDocument = errors.New("no assume role policy document")

	// errAssumeRolePolicyDocumentMismatch is an error that occurs when the assume role policy document does not match the expected document.
	errAssumeRolePolicyDocumentMismatch = errors.New("assume role policy document mismatch")

	// errNoDefaultPolicyVersion is an error that occurs when the healthcheck fails to find the default policy version.
	errNoDefaultPolicyVersion = errors.New("no default policy version")

	// errPolicyVersionOrDocumentNil is an error that occurs when the policy version or document is nil.
	errPolicyVersionOrDocumentNil = errors.New("policy version or document is nil")

	// errPolicyDocumentMismatch is an error that occurs when the policy document does not match the expected document.
	errPolicyDocumentMismatch = errors.New("policy document does not match")
)

// rolePolicyCondition is the struct for the AWS role policy condition.
type rolePolicyCondition struct {
	// StringEquals is the string equals of the AWS role policy condition.
	StringEquals *map[string]*string `json:"StringEquals,omitempty"`
	// StringLike is the string like of the AWS role policy condition.
	StringLike *map[string]*string `json:"StringLike,omitempty"`
}

// rolePolicyPrincipal is the struct for the AWS role policy principal.
type rolePolicyPrincipal struct {
	// Federated is the federated of the AWS role policy principal.
	Federated *string `json:"Federated,omitempty"`
}

// rolePolicyStatement is the struct for the AWS role policy statement.
type rolePolicyStatement struct {
	// Condition is the condition of the AWS role policy statement.
	Condition *rolePolicyCondition `json:"Condition,omitempty"`
	// Effect is the effect of the AWS role policy statement.
	Effect *string `json:"Effect,omitempty"`
	// Action is the action of the AWS role policy statement.
	Action *[]*string `json:"-"`
	// NotAction is the not action of the AWS role policy statement.
	NotAction *[]*string `json:"-"`
	// Principal is the principal of the AWS role policy statement.
	Principal *rolePolicyPrincipal `json:"Principal,omitempty"`
	// Resource is the resource of the AWS role policy statement.
	Resource *string `json:"Resource,omitempty"`
	// SID is the SID of the AWS role policy statement.
	SID *string `json:"Sid,omitempty"`
}

// MarshalJSON is a custom JSON marshaller for awsRolePolicyStatement.
func (s *rolePolicyStatement) MarshalJSON() ([]byte, error) {
	type Alias rolePolicyStatement

	aux := &struct {
		Action    any `json:"Action,omitempty"`
		NotAction any `json:"NotAction,omitempty"`

		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if s.Action != nil {
		if len(*s.Action) == 1 {
			aux.Action = (*s.Action)[0]
		} else {
			aux.Action = s.Action
		}
	}

	if s.NotAction != nil {
		if len(*s.NotAction) == 1 {
			aux.NotAction = (*s.NotAction)[0]
		} else {
			aux.NotAction = s.NotAction
		}
	}

	return json.Marshal(aux)
}

// UnmarshalJSON is a custom JSON unmarshaller for awsRolePolicyStatement.
func (s *rolePolicyStatement) UnmarshalJSON(data []byte) error {
	type Alias rolePolicyStatement

	aux := &struct {
		Action    any `json:"Action,omitempty"`
		NotAction any `json:"NotAction,omitempty"`

		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.Action.(type) {
	case string:
		s.Action = &[]*string{aws.String(v)}
	case []any:
		var actions []*string

		for _, a := range v {
			if str, ok := a.(string); ok {
				actions = append(actions, aws.String(str))
			}
		}

		s.Action = &actions
	}

	switch v := aux.NotAction.(type) {
	case string:
		s.NotAction = &[]*string{aws.String(v)}
	case []any:
		var notActions []*string

		for _, a := range v {
			if str, ok := a.(string); ok {
				notActions = append(notActions, aws.String(str))
			}
		}

		s.NotAction = &notActions
	}

	return nil
}

// rolePolicyDocument is the struct for the AWS role policy document.
type rolePolicyDocument struct {
	// Version is the version of the AWS role policy document.
	Version *string `json:"Version,omitempty"`
	// Statement is the statement of the AWS role policy document.
	Statement []*rolePolicyStatement `json:"Statement,omitempty"`
}

var (
	// constExpectedAssumeRolePolicyDocument is the expected AWS assume role policy document.
	//
	// This is listed at https://developer.alpha-sense.com/enterprise/technical-requirements/aws.
	//
	// Do not modify this variable, it is supposed to be constant.
	constExpectedAssumeRolePolicyDocument = rolePolicyDocument{
		Version: aws.String("2012-10-17"),
		Statement: []*rolePolicyStatement{
			{
				Effect: aws.String("Allow"),
				Principal: &rolePolicyPrincipal{
					Federated: aws.String("arn:aws:iam::${ACCOUNT_ID}:oidc-provider/${OIDC_ID}"),
				},
				Action: &[]*string{
					aws.String("sts:AssumeRoleWithWebIdentity"),
				},
				Condition: &rolePolicyCondition{
					StringLike: &map[string]*string{
						"${OIDC_ID}:sub": aws.String("system:serviceaccount:crossplane:aws-*"),
					},
				},
			},
		},
	}

	// constAWSPoliciesNameSuffixes is the map of suffixes and the expected policy document for the AWS policies.
	//
	// These are listed at https://developer.alpha-sense.com/enterprise/technical-requirements/aws.
	//
	// Do not modify this variable, it is supposed to be constant.
	constExpectedPolicyDocuments = map[string]rolePolicyDocument{
		"boundary": {
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

		"policy": {
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
					Resource: aws.String("arn:aws:iam::${ACCOUNT_ID}:role/web-identity/${CLUSTER_NAME}/crossplane-provider-${CLUSTER_NAME}"),
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
					Resource: aws.String("arn:aws:iam::${ACCOUNT_ID}:policy/web-identity/${CLUSTER_NAME}/crossplane-provider-${CLUSTER_NAME}-boundary"),
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
					Resource: aws.String("arn:aws:iam::${ACCOUNT_ID}:role/web-identity/${CLUSTER_NAME}/crossplane/*"),
					SID:      aws.String("EnforcePermissionBoundaryOnSpecificIAMActions"),
					Condition: &rolePolicyCondition{
						StringEquals: &map[string]*string{
							"iam:PermissionsBoundary": aws.String(
								"arn:aws:iam::${ACCOUNT_ID}:policy/web-identity/${CLUSTER_NAME}/crossplane-provider-${CLUSTER_NAME}-boundary",
							),
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
					Resource: aws.String("arn:aws:iam::${ACCOUNT_ID}:role/web-identity/${CLUSTER_NAME}/crossplane/*"),
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
					Resource: aws.String("arn:aws:iam::${ACCOUNT_ID}:policy/web-identity/${CLUSTER_NAME}/crossplane/*"),
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

		"redis": {
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
	}
)

// AWSCrossplaneRoleChecker is the type that contains the check functions for AWS Crossplane role.
type AWSCrossplaneRoleChecker struct {
	// logger is the logger.
	logger *log.Logger
	// envConfig is the environment configuration.
	envConfig *envconfig.EnvConfig
	// iam is the AWS IAM client.
	iam *iam.Client
}

var _ handler.Handler = &AWSCrossplaneRoleChecker{}

// fillPlaceholdersString is a function that fills the placeholders in the string.
func (c *AWSCrossplaneRoleChecker) fillPlaceholdersString(s string) string {
	const (
		// clusterNamePlaceholder is the placeholder for the cluster name.
		clusterNamePlaceholder = "${CLUSTER_NAME}"

		// accountIDPlaceholder is the placeholder for the account ID.
		accountIDPlaceholder = "${ACCOUNT_ID}"

		// oidcURLPlaceholder is the placeholder for the OIDC URL.
		oidcURLPlaceholder = "${OIDC_ID}"
	)

	s = strings.ReplaceAll(s, clusterNamePlaceholder, c.envConfig.Spec.ClusterName)

	s = strings.ReplaceAll(s, accountIDPlaceholder, c.envConfig.Spec.CloudSpec.AWS.AccountID)

	s = strings.ReplaceAll(s, oidcURLPlaceholder, c.envConfig.Spec.CloudSpec.AWS.OIDCURL)

	return s
}

// fillPlaceholdersMap is a function that fills the placeholders in the map.
func (c *AWSCrossplaneRoleChecker) fillPlaceholdersMap(m *map[string]*string) *map[string]*string {
	newMap := make(map[string]*string, len(*m))

	for key, value := range *m {
		newMap[c.fillPlaceholdersString(key)] = aws.String(c.fillPlaceholdersString(util.Deref(value)))
	}

	return &newMap
}

// validatePolicyDocument is a function that validates the AWS policy document.
//
// nolint:gocognit
func (c *AWSCrossplaneRoleChecker) validatePolicyDocument(document rolePolicyDocument, expectedDocument rolePolicyDocument) diff.Changelog {
	for _, stmt := range expectedDocument.Statement {
		if stmt.Principal != nil && stmt.Principal.Federated != nil {
			*stmt.Principal.Federated = c.fillPlaceholdersString(util.Deref(stmt.Principal.Federated))
		}

		if stmt.Resource != nil {
			*stmt.Resource = c.fillPlaceholdersString(util.Deref(stmt.Resource))
		}

		if stmt.Condition != nil {
			if stmt.Condition.StringEquals != nil {
				stmt.Condition.StringEquals = c.fillPlaceholdersMap(stmt.Condition.StringEquals)
			}

			if stmt.Condition.StringLike != nil {
				stmt.Condition.StringLike = c.fillPlaceholdersMap(stmt.Condition.StringLike)
			}
		}
	}

	changelog, err := diff.Diff(expectedDocument, document)
	if err != nil {
		panic(err)
	}

	const (
		// statementPath is the path to the statement.
		statementPath = "Statement"
		// statementPathIndex is the index of the statement path.
		statementPathIndex = 0
		// actionPath is the path to the action.
		actionPath = "Action"
		// notActionPath is the path to the not action.
		notActionPath = "NotAction"
		// conditionPath is the path to the condition.
		conditionPath = "Condition"
		// actionNotActionPathLength is the desired length of the path for Action/NotAction.
		actionNotActionPathLength = 4
		// conditionPathLength is the desired length of the path for Condition.
		conditionPathLength = 5
		// actionNotActionConditionPathIndex is the index of the action or not action or condition path.
		actionNotActionConditionPathIndex = 2
	)

	// We need to allow extra items in Action/NotAction, and prohibit removing expected ones.
	// This is why we filter out CREATE changelog entries that are in the Action/NotAction path.
	filteredChangelog := changelog[:0]

	for _, change := range changelog {
		if change.Type == diff.CREATE && change.Path[statementPathIndex] == statementPath {
			if (len(change.Path) == actionNotActionPathLength &&
				(change.Path[actionNotActionConditionPathIndex] == actionPath || change.Path[actionNotActionConditionPathIndex] == notActionPath)) ||
				(len(change.Path) == conditionPathLength && change.Path[actionNotActionConditionPathIndex] == conditionPath) {
				continue
			}
		}

		filteredChangelog = append(filteredChangelog, change)
	}

	changelog = filteredChangelog

	return changelog
}

// processPolicyDocument is a function that processes the AWS policy document.
func (c *AWSCrossplaneRoleChecker) processPolicyDocument(ctx context.Context, roleName, suffix string, expectedPolicyDocument rolePolicyDocument) error {
	policyARN := aws.String(awscloudutil.ARN(
		c.envConfig.Spec.CloudSpec.AWS.AccountID,
		c.envConfig.Spec.ClusterName,
		awscloudutil.ARNTypePolicy,
		roleName,
		&suffix,
	))

	policyVersions, err := c.iam.ListPolicyVersions(ctx, &iam.ListPolicyVersionsInput{PolicyArn: policyARN})
	if err != nil {
		return err
	}

	var defaultVersionID *string

	for _, version := range policyVersions.Versions {
		if !version.IsDefaultVersion {
			continue
		}

		defaultVersionID = version.VersionId
	}

	if defaultVersionID == nil {
		return errNoDefaultPolicyVersion
	}

	policyVersion, err := c.iam.GetPolicyVersion(ctx, &iam.GetPolicyVersionInput{PolicyArn: policyARN, VersionId: defaultVersionID})
	if err != nil {
		return err
	}

	if policyVersion.PolicyVersion == nil || policyVersion.PolicyVersion.Document == nil {
		return errPolicyVersionOrDocumentNil
	}

	var policyDocument rolePolicyDocument

	policyDocumentData, err := url.QueryUnescape(*policyVersion.PolicyVersion.Document)
	if err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(policyDocumentData), &policyDocument); err != nil {
		return err
	}

	changelog := c.validatePolicyDocument(policyDocument, expectedPolicyDocument)
	if len(changelog) > 0 {
		return pkgerrors.NewErrWithChangelog(errPolicyDocumentMismatch, changelog)
	}

	return nil
}

// Handle is the function that handles the AWS Crossplane role check.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *AWSCrossplaneRoleChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	roleName := awscloudutil.CrossplaneRoleName(c.envConfig.Spec.ClusterName)

	role, err := c.iam.GetRole(ctx, &iam.GetRoleInput{RoleName: aws.String(roleName)})
	if err != nil {
		return nil, err
	}

	if role.Role.AssumeRolePolicyDocument == nil {
		return nil, errNoAssumeRolePolicyDocument
	}

	var assumeRolePolicyDocument rolePolicyDocument

	assumeRolePolicyDocumentData, err := url.QueryUnescape(*role.Role.AssumeRolePolicyDocument)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(assumeRolePolicyDocumentData), &assumeRolePolicyDocument); err != nil {
		return nil, err
	}

	changelog := c.validatePolicyDocument(assumeRolePolicyDocument, constExpectedAssumeRolePolicyDocument)
	if len(changelog) > 0 {
		return nil, pkgerrors.NewErrWithChangelog(errAssumeRolePolicyDocumentMismatch, changelog)
	}

	for suffix, expectedPolicyDocument := range constExpectedPolicyDocuments {
		if err := c.processPolicyDocument(ctx, roleName, suffix, expectedPolicyDocument); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// New is the function that creates a new AWSCrossplaneRoleChecker.
func New(logger *log.Logger, envConfig *envconfig.EnvConfig, iam *iam.Client) *AWSCrossplaneRoleChecker {
	return &AWSCrossplaneRoleChecker{
		logger:    logger,
		envConfig: envConfig,
		iam:       iam,
	}
}
