// Package util is the package that contains the utility functions.
package util

import (
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/spf13/cobra"
)

// Flag returns the value of the flag as a string, or an empty string if the flag is not set.
func Flag(cmd *cobra.Command, name string) string {
	flag := cmd.Flag(name)
	if flag == nil {
		return constant.EmptyString
	}

	value := flag.Value
	if value == nil {
		return constant.EmptyString
	}

	return value.String()
}
