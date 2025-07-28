package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set at compile time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

// createVersionCommand creates and returns the version command
func createVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of resize-tool",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("resize-tool version %s\n", version)
			fmt.Printf("commit: %s\n", commit)
			fmt.Printf("date: %s\n", date)
			fmt.Printf("built by: %s\n", builtBy)
		},
	}
}

// getVersion returns the version string
func getVersion() string {
	return version
}
