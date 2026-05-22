package cmd

import (
	"fmt"
	"os"
	"sort"
	"syscall"
	"time"

	"github.com/dhananjay6561/diskwhy/internal/docker"
	"github.com/dhananjay6561/diskwhy/internal/jsonout"
	"github.com/dhananjay6561/diskwhy/internal/scan"
	"github.com/dhananjay6561/diskwhy/internal/tui"
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
	RunE: runScan,
}

func init() {
	scanCmd.Flags().Bool("deep", false, "Full recursive scan (~10-15 seconds on large disks)")
	scanCmd.Flags().StringP("path", "p", "", "Scan a specific path instead of the full disk")
}

func runScan(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	deep, _ := cmd.Flags().GetBool("deep")
	scanPath, _ := cmd.Flags().GetString("path")

	workers := 4
	staleDays := 90
	noColor := false
	verbose := false
	skipDocker := false
	jsonOutput := false
	if GlobalConfig != nil {
		workers = GlobalConfig.Workers
		staleDays = GlobalConfig.StaleDays
		noColor = GlobalConfig.NoColor
		verbose = GlobalConfig.Verbose
		skipDocker = GlobalConfig.SkipDocker
		jsonOutput = GlobalConfig.JSON
	}

	noColorFlag, _ := cmd.Root().PersistentFlags().GetBool("no-color")
	caps := tui.Detect((noColor || noColorFlag) && !jsonOutput)

	// Lower process priority so the scan does not affect the user's session.
	_ = syscall.Setpriority(syscall.PRIO_PROCESS, 0, 10)

	cfg := scan.Config{
		Root:      scanPath,
		Deep:      deep,
		StaleDays: staleDays,
		Workers:   workers,
	}

	label := "Scanning"
	if deep {
		label = "Scanning (deep mode — this may take ~15s)"
	}
	spinnerCaps := caps
	if jsonOutput {
		spinnerCaps.IsTTY = false // suppress spinner when emitting JSON
	}
	stopSpinner := tui.StartSpinner(label, spinnerCaps)

	start := time.Now()
	result, err := scan.Scan(ctx, cfg)
	elapsed := time.Since(start)
	stopSpinner()

	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("scan failed: %w\nFix: check that the target path exists and is readable", err)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	home, _ := os.UserHomeDir()
	statsPath := scanPath
	if statsPath == "" {
		statsPath = home
	}
	total, used, free := tui.DiskUsage(statsPath)

	var dockerResult *docker.Result
	if !skipDocker {
		dockerResult, _ = docker.Query(ctx, home, verbose)
	}

	if jsonOutput {
		disk := jsonout.DiskInfo{TotalBytes: total, UsedBytes: used, FreeBytes: free}
		return jsonout.WriteScan(os.Stdout, result, dockerResult, disk, elapsed.Milliseconds())
	}

	printScanResult(result, dockerResult, elapsed, total, used, free, caps, verbose)
	return nil
}

