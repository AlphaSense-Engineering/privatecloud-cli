// Package main is the package that contains the entry point for the application.
package main

import (
	"os"

	"github.com/AlphaSense-Engineering/privatecloud-cli/cmd"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/constant"
	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/util"
	"github.com/charmbracelet/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

// addCommand is the function that adds a command to the root command and hooks its Run function.
func addCommand(logger *log.Logger, rootCmd *cobra.Command, cmdFn func(*log.Logger) *cobra.Command) {
	cobraCmd := cmdFn(logger)

	oldRun := cobraCmd.Run

	cobraCmd.Run = func(cobraCmd *cobra.Command, args []string) {
		if util.FlagBool(cobraCmd, cmd.FlagVerbose) {
			logger.SetLevel(log.DebugLevel)
		}

		oldRun(cobraCmd, args)
	}

	rootCmd.AddCommand(cobraCmd)
}

// main is the entry point for the application.
func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFunction:    constant.LogDefaultTimeFunc,
	})

	rootCmd := cmd.Root()

	cmdFns := []func(*log.Logger) *cobra.Command{
		cmd.Check,
		cmd.Install,
		cmd.Pod,
	}

	for _, cmdFn := range cmdFns {
		addCommand(logger, rootCmd, cmdFn)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
