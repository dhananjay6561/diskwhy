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
		spinnerCaps.IsTTY = false
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

	var dockerTotal int64
	if dockerResult != nil {
		dockerTotal = dockerResult.UnusedImageBytes + dockerResult.UsedImageBytes + dockerResult.VolumeBytes
	}

	if len(result.Items) == 0 && dockerTotal == 0 {
		fmt.Fprintln(os.Stdout, "  Nothing significant found.")
		fmt.Fprintln(os.Stdout, "  Run diskwhy scan --deep for a full recursive scan.")
		return
	}

	sorted := make([]scan.CandidateItem, len(result.Items))
	copy(sorted, result.Items)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SizeBytes > sorted[j].SizeBytes
	})

	var maxBytes int64
	for _, item := range sorted {
		if item.SizeBytes > maxBytes {
			maxBytes = item.SizeBytes
		}
	}
	if dockerTotal > maxBytes {
		maxBytes = dockerTotal
	}

	var totalFound, safeBytes int64
	for _, item := range sorted {
		totalFound += item.SizeBytes
		if item.StalenessScore == scan.StalenessStale || item.StalenessScore == scan.StalenessUnused {
			safeBytes += item.SizeBytes
		}
	}
	totalFound += dockerTotal
	if dockerResult != nil {
		safeBytes += dockerResult.UnusedImageBytes
	}

	const maxDisplay = 20
	display := sorted
	extra := 0
	if len(sorted) > maxDisplay {
		extra = len(sorted) - maxDisplay
		display = sorted[:maxDisplay]
	}

	fmt.Fprintln(os.Stdout, tui.ScanTableHeader(caps))
	for _, item := range display {
		label := tui.CategoryLabel[item.Category]
		if label == "" {
			label = item.Category
		}
		fmt.Fprintln(os.Stdout, tui.ItemLine(label, item.Path, item.SizeBytes, maxBytes, item.StalenessScore, "", caps))
	}

	if extra > 0 {
		fmt.Fprintf(os.Stdout, "\n  ... and %d more\n", extra)
	}

	if dockerResult != nil && dockerTotal > 0 {
		staleness := scan.StalenessActive
		if dockerResult.UnusedImageCount > 0 {
			staleness = scan.StalenessUnused
		}
		fmt.Fprintln(os.Stdout, tui.ItemLine("docker", "images & volumes", dockerTotal, maxBytes, staleness, dockerUnusedNote(dockerResult), caps))
	}

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, tui.ScanSummary(totalFound, safeBytes, caps))

	if result.SkippedCount > 0 {
		fmt.Fprintf(os.Stderr, "\n  %d path(s) skipped — run with sudo to include system directories\n",
			result.SkippedCount)
	}
}

func dockerUnusedNote(d *docker.Result) string {
	if d.UnusedImageCount == 0 {
		return "all images in use"
	}
	gb := float64(d.UnusedImageBytes) / (1 << 30)
	return fmt.Sprintf("%d unused (%.1f GB)", d.UnusedImageCount, gb)
}