// printScanResult renders the full scan output using the TUI package.
func printScanResult(
	result *scan.Result,
	dockerResult *docker.Result,
	elapsed time.Duration,
	total, used, free int64,
	caps tui.Caps,
	verbose bool,
) {
	fmt.Fprintln(os.Stdout, tui.Header(result.Header, caps))

	if total > 0 {
		fmt.Fprintln(os.Stdout, tui.DiskStatsLine(total, used, free, caps))
	}
	fmt.Fprintln(os.Stdout)

	if verbose {
		fmt.Fprintf(os.Stdout, "  Scan mode: %s   Elapsed: %.1fs\n\n",
			result.ScanMode, elapsed.Seconds())
	}

	if len(result.Items) == 0 {
		fmt.Fprintln(os.Stdout, "  Nothing significant found.")
		fmt.Fprintln(os.Stdout, "  Run diskwhy scan --deep for a full recursive scan.")
		return
	}

	// Aggregate multiple items of the same category (e.g. many node_modules dirs).
	type agg struct {
		total     int64
		count     int
		staleness scan.StalenessLevel
	}
	byCategory := make(map[string]*agg)
	var order []string

	for _, item := range result.Items {
		a, ok := byCategory[item.Category]
		if !ok {
			a = &agg{staleness: item.StalenessScore}
			byCategory[item.Category] = a
			order = append(order, item.Category)
		}
		a.total += item.SizeBytes
		a.count++
		if item.StalenessScore > a.staleness {
			a.staleness = item.StalenessScore
		}
	}

	// Sort by total size descending.
	sort.Slice(order, func(i, j int) bool {
		return byCategory[order[i]].total > byCategory[order[j]].total
	})

	// Compute max for proportional bar scaling.
	var maxBytes int64
	for _, cat := range order {
		if v := byCategory[cat].total; v > maxBytes {
			maxBytes = v
		}
	}

	// Include Docker total in maxBytes computation so bars scale correctly.
	var dockerTotal int64
	if dockerResult != nil {
		dockerTotal = dockerResult.UnusedImageBytes + dockerResult.UsedImageBytes + dockerResult.VolumeBytes
		if dockerTotal > maxBytes {
			maxBytes = dockerTotal
		}
	}

	var safeBytes int64
	for _, cat := range order {
		a := byCategory[cat]
		label := tui.CategoryLabel[cat]
		if label == "" {
			label = cat
		}
		emoji := tui.CategoryEmoji[cat]

		note := stalenessNote(a.staleness, caps)
		fmt.Fprintln(os.Stdout, tui.CategoryLine(label, emoji, a.total, maxBytes, a.count, note, caps))

		if a.staleness == scan.StalenessStale || a.staleness == scan.StalenessUnused {
			safeBytes += a.total
		}
	}

	if dockerResult != nil && dockerTotal > 0 {
		dockerNote := dockerUnusedNote(dockerResult, caps)
		fmt.Fprintln(os.Stdout, tui.CategoryLine("Docker", "🐳", dockerTotal, maxBytes,
			dockerResult.UnusedImageCount+dockerResult.UsedImageCount+dockerResult.VolumeCount,
			dockerNote, caps))
		safeBytes += dockerResult.UnusedImageBytes
	}

	if safeBytes > 0 {
		fmt.Fprintln(os.Stdout, tui.SafeToCleanLine(safeBytes, caps))
	}
	fmt.Fprintln(os.Stdout, "  Run: diskwhy clean  to begin")

	if result.SkippedCount > 0 {
		fmt.Fprintf(os.Stderr, "\n  %d path(s) skipped — run with sudo to include system directories\n",
			result.SkippedCount)
	}
}

// dockerUnusedNote returns a display note summarising unused Docker images.
func dockerUnusedNote(d *docker.Result, caps tui.Caps) string {
	if d.UnusedImageCount == 0 {
		ok := "[OK]"
		if caps.Emoji {
			ok = "✅"
		}
		return ok + " All images in use"
	}
	warn := "[!]"
	if caps.Emoji {
		warn = "🟡"
	}
	gb := float64(d.UnusedImageBytes) / (1 << 30)
	return fmt.Sprintf("%s %d unused image(s) (%.1f GB) — safe to remove", warn, d.UnusedImageCount, gb)
}

// stalenessNote returns a short display note for a staleness level.
func stalenessNote(s scan.StalenessLevel, caps tui.Caps) string {
	ok := "[OK]"
	warn := "[!]"
	danger := "[X]"
	if caps.Emoji {
		ok = "✅"
		warn = "🟡"
		danger = "🔴"
	}
	switch s {
	case scan.StalenessActive:
		return ok + " Active"
	case scan.StalenessRecent:
		return warn + " Recent"
	case scan.StalenessStale:
		return warn + " Stale — safe to review"
	case scan.StalenessUnused:
		return danger + " Unused — safe to delete"
	default:
		return "? Unknown"
	}
}
