package cli

import (
	"encoding/json"
	"fmt"
	"net/url"

	qrcode "github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var (
	shortenAlias string
	shortenJSON  bool
	shortenQR    bool
)

var shortenCmd = &cobra.Command{
	Use:   "short <url>",
	Short: "Shorten a URL",
	Long: `Shorten a URL and print the short link.

Use --alias to choose a custom short ID instead of a random one.
Use --qr to also print a scannable QR code in the terminal.
Use --json to print the full response as JSON (useful for scripts).

Examples:
  custom short https://example.com/very/long/url
  custom short https://example.com --alias my-link
  custom short https://example.com --qr
  custom short https://example.com --json`,
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

		if shortenQR {
			qr, err := qrcode.New(resp.Short, qrcode.Medium)
			if err != nil {
				return fmt.Errorf("failed to generate QR: %w", err)
			}
			fmt.Println()
			// ToSmallString packs 2 rows per line — fits in a normal terminal.
			fmt.Print(qr.ToSmallString(false))
		}
		return nil
	},
}

func init() {
	shortenCmd.Flags().StringVarP(&shortenAlias, "alias", "a", "", "custom alias for the short URL")
	shortenCmd.Flags().BoolVar(&shortenJSON, "json", false, "output as JSON")
	shortenCmd.Flags().BoolVar(&shortenQR, "qr", false, "also print a QR code in the terminal")
}
