// Package errors is the package that contains the error types.
//
// nolint:errname
package errors

import (
	"fmt"
	"strings"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/cloud"
	"github.com/r3labs/diff/v3"
)

// ErrWithChangelog is the error that is returned when there is an error and a changelog.
type ErrWithChangelog struct {
	// err is the error.
	err error
	// changelog is the changelog.
	changelog diff.Changelog
}

var _ error = &ErrWithChangelog{}

// Error is a function that returns the error message.
func (e *ErrWithChangelog) Error() string {
	return fmt.Errorf("%w: %#v", e.err, e.changelog).Error()
}

// NewErrWithChangelog is a function that returns a new ErrWithChangelog error.
func NewErrWithChangelog(err error, changelog diff.Changelog) error {
	return &ErrWithChangelog{err: err, changelog: changelog}
}

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

// KeysEmpty is the error that is returned when the keys are empty.
type KeysEmpty[K comparable] struct {
	// keys is the list of keys that are empty.
	keys []K
}

var _ error = &KeysEmpty[any]{}

// Error is a function that returns the error message.
func (e *KeysEmpty[K]) Error() string {
	strKeys := make([]string, len(e.keys))

	for i, key := range e.keys {
		strKeys[i] = fmt.Sprintf("%v", key)
	}

	return fmt.Sprintf("keys empty: %s", strings.Join(strKeys, ", "))
}

// NewKeysEmpty is a function that returns a new KeysEmpty error.
func NewKeysEmpty[K comparable](keys []K) error {
	return &KeysEmpty[K]{keys: keys}
}

// KeysMissing is the error that is returned when the keys are missing.
type KeysMissing[K comparable] struct {
	// keys is the list of keys that are missing.
	keys []K
}

var _ error = &KeysMissing[any]{}

// Error is a function that returns the error message.
func (e *KeysMissing[K]) Error() string {
	strKeys := make([]string, len(e.keys))

	for i, key := range e.keys {
		strKeys[i] = fmt.Sprintf("%v", key)
	}

	return fmt.Sprintf("keys missing: %s", strings.Join(strKeys, ", "))
}

// NewKeysMissing is a function that returns a new KeysMissing error.
func NewKeysMissing[K comparable](keys []K) error {
	return &KeysMissing[K]{keys: keys}
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
