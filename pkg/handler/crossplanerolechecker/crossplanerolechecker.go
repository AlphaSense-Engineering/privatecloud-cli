// Package crossplanerolechecker is the package that contains the check functions for Crossplane role.
package crossplanerolechecker

import (
	"errors"
	"fmt"
	"strings"
)

// LogMsgCrossplaneRoleChecked is the message that is logged when the Crossplane role is checked.
const LogMsgCrossplaneRoleChecked = "checked Crossplane role"

// ErrFailedToCheckCrossplaneRole is the error that occurs when the Crossplane role is not checked.
var ErrFailedToCheckCrossplaneRole = errors.New("failed to check Crossplane role")

// RoleMissingPermissionsError is the error that is returned when the role is missing permissions.
type RoleMissingPermissionsError struct {
	// missingPermissions is the list of missing permissions.
	missingPermissions []string
}

var _ error = &RoleMissingPermissionsError{}

// Error is a function that returns the error message.
func (e *RoleMissingPermissionsError) Error() string {
	return fmt.Sprintf("role missing permissions: %s", strings.Join(e.missingPermissions, ", "))
}

// NewRoleMissingPermissionsError is a function that returns a new role missing permissions error.
func NewRoleMissingPermissionsError(missingPermissions []string) error {
	return &RoleMissingPermissionsError{missingPermissions: missingPermissions}
}
