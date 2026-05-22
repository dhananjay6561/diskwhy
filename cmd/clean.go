package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/dhananjay6561/diskwhy/internal/clean"
	"github.com/dhananjay6561/diskwhy/internal/docker"
	"github.com/dhananjay6561/diskwhy/internal/jsonout"
	"github.com/dhananjay6561/diskwhy/internal/scan"
	"github.com/dhananjay6561/diskwhy/internal/tui"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean identified space hogs with confirmation",
	Long: `Clean selected categories of disk hogs. Always shows a preview before asking
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
	RunE: runClean,
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
}

func runClean(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	flagDocker, _ := cmd.Flags().GetBool("docker")
	flagNode, _ := cmd.Flags().GetBool("node")
	flagCache, _ := cmd.Flags().GetBool("cache")
	flagGit, _ := cmd.Flags().GetBool("git")
	flagLogs, _ := cmd.Flags().GetBool("logs")
	flagAll, _ := cmd.Flags().GetBool("all")
	flagDryRun, _ := cmd.Flags().GetBool("dry-run")
	flagYes, _ := cmd.Flags().GetBool("yes")
	flagTrash, _ := cmd.Flags().GetBool("trash")
	flagNoTrash, _ := cmd.Flags().GetBool("no-trash")

	noColor := false
	verbose := false
	staleDays := 90
	workers := 4
	gitTimeoutSecs := 30
	jsonOutput := false
	if GlobalConfig != nil {
		noColor = GlobalConfig.NoColor
		verbose = GlobalConfig.Verbose
		staleDays = GlobalConfig.StaleDays
		workers = GlobalConfig.Workers
		gitTimeoutSecs = GlobalConfig.GitTimeoutSecs
		jsonOutput = GlobalConfig.JSON
		if GlobalConfig.Trash {
			flagTrash = true
		}
	}

	noColorFlag, _ := cmd.Root().PersistentFlags().GetBool("no-color")
	caps := tui.Detect(noColor || noColorFlag)
	if jsonOutput {
		flagYes = true // non-interactive when emitting JSON
	}

	useTrash := flagTrash && !flagNoTrash

	// Resolve which categories to clean.
	categories := resolveCategories(flagAll, flagNode, flagCache, flagGit, flagLogs)
	needFileScan := len(categories) > 0
	needDocker := flagDocker || flagAll

	if !needFileScan && !needDocker {
		return cmd.Help()
	}

	home, _ := os.UserHomeDir()

	// Lower I/O priority for the re-scan.
	_ = syscall.Setpriority(syscall.PRIO_PROCESS, 0, 10)

	// Re-scan to get fresh candidates.
	var scanItems []scan.CandidateItem
	if needFileScan {
		spinnerCaps := caps
		if jsonOutput {
			spinnerCaps.IsTTY = false
		}
		stopSpinner := tui.StartSpinner("Scanning", spinnerCaps)
		scanCfg := scan.Config{
			Root:      "",
			Deep:      true,
			StaleDays: staleDays,
			Workers:   workers,
		}
		result, err := scan.Scan(ctx, scanCfg)
		stopSpinner()
		if err != nil && ctx.Err() == nil {
			return fmt.Errorf("scan failed: %w\nFix: check that the target path exists and is readable", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		scanItems = result.Items
	}

	// Build candidate list filtered to requested categories.
	var toClean []scan.CandidateItem
	catSet := make(map[string]bool, len(categories))
	for _, c := range categories {
		catSet[c] = true
	}
	for _, item := range scanItems {
		if catSet[item.Category] {
			toClean = append(toClean, item)
		}
	}

	// Docker query.
	var dockerFreeable int64
	if needDocker {
		dResult, _ := docker.Query(ctx, home, verbose)
		dockerFreeable = dResult.UnusedImageBytes + dResult.VolumeBytes
	}

	if len(toClean) == 0 && dockerFreeable == 0 {
		if jsonOutput {
			return jsonout.WriteClean(os.Stdout, nil, 0, flagDryRun, useTrash)
		}
		fmt.Fprintln(os.Stdout, "  Nothing to clean.")
		return nil
	}

	if !jsonOutput {
		// Show preview and confirm only for interactive (non-JSON) output.
		printCleanPreview(toClean, dockerFreeable, caps, flagDryRun)

		if flagDryRun {
			return nil
		}

		if !flagYes {
			if !confirm(flagAll) {
				fmt.Fprintln(os.Stdout, "  Aborted.")
				return nil
			}
		}
	}

	// Clean file-system items.
	var cleanResults []clean.ItemResult
	if len(toClean) > 0 {
		cleanCfg := clean.Config{
			Categories:     categories,
			DryRun:         flagDryRun,
			UseTrash:       useTrash,
			GitTimeoutSecs: gitTimeoutSecs,
			Home:           home,
		}
		cleanResults = clean.Run(ctx, cleanCfg, toClean)
	}

	// Docker prune.
	var dockerFreed int64
	if needDocker && dockerFreeable > 0 && !flagDryRun {
		var dockerErr error
		dockerFreed, dockerErr = docker.PruneUnused(ctx, home, false, verbose)
		if !jsonOutput {
			if dockerErr != nil {
				fmt.Fprintf(os.Stderr, "  docker prune failed: %s\n", dockerErr)
			} else {
				gb := float64(dockerFreed) / (1 << 30)
				fmt.Fprintf(os.Stdout, "  Docker: freed %.1f GB\n", gb)
			}
		}
	} else if flagDryRun && needDocker {
		dockerFreed = dockerFreeable
	}

	if jsonOutput {
		return jsonout.WriteClean(os.Stdout, cleanResults, dockerFreed, flagDryRun, useTrash)
	}

	totalFreed, errCount := printCleanResults(cleanResults, caps)
	totalFreed += dockerFreed

	fmt.Fprintln(os.Stdout)
	if errCount > 0 {
		fmt.Fprintf(os.Stdout, "  Done with %d error(s). Run with --verbose for details.\n", errCount)
	} else {
		gb := float64(totalFreed) / (1 << 30)
		fmt.Fprintf(os.Stdout, "  Done. Freed ~%.1f GB.\n", gb)
	}

	return nil
}

// resolveCategories maps flag booleans to category name slices.
func resolveCategories(all, node, cache, git, logs bool) []string {
	var cats []string
	add := func(c ...string) { cats = append(cats, c...) }

	if all || node {
		add(scan.CatNodeModules)
	}
	if all || cache {
		add(scan.CatBrewCache, scan.CatPipCache, scan.CatNpmCache, scan.CatAptCache, scan.CatXcodeDerived, scan.CatPycache)
	}
	if all || git {
		add(scan.CatGitObjects)
	}
	if all || logs {
		add(scan.CatLogs, scan.CatJournald)
	}
	if all {
		add(scan.CatTrash, scan.CatDownloads)
	}
	return cats
}

// printCleanPreview shows what would be cleaned.
func printCleanPreview(items []scan.CandidateItem, dockerBytes int64, caps tui.Caps, isDryRun bool) {
	mode := "Preview"
	if isDryRun {
		mode = "Dry Run"
	}

	disk := "💽"
	if !caps.Emoji {
		disk = "[disk]"
	}
	fmt.Fprintf(os.Stdout, "\n%s  diskwhy — Clean %s\n\n", disk, mode)

	// Aggregate by category.
	type agg struct {
		total int64
		count int
	}
	byCategory := make(map[string]*agg)
	var order []string
	for _, item := range items {
		a, ok := byCategory[item.Category]
		if !ok {
			a = &agg{}
			byCategory[item.Category] = a
			order = append(order, item.Category)
		}
		a.total += item.SizeBytes
		a.count++
	}

	var maxBytes int64
	for _, cat := range order {
		if v := byCategory[cat].total; v > maxBytes {
			maxBytes = v
		}
	}
	if dockerBytes > maxBytes {
		maxBytes = dockerBytes
	}

	var totalBytes int64
	for _, cat := range order {
		a := byCategory[cat]
		label := tui.CategoryLabel[cat]
		if label == "" {
			label = cat
		}
		emoji := tui.CategoryEmoji[cat]
		fmt.Fprintln(os.Stdout, tui.CategoryLine(label, emoji, a.total, maxBytes, a.count, "", caps))
		totalBytes += a.total
	}
	if dockerBytes > 0 {
		fmt.Fprintln(os.Stdout, tui.CategoryLine("Docker (unused)", "🐳", dockerBytes, maxBytes, 1, "", caps))
		totalBytes += dockerBytes
	}

	gb := float64(totalBytes) / (1 << 30)
	fmt.Fprintf(os.Stdout, "\n  ~%.1f GB to be freed\n\n", gb)
}

// confirm asks the user to type "yes" for --all operations, or y/n otherwise.
func confirm(requireYes bool) bool {
	if requireYes {
		fmt.Fprint(os.Stdout, "  Type 'yes' to continue: ")
	} else {
		fmt.Fprint(os.Stdout, "  Proceed? [y/N]: ")
	}

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))

	if requireYes {
		return answer == "yes"
	}
	return answer == "y" || answer == "yes"
}

// printCleanResults prints one line per result and returns (bytesFreed, errCount).
func printCleanResults(results []clean.ItemResult, caps tui.Caps) (int64, int) {
	var freed int64
	var errCount int

	for _, r := range results {
		switch r.Outcome {
		case clean.OutcomeDeleted:
			gb := float64(r.BytesDelta) / (1 << 30)
			freed += r.BytesDelta
			fmt.Fprintf(os.Stdout, "  deleted  %-50s  %.1f GB\n", truncate(r.Path, 50), gb)
		case clean.OutcomeTrashed:
			freed += r.BytesDelta
			fmt.Fprintf(os.Stdout, "  trashed  %-50s\n", truncate(r.Path, 50))
		case clean.OutcomeGCRun:
			fmt.Fprintf(os.Stdout, "  git gc   %-50s\n", truncate(r.Path, 50))
		case clean.OutcomeSkipped:
			if caps.Emoji {
				fmt.Fprintf(os.Stdout, "  ⏭  skipped  %s\n", truncate(r.Path, 50))
			} else {
				fmt.Fprintf(os.Stdout, "  skipped  %s\n", truncate(r.Path, 50))
			}
		case clean.OutcomeError:
			errCount++
			fmt.Fprintf(os.Stderr, "  error    %s: %s\n", truncate(r.Path, 40), r.Err)
		}
	}
	return freed, errCount
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "..." + s[len(s)-(max-3):]
}

