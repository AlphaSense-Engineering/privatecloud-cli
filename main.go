// Package main is the package that contains the entry point for the application.
package main

import (
	"os"

	"github.com/AlphaSense-Engineering/privatecloud-installer/cmd"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/charmbracelet/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

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
		rootCmd.AddCommand(cmdFn(logger))
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
