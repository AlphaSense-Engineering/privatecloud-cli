// Package util is the package that contains the utility functions.
package util

import (
	"strconv"

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

// FlagBool returns the value of the flag as a bool, or false if the flag is not set.
func FlagBool(cmd *cobra.Command, name string) bool {
	flag := cmd.Flag(name)
	if flag == nil {
		return false
	}

	value := flag.Value
	if value == nil {
		return false
	}

	boolValue, err := strconv.ParseBool(value.String())
	if err != nil {
		return false
	}

	return boolValue
}
