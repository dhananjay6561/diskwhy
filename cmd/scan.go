package cmd

import (
	"fmt"
	"os"
	"sort"
	"syscall"
	"time"

	"github.com/dhananjay6561/diskwhy/internal/scan"
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
	scanCmd.Flags().Bool("json", false, "Output results as JSON (schema_version: 1)")
}

func runScan(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	deep, _ := cmd.Flags().GetBool("deep")
	scanPath, _ := cmd.Flags().GetString("path")

	workers := 4 // safe default before config is loaded
	staleDays := 90
	if GlobalConfig != nil {
		workers = GlobalConfig.Workers
		staleDays = GlobalConfig.StaleDays
	}

	// Lower process priority so a disk scan doesn't affect the user's session.
	_ = syscall.Setpriority(syscall.PRIO_PROCESS, 0, 10)

	cfg := scan.Config{
		Root:      scanPath,
		Deep:      deep,
		StaleDays: staleDays,
		Workers:   workers,
	}

	start := time.Now()
	result, err := scan.Scan(ctx, cfg)
	elapsed := time.Since(start)
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("scan failed: %w\nFix: check that the target path exists and is readable", err)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	printScanResult(result, elapsed)
	return nil
}

// printScanResult renders a basic text table. Phase 3 replaces this with
// lipgloss bars and colour coding.
func printScanResult(result *scan.Result, elapsed time.Duration) {
	fmt.Fprintf(os.Stdout, "\ndiskwhy — Disk Analysis  %s\n", result.Header)
	fmt.Fprintf(os.Stdout, "Scan mode: %s   Elapsed: %.1fs\n\n", result.ScanMode, elapsed.Seconds())

	if len(result.Items) == 0 {
		fmt.Fprintln(os.Stdout, "  Nothing significant found.")
		return
	}

	// Sort by size descending.
	items := make([]scan.CandidateItem, len(result.Items))
	copy(items, result.Items)
	sort.Slice(items, func(i, j int) bool {
		return items[i].SizeBytes > items[j].SizeBytes
	})

	// Aggregate: merge multiple items of the same category.
	type aggregate struct {
		total    int64
		count    int
		staleness scan.StalenessLevel
	}
	agg := make(map[string]*aggregate)
	order := []string{}
	for _, item := range items {
		a, ok := agg[item.Category]
		if !ok {
			a = &aggregate{staleness: item.StalenessScore}
			agg[item.Category] = a
			order = append(order, item.Category)
		}
		a.total += item.SizeBytes
		a.count++
		// Keep the worst (most stale) staleness signal.
		if item.StalenessScore > a.staleness {
			a.staleness = item.StalenessScore
		}
	}

	fmt.Fprintf(os.Stdout, "  %-20s  %10s  %6s  %s\n", "Category", "Size", "Items", "Freshness")
	fmt.Fprintf(os.Stdout, "  %s\n", dashes(60))

	var totalSafe int64
	for _, cat := range order {
		a := agg[cat]
		gb := float64(a.total) / (1 << 30)
		freshness := a.staleness.String()
		fmt.Fprintf(os.Stdout, "  %-20s  %7.1f GB  %6d  %s\n", cat, gb, a.count, freshness)
		if a.staleness == scan.StalenessStale || a.staleness == scan.StalenessUnused {
			totalSafe += a.total
		}
	}

	fmt.Fprintf(os.Stdout, "\n  Estimated safe to clean: ~%.1f GB\n", float64(totalSafe)/(1<<30))
	fmt.Fprintln(os.Stdout, "  Run: diskwhy clean  to begin")

	if result.SkippedCount > 0 {
		fmt.Fprintf(os.Stderr, "\n  %d path(s) skipped — run with sudo to include system directories\n",
			result.SkippedCount)
	}
}

func dashes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '-'
	}
	return string(b)
}
