package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cshort",
	Short: "Shorten URLs from your terminal",
	Long: `cshort is a command-line client for customurls.in.

Shorten URLs, use custom aliases, and check click stats —
all without leaving your terminal.

Examples:
  cshort shorten https://example.com/very/long/url
  cshort shorten https://example.com --alias my-link
  cshort stats my-link`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command. Called by cmd/cshort/main.go.
func Execute() {
	rootCmd.AddCommand(shortenCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, errorStyle("error:"), err)
		os.Exit(1)
	}
}
