// Package main is the package that contains the entry point for the application.
package main

import (
	"os"

	"github.com/AlphaSense-Engineering/privatecloud-installer/cmd"
)

// main is the entry point for the application.
func main() {
	rootCmd := cmd.Root()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
