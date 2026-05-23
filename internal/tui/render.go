package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dhananjay6561/diskwhy/internal/scan"
)

const (
	barWidth   = 16
	filledRune = "█"
	emptyRune  = "░"
	asciiFill  = "="
	asciiEmpty = "-"
)

// Color thresholds for the proportional bar (GB).
const (
	redThreshGB    = 10.0
	yellowThreshGB = 2.0
)

// Legacy ANSI-256 palette — kept for Bar/SizeStr/sizeColor used by tests.
var (
	colRed    = lipgloss.Color("196")
	colYellow = lipgloss.Color("226")
	colGreen  = lipgloss.Color("82")
	colDim    = lipgloss.Color("240")
	colBold   = lipgloss.Color("15")
)

// Truecolor palette for all new rendering functions.
var (
	tcBrand  = lipgloss.Color("#22c55e") // active / brand
	tcMint   = lipgloss.Color("#6ee7b7") // recent
	tcAmber  = lipgloss.Color("#f59e0b") // stale
	tcRose   = lipgloss.Color("#f87171") // unused / error
	tcBlue   = lipgloss.Color("#60a5fa") // category labels
	tcSlate  = lipgloss.Color("#94a3b8") // sizes
	tcDim2   = lipgloss.Color("#4b5563") // separators / bar empty
	tcWhite  = lipgloss.Color("#f1f5f9") // summary highlights
	tcViolet = lipgloss.Color("#a78bfa") // gc outcome
)

// CategoryEmoji maps category names to emoji glyphs.
var CategoryEmoji = map[string]string{
	scan.CatNodeModules:  "📦",
	scan.CatDocker:       "🐳",
	scan.CatBrewCache:    "🍺",
	scan.CatXcodeDerived: "🔧",
	scan.CatPipCache:     "🐍",
	scan.CatNpmCache:     "📦",
	scan.CatAptCache:     "📦",
	scan.CatSnap:         "📦",
	scan.CatLogs:         "📝",
	scan.CatJournald:     "📝",
	scan.CatTrash:        "🗑 ",
	scan.CatGitObjects:   "🔧",
	scan.CatPycache:      "🐍",
	scan.CatDownloads:    "📁",
}

// CategoryLabel maps category keys to display labels.
var CategoryLabel = map[string]string{
	scan.CatNodeModules:  "node_modules",
	scan.CatDocker:       "docker",
	scan.CatBrewCache:    "brew_cache",
	scan.CatXcodeDerived: "xcode_derived",
	scan.CatPipCache:     "pip_cache",
	scan.CatNpmCache:     "npm_cache",
	scan.CatAptCache:     "apt_cache",
	scan.CatSnap:         "snap",
	scan.CatLogs:         "logs",
	scan.CatJournald:     "journald",
	scan.CatTrash:        "trash",
	scan.CatGitObjects:   "git_objects",
	scan.CatPycache:      "pycache",
	scan.CatDownloads:    "downloads",
}

// ── New per-item rendering ────────────────────────────────────────────────────

// ScanTableHeader renders the column-header row and a separator line.
func ScanTableHeader(caps Caps) string {
	const catW, pathW = 14, 42
	header := fmt.Sprintf("  %-*s  %-*s  %7s  %-16s  %s",
		catW, "category",
		pathW, "path",
		"size", "usage", "status",
	)
	sep := "  " + strings.Repeat("─", catW) +
		"  " + strings.Repeat("─", pathW) +
		"  " + strings.Repeat("─", 7) +
		"  " + strings.Repeat("─", 16) +
		"  " + strings.Repeat("─", 6)
	if !caps.Color {
		return header + "\n" + sep
	}
	dim := lipgloss.NewStyle().Foreground(tcDim2)
	return dim.Render(header) + "\n" + dim.Render(sep)
}

// ItemLine renders one row of the scan table with per-item path detail.
//
// Format:  {category:14}  {path:<42}  {size:7}  {bar:16}  {staleness}
func ItemLine(category, displayPath string, sizeBytes, maxBytes int64, staleness scan.StalenessLevel, note string, caps Caps) string {
	const catW, pathW = 14, 42
	cat := fmt.Sprintf("%-*s", catW, truncStr(category, catW))
	path := fmt.Sprintf("%-*s", pathW, truncPath(displayPath, pathW))
	sz := fmtSizeCol(sizeBytes)
	bar := Bar(sizeBytes, maxBytes, caps)

	suffix := ""
	if note != "" {
		suffix = "  " + note
	}

	if !caps.Color {
		return fmt.Sprintf("  %s  %s  %s  %s  %s%s",
			cat, path, sz, bar, stalenessPlain(staleness), suffix)
	}

	catS := lipgloss.NewStyle().Foreground(tcBlue).Render(cat)
	szS := lipgloss.NewStyle().Foreground(tcSlate).Render(sz)
	staleS := stalenessColored(staleness)
	noteS := ""
	if note != "" {
		noteS = "  " + lipgloss.NewStyle().Foreground(tcDim2).Render(note)
	}
	return "  " + catS + "  " + path + "  " + szS + "  " + bar + "  " + staleS + noteS
}

