// Package azurecrossplanerolechecker is the package that contains the check functions for Azure Crossplane role.
package azurecrossplanerolechecker

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud/azurecloudutil"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/envconfig"
	selferrors "github.com/AlphaSense-Engineering/privatecloud-installer/pkg/errors"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/util"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
)

var (
	// errRoleIDNotFound is the error that the role ID is not found.
	errRoleIDNotFound = errors.New("role ID not found")

	// errDuplicatePermission is the error that the permission is duplicated.
	errDuplicatePermission = errors.New("duplicate permission")
)

// constExpectedRolePermissions are the expected permissions for the Crossplane role in Azure.
//
// These are listed at https://developer.alpha-sense.com/enterprise/technical-requirements/azure.
//
// Do not modify this variable, it is supposed to be constant.
var constExpectedRolePermissions = map[string]struct{}{
	"Microsoft.Authorization/policies/audit/action":                                        {},
	"Microsoft.Authorization/policies/auditIfNotExists/action":                             {},
	"Microsoft.Authorization/roleAssignments/delete":                                       {},
	"Microsoft.Authorization/roleAssignments/read":                                         {},
	"Microsoft.Authorization/roleAssignments/write":                                        {},
	"Microsoft.Authorization/roleDefinitions/delete":                                       {},
	"Microsoft.Authorization/roleDefinitions/read":                                         {},
	"Microsoft.Authorization/roleDefinitions/write":                                        {},
	"Microsoft.ManagedIdentity/userAssignedIdentities/delete":                              {},
	"Microsoft.ManagedIdentity/userAssignedIdentities/federatedIdentityCredentials/delete": {},
	"Microsoft.ManagedIdentity/userAssignedIdentities/federatedIdentityCredentials/read":   {},
	"Microsoft.ManagedIdentity/userAssignedIdentities/federatedIdentityCredentials/write":  {},
	"Microsoft.ManagedIdentity/userAssignedIdentities/read":                                {},
	"Microsoft.ManagedIdentity/userAssignedIdentities/write":                               {},
	"Microsoft.Network/virtualNetworks/read":                                               {},
	"Microsoft.Network/virtualNetworks/subnets/join/action":                                {},
	"Microsoft.Network/virtualNetworks/subnets/joinViaServiceEndpoint/action":              {},
	"Microsoft.Network/virtualNetworks/subnets/read":                                       {},
	"Microsoft.ServiceBus/namespaces/Delete":                                               {},
	"Microsoft.ServiceBus/namespaces/queues/Delete":                                        {},
	"Microsoft.ServiceBus/namespaces/queues/read":                                          {},
	"Microsoft.ServiceBus/namespaces/queues/write":                                         {},
	"Microsoft.ServiceBus/namespaces/read":                                                 {},
	"Microsoft.ServiceBus/namespaces/topics/Delete":                                        {},
	"Microsoft.ServiceBus/namespaces/topics/read":                                          {},
	"Microsoft.ServiceBus/namespaces/topics/subscriptions/Delete":                          {},
	"Microsoft.ServiceBus/namespaces/topics/subscriptions/read":                            {},
	"Microsoft.ServiceBus/namespaces/topics/subscriptions/rules/Delete":                    {},
	"Microsoft.ServiceBus/namespaces/topics/subscriptions/rules/read":                      {},
	"Microsoft.ServiceBus/namespaces/topics/subscriptions/rules/write":                     {},
	"Microsoft.ServiceBus/namespaces/topics/subscriptions/write":                           {},
	"Microsoft.ServiceBus/namespaces/topics/write":                                         {},
	"Microsoft.ServiceBus/namespaces/write":                                                {},
	"Microsoft.Storage/skus/read":                                                          {},
	"Microsoft.Storage/storageAccounts/blobServices/containers/delete":                     {},
	"Microsoft.Storage/storageAccounts/blobServices/containers/read":                       {},
	"Microsoft.Storage/storageAccounts/blobServices/containers/write":                      {},
	"Microsoft.Storage/storageAccounts/blobServices/generateUserDelegationKey/action":      {},
	"Microsoft.Storage/storageAccounts/blobServices/read":                                  {},
	"Microsoft.Storage/storageAccounts/blobServices/write":                                 {},
	"Microsoft.Storage/storageAccounts/delete":                                             {},
	"Microsoft.Storage/storageAccounts/fileServices/read":                                  {},
	"Microsoft.Storage/storageAccounts/listkeys/action":                                    {},
	"Microsoft.Storage/storageAccounts/managementPolicies/delete":                          {},
	"Microsoft.Storage/storageAccounts/managementPolicies/read":                            {},
	"Microsoft.Storage/storageAccounts/managementPolicies/write":                           {},
	"Microsoft.Storage/storageAccounts/read":                                               {},
	"Microsoft.Storage/storageAccounts/regeneratekey/action":                               {},
	"Microsoft.Storage/storageAccounts/write":                                              {},
}

// AzureCrossplaneRoleChecker is the type that contains the check functions for Azure Crossplane role.
type AzureCrossplaneRoleChecker struct {
	// envConfig is the environment configuration.
	envConfig *envconfig.EnvConfig
	// roleDefClient is the Azure role definitions client.
	roleDefClient *armauthorization.RoleDefinitionsClient
}

var _ handler.Handler = &AzureCrossplaneRoleChecker{}

// Handle is the function that handles the Azure Crossplane role check.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
//
// nolint:funlen
func (c *AzureCrossplaneRoleChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	scope := fmt.Sprintf("subscriptions/%s/resourceGroups/%s", c.envConfig.Spec.CloudSpec.Azure.SubscriptionID, c.envConfig.Spec.CloudSpec.Azure.ResourceGroup)

	listPager := c.roleDefClient.NewListPager(scope, nil)

	var roleID *string

	for listPager.More() {
		nextResult, err := listPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range nextResult.Value {
			if *v.Properties.RoleName != azurecloudutil.CrossplaneRoleName(c.envConfig.Spec.ClusterName) {
				continue
			}

			// Extract the role ID in UUID format from the full resource ID.
			roleID = util.Ref((*v.ID)[strings.LastIndex(*v.ID, string(constant.HTTPPathSeparator))+1:])

			break
		}
	}

	if roleID == nil {
		return nil, errRoleIDNotFound
	}

	roleDef, err := c.roleDefClient.Get(ctx, scope, *roleID, nil)
	if err != nil {
		return nil, err
	}

	foundPermissions := make(map[string]struct{})

	missingPermissions := []string{}

	for _, permission := range roleDef.Properties.Permissions {
		if permission.Actions == nil {
			continue
		}

		for _, action := range permission.Actions {
			if _, ok := foundPermissions[*action]; ok {
				return nil, errDuplicatePermission
			}

			foundPermissions[*action] = struct{}{}
		}
	}

	for k := range constExpectedRolePermissions {
		if _, ok := foundPermissions[k]; ok {
			continue
		}

		missingPermissions = append(missingPermissions, k)
	}

	if len(missingPermissions) > 0 {
		return nil, selferrors.NewRoleMissingPermissions(missingPermissions)
	}

	return nil, nil
}

// New is the function that creates a new AzureCrossplaneRoleChecker.
func New(envConfig *envconfig.EnvConfig, roleDefClient *armauthorization.RoleDefinitionsClient) *AzureCrossplaneRoleChecker {
	return &AzureCrossplaneRoleChecker{
		envConfig:     envConfig,
		roleDefClient: roleDefClient,
	}
}
