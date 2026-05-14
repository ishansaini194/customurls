package cli

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

// Color styles. fatih/color auto-disables colors when output is piped
// or the terminal doesn't support them, so JSON/script output stays clean.
var (
	successStyle = color.New(color.FgGreen, color.Bold).SprintFunc()
	errorStyle   = color.New(color.FgRed, color.Bold).SprintFunc()
	labelStyle   = color.New(color.FgHiBlack).SprintFunc()
	valueStyle   = color.New(color.FgCyan, color.Bold).SprintFunc()
)

// formatExpiry turns an hour count (from /shorten) into a human-readable string.
func formatExpiry(hours int) string {
	switch {
	case hours <= 0:
		return "never"
	case hours >= 8760:
		return fmt.Sprintf("%dy", hours/8760)
	case hours >= 720:
		return fmt.Sprintf("%dmo", hours/720)
	case hours >= 24:
		return fmt.Sprintf("%dd", hours/24)
	default:
		return fmt.Sprintf("%dh", hours)
	}
}

// formatExpiryTime turns an expiry timestamp (from /stats) into a readable string.
func formatExpiryTime(t *time.Time) string {
	if t == nil {
		return "never"
	}
	remaining := time.Until(*t)
	if remaining <= 0 {
		return "expired"
	}
	days := int(remaining.Hours()) / 24
	hours := int(remaining.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("in %dd %dh (%s)", days, hours, t.Format("2006-01-02"))
	}
	return fmt.Sprintf("in %dh (%s)", hours, t.Format("2006-01-02 15:04"))
}
