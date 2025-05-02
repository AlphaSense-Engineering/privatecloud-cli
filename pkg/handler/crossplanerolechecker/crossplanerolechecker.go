// Package crossplanerolechecker is the package that contains the check functions for Crossplane role.
package crossplanerolechecker

import (
	"errors"
)

// LogMsgCrossplaneRoleCheckedSuccessfully is the message that is logged when the Crossplane role is checked successfully.
const LogMsgCrossplaneRoleCheckedSuccessfully = "checked Crossplane role successfully"

// ErrFailedToCheckCrossplaneRole is the error that occurs when the Crossplane role is not checked.
var ErrFailedToCheckCrossplaneRole = errors.New("failed to check Crossplane role")
