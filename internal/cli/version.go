package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These are set at build time via -ldflags. See README for the build command.
var (
	version = "dev"
	commit  = "none"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the cshort version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cshort %s (commit %s)\n", version, commit)
	},
}
