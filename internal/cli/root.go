package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "custom",
	Short: "Shorten URLs from your terminal",
	Long: `custom is a command-line client for customurls.in.

Shorten URLs, use custom aliases, and check click stats —
all without leaving your terminal.

Examples:
  custom short https://example.com/very/long/url
  custom short https://example.com --alias my-link
  custom stats my-link`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command. Called by cmd/custom/main.go.
func Execute() {
	rootCmd.AddCommand(shortenCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, errorStyle("error:"), err)
		os.Exit(1)
	}
}
