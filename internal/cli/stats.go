package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

var statsJSON bool

var statsCmd = &cobra.Command{
	Use:   "stats <short-id-or-url>",
	Short: "Show click stats for a short URL",
	Long: `Show stats for a short URL: total clicks, original URL, and expiry.

You can pass any of these and custom will figure out the short ID:
  - the bare short ID:        8pFVcr
  - the alias + ID path:      s26-blue/8pFVcr
  - the full short URL:       https://customurls.in/s26-blue/8pFVcr

Examples:
  custom stats 8pFVcr
  custom stats s26-blue/8pFVcr
  custom stats https://customurls.in/s26-blue/8pFVcr --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shortID := extractShortID(args[0])
		if shortID == "" {
			return fmt.Errorf("could not extract a short ID from input")
		}

		resp, err := stats(shortID)
		if err != nil {
			return err
		}

		if statsJSON {
			out, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		// Pretty output
		fmt.Println(successStyle("✓ Stats for ") + valueStyle(resp.ShortID))
		fmt.Printf("  %s %s\n", labelStyle("clicks:"), valueStyle(fmt.Sprintf("%d", resp.Hits)))
		fmt.Printf("  %s %s\n", labelStyle("target:"), resp.OriginalURL)
		fmt.Printf("  %s %s\n", labelStyle("expiry:"), formatExpiryTime(resp.ExpiresAt))
		return nil
	},
}

// extractShortID pulls the short ID out of whatever the user pasted:
// a bare ID, an "alias/id" path, or a full URL. The short ID is always
// the LAST path segment.
func extractShortID(input string) string {
	s := input

	// If it's a full URL, take just the path.
	if parsed, err := url.ParseRequestURI(input); err == nil && parsed.Host != "" {
		s = parsed.Path
	}

	// Trim slashes and take the last segment.
	s = strings.Trim(s, "/")
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "/")
	return parts[len(parts)-1]
}

func init() {
	statsCmd.Flags().BoolVar(&statsJSON, "json", false, "output as JSON")
}
