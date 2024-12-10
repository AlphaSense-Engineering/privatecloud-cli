// Package main is the package that contains the entry point for the application.
package main

import (
	"os"

	"github.com/AlphaSense-Engineering/privatecloud-installer/cmd"
	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/constant"
	"github.com/charmbracelet/log"
	_ "github.com/go-sql-driver/mysql"
)

// main is the entry point for the application.
func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFunction:    constant.LogDefaultTimeFunc,
	})

	rootCmd := cmd.Root()

	rootCmd.AddCommand(cmd.Check(logger))

	rootCmd.AddCommand(cmd.Pod(logger))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
