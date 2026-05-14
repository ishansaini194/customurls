package cli

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var (
	shortenAlias string
	shortenJSON  bool
)

var shortenCmd = &cobra.Command{
	Use:   "shorten <url>",
	Short: "Shorten a URL",
	Long: `Shorten a URL and print the short link.

Use --alias to choose a custom short ID instead of a random one.
Use --json to print the full response as JSON (useful for scripts).

Examples:
  cshort shorten https://example.com/very/long/url
  cshort shorten https://example.com --alias my-link
  cshort shorten https://example.com --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		// Validate URL
		parsed, err := url.ParseRequestURI(input)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return fmt.Errorf("invalid URL: must start with http:// or https://")
		}

		resp, err := shorten(input, shortenAlias)
		if err != nil {
			return err
		}

		if shortenJSON {
			out, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		// Pretty output
		fmt.Println(successStyle("✓ Shortened"))
		fmt.Printf("  %s %s\n", labelStyle("short: "), valueStyle(resp.Short))
		fmt.Printf("  %s %s\n", labelStyle("target:"), resp.URL)
		fmt.Printf("  %s %s\n", labelStyle("expiry:"), formatExpiry(resp.Expiry))
		return nil
	},
}

func init() {
	shortenCmd.Flags().StringVarP(&shortenAlias, "alias", "a", "", "custom alias for the short URL")
	shortenCmd.Flags().BoolVar(&shortenJSON, "json", false, "output as JSON")
}
