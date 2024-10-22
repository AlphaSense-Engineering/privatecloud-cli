// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"github.com/spf13/cobra"
)

// Root returns the root command for the application.
func Root() *cobra.Command {
	return &cobra.Command{
		Use: "privatecloud-installer",
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}
}
