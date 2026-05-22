package cmd

import (
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean identified space hogs with confirmation",
	Long: `Clean selected categories of disk hogs. Always shows a dry-run preview before asking
for confirmation. Defaults to moving items to the OS trash (recoverable).

Examples:
  diskwhy clean                     # Interactive selection from last scan
  diskwhy clean --docker            # Unused Docker images and volumes
  diskwhy clean --node              # node_modules older than DISKWHY_STALE_DAYS
  diskwhy clean --cache             # npm, pip, brew, and yarn caches
  diskwhy clean --git               # Run git gc --prune=now in discovered repos
  diskwhy clean --logs              # Compressed logs older than 7 days
  diskwhy clean --all --dry-run     # Preview everything without changing anything
  diskwhy clean --all --trash       # Move all safe items to OS trash
  diskwhy clean --docker --yes      # Clean Docker without interactive prompt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	cleanCmd.Flags().Bool("docker", false, "Clean unused Docker images and volumes")
	cleanCmd.Flags().Bool("node", false, "Clean node_modules older than DISKWHY_STALE_DAYS")
	cleanCmd.Flags().Bool("cache", false, "Clean npm, pip, brew, and yarn caches")
	cleanCmd.Flags().Bool("git", false, "Run git gc --prune=now in all discovered repositories")
	cleanCmd.Flags().Bool("logs", false, "Clean compressed log files older than 7 days")
	cleanCmd.Flags().Bool("trash", false, "Move selected items to OS trash instead of permanently deleting")
	cleanCmd.Flags().Bool("all", false, "Clean everything safe (requires typing 'yes' to confirm)")
	cleanCmd.Flags().Bool("dry-run", false, "Show what would be cleaned without making any changes")
	cleanCmd.Flags().BoolP("yes", "y", false, "Skip interactive confirmation prompt")
	cleanCmd.Flags().Bool("no-trash", false, "Permanently delete instead of moving to trash")
	cleanCmd.Flags().Bool("json", false, "Output results as JSON (schema_version: 1)")
}
