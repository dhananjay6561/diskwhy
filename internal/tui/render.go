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

// Color thresholds per PRD §5.4.2.
const (
	redThreshGB    = 10.0
	yellowThreshGB = 2.0
)

// lipgloss color constants — ANSI 256-color palette.
var (
	colRed    = lipgloss.Color("196")
	colYellow = lipgloss.Color("226")
	colGreen  = lipgloss.Color("82")
	colDim    = lipgloss.Color("240")
	colBold   = lipgloss.Color("15")
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

// CategoryLabel maps category names to human-readable display labels.
var CategoryLabel = map[string]string{
	scan.CatNodeModules:  "node_modules",
	scan.CatDocker:       "Docker",
	scan.CatBrewCache:    "Brew Cache",
	scan.CatXcodeDerived: "Xcode Derived",
	scan.CatPipCache:     "pip / uv cache",
	scan.CatNpmCache:     "npm / yarn / pnpm",
	scan.CatAptCache:     "apt cache",
	scan.CatSnap:         "snap packages",
	scan.CatLogs:         "Logs",
	scan.CatJournald:     "journald logs",
	scan.CatTrash:        "Trash",
	scan.CatGitObjects:   "Git Objects",
	scan.CatPycache:      "__pycache__",
	scan.CatDownloads:    "Downloads",
}

// Bar renders a proportional fill/empty bar scaled to maxBytes.
// It emits ANSI color codes only when caps.Color is true.
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

// SizeStr formats bytes as a right-aligned "XX.X GB" string, colorised when
// caps.Color is true.
func SizeStr(sizeBytes int64, caps Caps) string {
	gb := float64(sizeBytes) / (1 << 30)
	text := fmt.Sprintf("%6.1f GB", gb)
	if !caps.Color {
		return text
	}
	return lipgloss.NewStyle().Foreground(sizeColor(gb)).Render(text)
}

// Header renders the disk analysis header line.
func Header(header string, caps Caps) string {
	disk := "💽"
	if !caps.Emoji {
		disk = "[disk]"
	}
	line := fmt.Sprintf("%s  diskwhy — Disk Analysis  %s", disk, header)
	if !caps.Color {
		return line
	}
	return lipgloss.NewStyle().Bold(true).Foreground(colBold).Render(line)
}

// DiskStatsLine formats the total/used/free line.
func DiskStatsLine(total, used, free int64, caps Caps) string {
	totalGB := float64(total) / (1 << 30)
	usedGB := float64(used) / (1 << 30)
	freeGB := float64(free) / (1 << 30)
	line := fmt.Sprintf("    %.0f GB total  •  %.0f GB used  •  %.0f GB free", totalGB, usedGB, freeGB)
	if !caps.Color {
		return line
	}
	return lipgloss.NewStyle().Foreground(colDim).Render(line)
}

// CategoryLine formats one row of the scan results table.
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

// SafeToCleanLine formats the "estimated safe to clean" footer.
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
