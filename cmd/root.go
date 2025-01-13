// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"
	"github.com/spf13/cobra"
)

const (
	// FlagVerbose is the flag to enable verbose output.
	FlagVerbose = "verbose"

	// flagVerboseShort is the short flag to enable verbose output.
	flagVerboseShort = "v"
)

// rootCmd is the root command for the application.
type rootCmd struct{}

var _ cmd = &rootCmd{}

// run is the run function for the root command.
func (c *rootCmd) run(cmd *cobra.Command, _ []string) {
	_ = cmd.Help()
}

// newRootCmd returns a new rootCmd.
func newRootCmd() *rootCmd {
	return &rootCmd{}
}

// Root returns a Cobra command that is the root command of the application.
func Root() *cobra.Command {
	cmd := newRootCmd()

	cobraCmd := &cobra.Command{
		Use: constant.AppName,
		Run: cmd.run,
	}

	cobraCmd.PersistentFlags().BoolP(FlagVerbose, flagVerboseShort, false, "verbose output")

	return cobraCmd
}
