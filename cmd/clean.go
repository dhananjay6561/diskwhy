package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"

	"github.com/dhananjay6561/diskwhy/internal/clean"
	"github.com/dhananjay6561/diskwhy/internal/docker"
	"github.com/dhananjay6561/diskwhy/internal/jsonout"
	"github.com/dhananjay6561/diskwhy/internal/scan"
	"github.com/dhananjay6561/diskwhy/internal/trash"
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
		flagYes = true
	}

	useTrash := flagTrash && !flagNoTrash

	// §6.4 — root execution warning.
	if !jsonOutput && os.Getuid() == 0 {
		fmt.Fprint(os.Stdout, "\n  ⚠️  Running as root — blocklist is your only protection. Continue? (y/N): ")
		sc := bufio.NewScanner(os.Stdin)
		if !sc.Scan() || strings.TrimSpace(strings.ToLower(sc.Text())) != "y" {
			fmt.Fprintln(os.Stdout, "  Aborted.")
			return nil
		}
	}

	// §5.5.1 — trash state machine.
	if useTrash && !trash.Available() {
		if flagYes {
			// State 5: --yes + trash unavailable → hard abort, exit 1.
			return fmt.Errorf("trash unavailable on this system\nFix: re-run with --yes --no-trash to acknowledge permanent deletion")
		}
		// State 2: interactive + trash unavailable → warn, prompt for permanent delete.
		if !jsonOutput {
			fmt.Fprintln(os.Stdout, "  ⚠️  Trash is unavailable on this system.")
			fmt.Fprint(os.Stdout, "  Delete permanently instead? (y/N): ")
			sc := bufio.NewScanner(os.Stdin)
			if !sc.Scan() || strings.TrimSpace(strings.ToLower(sc.Text())) != "y" {
				fmt.Fprintln(os.Stdout, "  Aborted.")
				return nil
			}
			useTrash = false
		}
	}

	categories := resolveCategories(flagAll, flagNode, flagCache, flagGit, flagLogs)
	needFileScan := len(categories) > 0
	needDocker := flagDocker || flagAll

	if !needFileScan && !needDocker {
		return cmd.Help()
	}

	home, _ := os.UserHomeDir()

	_ = syscall.Setpriority(syscall.PRIO_PROCESS, 0, 10)

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
		printCleanPreview(toClean, dockerFreeable, caps, flagDryRun)

		if flagDryRun {
			return nil
		}

		if !flagYes {
			if !confirm(flagAll, flagNoTrash && !flagTrash) {
				fmt.Fprintln(os.Stdout, "  Aborted.")
				return nil
			}
		}
	}

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

	var dockerFreed int64
	var dockerOps, dockerErrCount int
	if needDocker && dockerFreeable > 0 && !flagDryRun {
		var dockerErr error
		dockerFreed, dockerErr = docker.PruneUnused(ctx, home, false, verbose)
		if !jsonOutput {
			if dockerErr != nil {
				dockerErrCount++
				fmt.Fprintln(os.Stderr, tui.CleanLine("error", "docker images & volumes", dockerErr.Error(), 0, caps))
			} else {
				dockerOps++
				fmt.Fprintln(os.Stdout, tui.CleanLine("ok", "docker images & volumes", "pruned", dockerFreed, caps))
			}
		}
	} else if flagDryRun && needDocker {
		dockerFreed = dockerFreeable
	}

	if jsonOutput {
		return jsonout.WriteClean(os.Stdout, cleanResults, dockerFreed, flagDryRun, useTrash)
	}

	totalFreed, ops, skipped, partial, errCount := printCleanResults(cleanResults, caps)
	totalFreed += dockerFreed
	ops += dockerOps
	errCount += dockerErrCount

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, tui.CleanSummary(totalFreed, ops, skipped, partial, errCount, caps))

	return nil
}

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

func printCleanPreview(items []scan.CandidateItem, dockerBytes int64, caps tui.Caps, isDryRun bool) {
	mode := "preview"
	if isDryRun {
		mode = "dry-run"
	}
	fmt.Fprintf(os.Stdout, "\n  diskwhy  clean  %s\n\n", mode)

	sorted := make([]scan.CandidateItem, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SizeBytes > sorted[j].SizeBytes
	})

	var maxBytes, totalBytes int64
	for _, item := range sorted {
		if item.SizeBytes > maxBytes {
			maxBytes = item.SizeBytes
		}
		totalBytes += item.SizeBytes
	}
	if dockerBytes > maxBytes {
		maxBytes = dockerBytes
	}
	totalBytes += dockerBytes

	fmt.Fprintln(os.Stdout, tui.ScanTableHeader(caps))
	for _, item := range sorted {
		label := tui.CategoryLabel[item.Category]
		if label == "" {
			label = item.Category
		}
		fmt.Fprintln(os.Stdout, tui.ItemLine(label, item.Path, item.SizeBytes, maxBytes, item.StalenessScore, "", caps))
	}
	if dockerBytes > 0 {
		fmt.Fprintln(os.Stdout, tui.ItemLine("docker", "images & volumes", dockerBytes, maxBytes, scan.StalenessUnused, "", caps))
	}

	gb := float64(totalBytes) / (1 << 30)
	fmt.Fprintf(os.Stdout, "\n  ~%.1f GB will be freed\n\n", gb)
}

func confirm(requireYes, permanent bool) bool {
	switch {
	case requireYes:
		fmt.Fprint(os.Stdout, "  Type 'yes' to continue: ")
	case permanent:
		// State 3: --no-trash without --yes — explicit permanent-delete prompt.
		fmt.Fprint(os.Stdout, "  ⚠️  Permanent deletion — cannot be undone. Proceed? (y/N): ")
	default:
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

func printCleanResults(results []clean.ItemResult, caps tui.Caps) (freed int64, ops, skipped, partial, errCount int) {
	for _, r := range results {
		switch r.Outcome {
		case clean.OutcomeDeleted:
			freed += r.BytesDelta
			ops++
			fmt.Fprintln(os.Stdout, tui.CleanLine("ok", r.Path, "deleted", r.BytesDelta, caps))
		case clean.OutcomeTrashed:
			freed += r.BytesDelta
			ops++
			fmt.Fprintln(os.Stdout, tui.CleanLine("ok", r.Path, "trashed", r.BytesDelta, caps))
		case clean.OutcomeGCRun:
			freed += r.BytesDelta
			ops++
			fmt.Fprintln(os.Stdout, tui.CleanLine("ok", r.Path, "gc", r.BytesDelta, caps))
		case clean.OutcomeSkipped:
			skipped++
			fmt.Fprintln(os.Stdout, tui.CleanLine("skip", r.Path, "skipped", 0, caps))
		case clean.OutcomePartial:
			partial++
			msg := fmt.Sprintf("partial (%d/%d files removed)", r.FilesRemoved, r.FilesTotal)
			fmt.Fprintln(os.Stderr, tui.CleanLine("error", r.Path, msg, 0, caps))
		case clean.OutcomeError:
			errCount++
			msg := "error"
			if r.Err != nil {
				msg = r.Err.Error()
			}
			fmt.Fprintln(os.Stderr, tui.CleanLine("error", r.Path, msg, 0, caps))
		}
	}
	return
}
