// Package errors is the package that contains the error types.
//
// nolint:errname
package errors

import (
	"fmt"
	"strings"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
)

// KeyExpectedGot is the error that is returned when the key is expected to be a certain value, but it is not.
type KeyExpectedGot struct {
	// key is the key that is mismatched.
	key string
	// expected is the expected value.
	expected string
	// got is the got value.
	got string
}

var _ error = &KeyExpectedGot{}

// Error is a function that returns the error message.
func (e *KeyExpectedGot) Error() string {
	return fmt.Sprintf("expected %s to be %s, got %s", e.key, e.expected, e.got)
}

// NewKeyExpectedGot is a function that returns a new KeyExpectedGot error.
func NewKeyExpectedGot(key, expected, got string) *KeyExpectedGot {
	return &KeyExpectedGot{key: key, expected: expected, got: got}
}

// RoleMissingPermissions is the error that is returned when the role is missing permissions.
type RoleMissingPermissions struct {
	// missingPermissions is the list of missing permissions.
	missingPermissions []string
}

var _ error = &RoleMissingPermissions{}

// Error is a function that returns the error message.
func (e *RoleMissingPermissions) Error() string {
	return fmt.Sprintf("role missing permissions: %s", strings.Join(e.missingPermissions, ", "))
}

// NewRoleMissingPermissions is a function that returns a new RoleMissingPermissions error.
func NewRoleMissingPermissions(missingPermissions []string) error {
	return &RoleMissingPermissions{missingPermissions: missingPermissions}
}

// UnsupportedCloud is the error that is returned when the cloud is unsupported.
type UnsupportedCloud struct {
	// cloud is the cloud that is unsupported.
	cloud cloud.Cloud
}

var _ error = &UnsupportedCloud{}

// Error is a function that returns the error message.
func (e *UnsupportedCloud) Error() string {
	return fmt.Sprintf("unsupported cloud type: %s", e.cloud)
}

// NewUnsupportedCloud is a function that returns a new UnsupportedCloud error.
func NewUnsupportedCloud(cloud cloud.Cloud) error {
	return &UnsupportedCloud{cloud: cloud}
}