// StalenessLabel returns a short colored staleness badge.
func StalenessLabel(s scan.StalenessLevel, caps Caps) string {
	if !caps.Color {
		return stalenessPlain(s)
	}
	return stalenessColored(s)
}

// CleanLine renders one result row for the clean command.
// status: "ok" | "skip" | "error"
func CleanLine(status, path, outcome string, sizeBytes int64, caps Caps) string {
	const pathW = 52
	p := fmt.Sprintf("%-*s", pathW, truncPath(path, pathW))
	sz := fmtSizeCol(sizeBytes)

	if !caps.Color {
		icon := "✓"
		switch status {
		case "skip":
			icon = "–"
		case "error":
			icon = "✗"
		}
		return fmt.Sprintf("  %s  %s  %s  %s", icon, p, sz, outcome)
	}

	var iconS, outcomeS string
	switch status {
	case "ok":
		iconS = lipgloss.NewStyle().Foreground(tcBrand).Render("✓")
		if outcome == "gc" {
			outcomeS = lipgloss.NewStyle().Foreground(tcViolet).Render(outcome)
		} else {
			outcomeS = lipgloss.NewStyle().Foreground(tcDim2).Render(outcome)
		}
	case "skip":
		iconS = lipgloss.NewStyle().Foreground(tcDim2).Render("–")
		outcomeS = lipgloss.NewStyle().Foreground(tcDim2).Render(outcome)
	case "error":
		iconS = lipgloss.NewStyle().Foreground(tcRose).Render("✗")
		outcomeS = lipgloss.NewStyle().Foreground(tcRose).Render(outcome)
	}

	szS := lipgloss.NewStyle().Foreground(tcSlate).Render(sz)
	return "  " + iconS + "  " + p + "  " + szS + "  " + outcomeS
}

// ScanSummary renders the footer line after a scan.
func ScanSummary(totalBytes, safeBytes int64, caps Caps) string {
	total := fmtGB(totalBytes)
	var line string
	if safeBytes > 0 {
		safe := fmtGB(safeBytes)
		line = fmt.Sprintf("  %s found  ·  ~%s safe to clean  ·  run diskwhy clean to begin", total, safe)
	} else {
		line = fmt.Sprintf("  %s found  ·  nothing stale  ·  run diskwhy scan --deep for more", total)
	}
	if !caps.Color {
		return line
	}
	hi := lipgloss.NewStyle().Foreground(tcWhite)
	dim := lipgloss.NewStyle().Foreground(tcDim2)
	if safeBytes > 0 {
		safe := fmtGB(safeBytes)
		return "  " + hi.Render(total+" found") + dim.Render("  ·  ~"+safe+" safe to clean  ·  run diskwhy clean to begin")
	}
	return "  " + hi.Render(total+" found") + dim.Render("  ·  nothing stale  ·  run diskwhy scan --deep for more")
}

// CleanSummary renders the footer line after a clean run.
func CleanSummary(freedBytes int64, ops, skipped, partial, errCount int, caps Caps) string {
	if errCount > 0 || partial > 0 {
		var line string
		if errCount > 0 && partial > 0 {
			line = fmt.Sprintf("  done with %d error(s), %d partial — run with --verbose for details", errCount, partial)
		} else if errCount > 0 {
			line = fmt.Sprintf("  done with %d error(s) — run with --verbose for details", errCount)
		} else {
			line = fmt.Sprintf("  done with %d partial deletion(s) — cancelled mid-directory", partial)
		}
		if !caps.Color {
			return line
		}
		return lipgloss.NewStyle().Foreground(tcRose).Render(line)
	}
	freed := fmtGB(freedBytes)
	line := fmt.Sprintf("  freed %s in %d operation(s)", freed, ops)
	if skipped > 0 {
		line += fmt.Sprintf("  (%d skipped)", skipped)
	}
	if !caps.Color {
		return line
	}
	return "  " + lipgloss.NewStyle().Foreground(tcBrand).Render("freed "+freed) +
		lipgloss.NewStyle().Foreground(tcDim2).Render(fmt.Sprintf(" in %d operation(s)", ops)) +
		func() string {
			if skipped > 0 {
				return lipgloss.NewStyle().Foreground(tcDim2).Render(fmt.Sprintf("  (%d skipped)", skipped))
			}
			return ""
		}()
}

// ── Legacy functions — kept for backward compat & tests ──────────────────────

