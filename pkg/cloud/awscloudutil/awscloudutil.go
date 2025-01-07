// Package awscloudutil is the package that contains the AWS cloud utility functions.
package awscloudutil

import (
	"fmt"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/cloud"
)

// ARNType is the type of the ARN.
type ARNType string

const (
	// ARNTypeRole is the role type of the ARN.
	ARNTypeRole ARNType = "role"

	// ARNTypePolicy is the policy type of the ARN.
	ARNTypePolicy ARNType = "policy"
)

// ARN is a function that returns the ARN for the desired resource.
func ARN(accountID string, clusterName string, arnType ARNType, name string, suffix *string) string {
	// arnFormat is the format of the ARN to check for permissions.
	const arnFormat = "arn:aws:iam::%s:%s/web-identity/%s/%s"

	arn := fmt.Sprintf(arnFormat, accountID, arnType, clusterName, name)

	if suffix != nil {
		arn = fmt.Sprintf("%s-%s", arn, *suffix)
	}

	return arn
}

// CrossplaneRoleName is a function that returns the name of the Crossplane role.
func CrossplaneRoleName(clusterName string) string {
	return fmt.Sprintf("%s-%s", cloud.CrossplaneRoleNameSuffix, clusterName)
}
