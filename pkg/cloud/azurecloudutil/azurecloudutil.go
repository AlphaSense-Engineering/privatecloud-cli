// Package azurecloudutil is the package that contains the Azure cloud utility functions.
package azurecloudutil

import (
	"fmt"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
)

// CrossplaneRoleName is a function that returns the name of the Crossplane role.
func CrossplaneRoleName(clusterName string) string {
	return fmt.Sprintf("%s-%s", clusterName, cloud.CrossplaneRoleNameSuffix)
}