// Bar renders a proportional fill/empty bar scaled to maxBytes.
func Bar(sizeBytes, maxBytes int64, caps Caps) string {
	filled := 0
	if maxBytes > 0 {
		ratio := float64(sizeBytes) / float64(maxBytes)
		filled = int(float64(barWidth) * ratio)
		if filled < 1 && sizeBytes > 0 {
			filled = 1
		}
	}
	empty := barWidth - filled

	if !caps.Color {
		return strings.Repeat(asciiFill, filled) + strings.Repeat(asciiEmpty, empty)
	}

	gb := float64(sizeBytes) / (1 << 30)
	fillStyle := lipgloss.NewStyle().Foreground(sizeColor(gb))
	dimStyle := lipgloss.NewStyle().Foreground(colDim)
	return fillStyle.Render(strings.Repeat(filledRune, filled)) +
		dimStyle.Render(strings.Repeat(emptyRune, empty))
}

// SizeStr formats bytes as a right-aligned "XX.X GB" string.
func SizeStr(sizeBytes int64, caps Caps) string {
	gb := float64(sizeBytes) / (1 << 30)
	text := fmt.Sprintf("%6.1f GB", gb)
	if !caps.Color {
		return text
	}
	return lipgloss.NewStyle().Foreground(sizeColor(gb)).Render(text)
}

// Header renders the compact scan header line (kept for tests).
func Header(header string, caps Caps) string {
	disk := "💽"
	if !caps.Emoji {
		disk = "[disk]"
	}
	line := fmt.Sprintf("  %s  diskwhy  %s", disk, header)
	if !caps.Color {
		return line
	}
	return lipgloss.NewStyle().Foreground(tcBrand).Bold(true).Render(line)
}

// DiskStatsLine formats the total/used/free disk line.
func DiskStatsLine(total, used, free int64, caps Caps) string {
	totalGB := float64(total) / (1 << 30)
	usedGB := float64(used) / (1 << 30)
	freeGB := float64(free) / (1 << 30)
	line := fmt.Sprintf("  %.0f GB total  ·  %.0f GB used  ·  %.0f GB free", totalGB, usedGB, freeGB)
	if !caps.Color {
		return line
	}
	return lipgloss.NewStyle().Foreground(tcDim2).Render(line)
}

// CategoryLine formats one aggregated row (used by legacy paths).
func CategoryLine(label, emoji string, sizeBytes, maxBytes int64, count int, note string, caps Caps) string {
	prefix := "  "
	if caps.Emoji && emoji != "" {
		prefix = "  " + emoji + " "
	}
	name := fmt.Sprintf("%-20s", label)
	size := SizeStr(sizeBytes, caps)
	bar := Bar(sizeBytes, maxBytes, caps)
	noteStr := ""
	if note != "" {
		noteStr = "  " + note
	}
	if count > 1 {
		noteStr = fmt.Sprintf("  %d items%s", count, noteStr)
	}
	return prefix + name + "  " + size + "  " + bar + noteStr
}

// SafeToCleanLine formats the legacy "estimated safe to clean" footer.
func SafeToCleanLine(safeBytes int64, caps Caps) string {
	gb := float64(safeBytes) / (1 << 30)
	tip := "💡"
	if !caps.Emoji {
		tip = "[!]"
	}
	line := fmt.Sprintf("\n  %s Estimated safe to clean:  ~%.1f GB", tip, gb)
	if !caps.Color {
		return line
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render(line)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func sizeColor(gb float64) lipgloss.Color {
	switch {
	case gb > redThreshGB:
		return colRed
	case gb > yellowThreshGB:
		return colYellow
	default:
		return colGreen
	}
}

func stalenessPlain(s scan.StalenessLevel) string {
	switch s {
	case scan.StalenessActive:
		return "● active"
	case scan.StalenessRecent:
		return "● recent"
	case scan.StalenessStale:
		return "▲ stale"
	case scan.StalenessUnused:
		return "✕ unused"
	default:
		return ""
	}
}

func stalenessColored(s scan.StalenessLevel) string {
	switch s {
	case scan.StalenessActive:
		return lipgloss.NewStyle().Foreground(tcBrand).Render("● active")
	case scan.StalenessRecent:
		return lipgloss.NewStyle().Foreground(tcMint).Render("● recent")
	case scan.StalenessStale:
		return lipgloss.NewStyle().Foreground(tcAmber).Render("▲ stale")
	case scan.StalenessUnused:
		return lipgloss.NewStyle().Foreground(tcRose).Render("✕ unused")
	default:
		return ""
	}
}

// fmtSizeCol formats bytes as a fixed-width 7-char size string.
func fmtSizeCol(n int64) string {
	return fmt.Sprintf("%7s", fmtBytes(n))
}

// fmtBytes formats bytes as human-readable "X.X GB" or "XXX MB".
func fmtBytes(n int64) string {
	gb := float64(n) / (1 << 30)
	if gb >= 1.0 {
		return fmt.Sprintf("%.1f GB", gb)
	}
	return fmt.Sprintf("%.0f MB", float64(n)/(1<<20))
}

// fmtGB formats bytes as "X.X GB" always.
func fmtGB(n int64) string {
	return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
}

// truncStr truncates s to max chars with a trailing "…" if needed.
func truncStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// truncPath truncates a path to max chars, keeping the tail (most informative).
func truncPath(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "…" + s[len(s)-(max-1):]
}
