// Package cmd is the package that contains all of the commands for the application.
package cmd

import (
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/spf13/cobra"
)

// Root returns the root command for the application.
func Root() *cobra.Command {
	return &cobra.Command{
		Use: constant.AppName,
		Run: root,
	}
}

// root is the run function for the Root command.
func root(cmd *cobra.Command, _ []string) {
	_ = cmd.Help()
}
