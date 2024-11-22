// Package constant is the package that contains the constant variables.
package constant

import "log"

// ErrAssertionFailed is the error message that occurs when the assertion fails.
const ErrAssertionFailed = "assertion failed"

const (
	// AppName is the name of the application.
	AppName = "privatecloud-installer"

	// LogFlags is the flags for the log.
	LogFlags = log.LstdFlags | log.LUTC

	// EmptyString is the empty string.
	EmptyString = ""

	// HTTPPathSeparator is the path separator for the HTTP URL.
	HTTPPathSeparator = '/'
)

const (
	// NamespaceAlphaSense is the namespace for the AlphaSense.
	NamespaceAlphaSense = "alphasense"

	// NamespaceCrossplane is the namespace for the Crossplane.
	NamespaceCrossplane = "crossplane"

	// NamespaceMySQL is the namespace for the MySQL.
	NamespaceMySQL = "mysql"

	// NamespacePlatform is the namespace for the platform.
	NamespacePlatform = "platform"
)
