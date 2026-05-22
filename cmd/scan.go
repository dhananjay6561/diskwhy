package cmd

import (
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan disk usage and identify space hogs",
	Long: `Scan the disk and categorize usage into developer-meaningful buckets.

Output shows proportional bars, last-used signals, and estimated safe-to-clean totals.

Examples:
  diskwhy                       # Quick scan of top culprits (default)
  diskwhy scan                  # Same as above
  diskwhy scan --deep           # Full recursive scan (~10-15s on large disks)
  diskwhy scan --path /var      # Scan a specific directory
  diskwhy scan --json           # Structured JSON output (schema_version: 1)
  diskwhy scan --deep --json    # Full scan with JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	scanCmd.Flags().Bool("deep", false, "Full recursive scan (~10-15 seconds on large disks)")
	scanCmd.Flags().StringP("path", "p", "", "Scan a specific path instead of the full disk")
	scanCmd.Flags().Bool("json", false, "Output results as JSON (schema_version: 1)")
}
