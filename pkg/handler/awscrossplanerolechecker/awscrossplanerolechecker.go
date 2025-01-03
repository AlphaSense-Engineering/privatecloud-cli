// Package awscrossplanerolechecker is the package that contains the check functions for AWS Crossplane role.
package awscrossplanerolechecker

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud/awscloudutil"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	pkgerrors "github.com/AlphaSense-Engineering/privatecloud-installer/pkg/errors"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
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

// fillPlaceholders is a function that fills the placeholders in the string.
func (c *AWSCrossplaneRoleChecker) fillPlaceholders(s string) string {
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

// validatePolicyDocument is a function that validates the AWS policy document.
//
// nolint:gocognit
func (c *AWSCrossplaneRoleChecker) validatePolicyDocument(document rolePolicyDocument, expectedDocument rolePolicyDocument) diff.Changelog {
	for _, stmt := range expectedDocument.Statement {
		if stmt.Principal != nil && stmt.Principal.Federated != nil {
			*stmt.Principal.Federated = c.fillPlaceholders(util.Deref(stmt.Principal.Federated))
		}

		if stmt.Resource != nil {
			*stmt.Resource = c.fillPlaceholders(util.Deref(stmt.Resource))
		}

		if stmt.Condition != nil && stmt.Condition.StringLike != nil {
			newStringLike := make(map[string]*string)

			for key, value := range *stmt.Condition.StringLike {
				newStringLike[c.fillPlaceholders(key)] = aws.String(c.fillPlaceholders(util.Deref(value)))
			}

			stmt.Condition.StringLike = &newStringLike
		}
	}

	changelog, err := diff.Diff(expectedDocument, document)
	if err != nil {
		panic(err)
	}

	const (
		// desiredPathLength is the desired length of the path.
		desiredPathLength = 4
		// statementPath is the path to the statement.
		statementPath = "Statement"
		// statementPathIndex is the index of the statement path.
		statementPathIndex = 0
		// actionPath is the path to the action.
		actionPath = "Action"
		// notActionPath is the path to the not action.
		notActionPath = "NotAction"
		// actionNotActionPathIndex is the index of the action or not action path.
		actionNotActionPathIndex = 2
	)

	// We need to allow extra items in Action/NotAction, and prohibit removing expected ones.
	// This is why we filter out CREATE changelog entries that are in the Action/NotAction path.
	filteredChangelog := changelog[:0]

	for _, change := range changelog {
		if change.Type == diff.CREATE && (len(change.Path) == desiredPathLength &&
			change.Path[statementPathIndex] == statementPath &&
			(change.Path[actionNotActionPathIndex] == actionPath || change.Path[actionNotActionPathIndex] == notActionPath)) {
			continue
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
