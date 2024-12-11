// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/spf13/cobra"
)

// rootCmd is the root command for the application.
type rootCmd struct{}

var _ cmd = &rootCmd{}

// Run is the run function for the root command.
func (c *rootCmd) Run(cmd *cobra.Command, _ []string) {
	_ = cmd.Help()
}

// newRootCmd returns a new rootCmd.
func newRootCmd() *rootCmd {
	return &rootCmd{}
}

// Root returns a Cobra command that is the root command of the application.
func Root() *cobra.Command {
	cmd := newRootCmd()

	return &cobra.Command{
		Use: constant.AppName,
		Run: cmd.Run,
	}
}
