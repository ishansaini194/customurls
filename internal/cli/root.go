package cli

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var alias string

var rootCmd = &cobra.Command{
	Use:   "cshort <url>",
	Short: "Shorten URLs from your terminal",
	Long: `cshort is a command-line client for customurls.in.

It shortens any URL and prints the short link.
Use --alias to choose a custom short ID.

Examples:
  cshort https://example.com/very/long/url
  cshort https://example.com --alias my-link`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		// Validate URL
		parsed, err := url.ParseRequestURI(input)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return fmt.Errorf("invalid URL: must start with http:// or https://")
		}

		short, err := shorten(input, alias)
		if err != nil {
			return err
		}

		fmt.Println(short)
		return nil
	},
	SilenceUsage:  true, // don't print help on every error
	SilenceErrors: true, // we print our own
}

// Execute runs the root command. Called by cmd/cshort/main.go.
func Execute() {
	rootCmd.Flags().StringVarP(&alias, "alias", "a", "", "custom alias for the short URL")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
