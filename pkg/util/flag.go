// Package util is the package that contains the utility functions.
package util

import (
	"strconv"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// flagVal returns the value of the flag as a string, or an empty string if the flag is not set.
func flagVal(flag *pflag.Flag) string {
	if flag == nil || flag.Value == nil {
		return constant.EmptyString
	}

	return flag.Value.String()
}

// Flag returns the value of the flag as a string, or an empty string if the flag is not set.
func Flag(cmd *cobra.Command, name string) string {
	return flagVal(cmd.Flag(name))
}

// PersistentFlag returns the value of the persistent flag as a string, or an empty string if the flag is not set.
func PersistentFlag(cmd *cobra.Command, name string) string {
	return flagVal(cmd.PersistentFlags().Lookup(name))
}

// FlagBool returns the value of the flag as a bool or the default value if the flag is not a boolean.
func FlagBool(cmd *cobra.Command, name string) bool {
	flag := cmd.Flag(name)

	val := flagVal(flag)

	boolValue, err := strconv.ParseBool(val)
	if err != nil {
		return DiscardErr(strconv.ParseBool(flag.DefValue))
	}

	return boolValue
}

// FlagInt returns the value of the flag as an int or the default value if the flag is not an integer.
func FlagInt(cmd *cobra.Command, name string) int {
	flag := cmd.Flag(name)

	val := flagVal(flag)

	intValue, err := strconv.Atoi(val)
	if err != nil {
		return DiscardErr(strconv.Atoi(flag.DefValue))
	}

	return intValue
}
