// Package constant is the package that contains the constant variables.
package constant

// ErrAssertionFailed is the error message that occurs when the assertion fails.
const ErrAssertionFailed = "assertion failed"

const (
	// AppName is the name of the application.
	AppName = "privatecloud-cli"

	// EmptyString is the empty string.
	EmptyString = ""

	// HTTPPathSeparator is the path separator for the HTTP URL.
	HTTPPathSeparator = '/'
)

var (
	// BuildVersion is the current build version. It is supposed to be a Git tag (without the v prefix), or the name of the snapshot, or "dev" if not set by
	// ldflags.
	//
	// Do not modify this variable, it is supposed to be constant.
	// This variable is set by ldflags during the build time.
	BuildVersion = "dev"

	// BuildCommit is the Git commit hash that was used during the build, or "none" if not set by ldflags.
	//
	// Do not modify this variable, it is supposed to be constant.
	// This variable is set by ldflags during the build time.
	BuildCommit = "none"

	// BuildDate is the build date in the RFC 3339 format, or "unknown" if not set by ldflags.
	//
	// Do not modify this variable, it is supposed to be constant.
	// This variable is set by ldflags during the build time.
	BuildDate = "unknown"
)
