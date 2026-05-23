package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dhananjay6561/diskwhy/internal/docker"
	"github.com/dhananjay6561/diskwhy/internal/scan"
)

// ── view states ───────────────────────────────────────────────────────────────

type appView int

const (
	viewHome appView = iota
	viewScanning
	viewScanResult
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ── messages ──────────────────────────────────────────────────────────────────

type tickMsg struct{}

type scanDoneMsg struct {
	result       *scan.Result
	dockerResult *docker.Result
	diskTotal    int64
	diskUsed     int64
	diskFree     int64
	err          error
}

type cleanDoneMsg struct{ err error }

// ── AppModel ──────────────────────────────────────────────────────────────────

// AppModel is the root bubbletea model. It owns all views.
type AppModel struct {
	view    appView
	cursor  int
	version string
	caps    Caps
	height  int
	width   int

	// scanning
	deep         bool
	spinnerFrame int

	// scan results
	scanResult   *scan.Result
	dockerResult *docker.Result
	diskTotal    int64
	diskUsed     int64
	diskFree     int64
	scanErr      string

	// computed after scan completes
	sortedItems []scan.CandidateItem
	dockerTotal int64
	maxBytes    int64
	totalFound  int64
	safeBytes   int64
	scrollTop   int
}

// NewAppModel returns an initialized AppModel ready to run.
func NewAppModel(version string, caps Caps) AppModel {
	return AppModel{
		version: version,
		caps:    caps,
		height:  24,
		width:   120,
	}
}

func (m AppModel) Init() tea.Cmd { return nil }

// ── Update ────────────────────────────────────────────────────────────────────

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case tickMsg:
		if m.view == viewScanning {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
			return m, doTick()
		}

	case scanDoneMsg:
		if msg.err != nil {
			m.scanErr = msg.err.Error()
			m.view = viewHome
			return m, nil
		}
		m.scanResult = msg.result
		m.dockerResult = msg.dockerResult
		m.diskTotal = msg.diskTotal
		m.diskUsed = msg.diskUsed
		m.diskFree = msg.diskFree
		m.computeScanData()
		m.view = viewScanResult
		m.scrollTop = 0

	case cleanDoneMsg:
		m.view = viewHome

	case tea.KeyMsg:
		return m.handleKey(msg.String())
	}

	return m, nil
}

func (m AppModel) handleKey(key string) (AppModel, tea.Cmd) {
	switch m.view {

	case viewHome:
		switch key {
		case "ctrl+c", "q", "Q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < 3 {
				m.cursor++
			}
		case "1":
			return m.startScan(false)
		case "2":
			return m.startScan(true)
		case "3":
			return m.runClean()
		case "4":
			return m, tea.Quit
		case "enter", " ":
			switch m.cursor {
			case 0:
				return m.startScan(false)
			case 1:
				return m.startScan(true)
			case 2:
				return m.runClean()
			case 3:
				return m, tea.Quit
			}
		}

	case viewScanning:
		if key == "ctrl+c" || key == "esc" {
			m.view = viewHome
		}

	case viewScanResult:
		switch key {
		case "ctrl+c", "q", "Q":
			return m, tea.Quit
		case "b", "B", "esc":
			m.view = viewHome
			m.cursor = 0
		case "up", "k":
			if m.scrollTop > 0 {
				m.scrollTop--
			}
		case "down", "j":
			max := len(m.sortedItems) - m.visibleRows()
			if max < 0 {
				max = 0
			}
			if m.scrollTop < max {
				m.scrollTop++
			}
		case "c", "C":
			return m.runClean()
		}
	}

	return m, nil
}

func (m AppModel) startScan(deep bool) (AppModel, tea.Cmd) {
	m.view = viewScanning
	m.deep = deep
	m.spinnerFrame = 0
	m.scanErr = ""
	return m, tea.Batch(doTick(), doScanCmd(deep))
}

func (m AppModel) runClean() (AppModel, tea.Cmd) {
	c := exec.Command(os.Args[0], "clean", "--node", "--cache", "--git", "--logs")
	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		return cleanDoneMsg{err: err}
	})
}

// ── async commands ─────────────────────────────────────────────────────────────

func doTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

func doScanCmd(deep bool) tea.Cmd {
	return func() tea.Msg {
		home, _ := os.UserHomeDir()
		cfg := scan.Config{
			Root:      "",
			Deep:      deep,
			StaleDays: 90,
			Workers:   4,
		}
		result, err := scan.Scan(context.Background(), cfg)
		if err != nil {
			return scanDoneMsg{err: err}
		}
		dockerResult, _ := docker.Query(context.Background(), home, false)
		total, used, free := DiskUsage(home)
		return scanDoneMsg{
			result:       result,
			dockerResult: dockerResult,
			diskTotal:    total,
			diskUsed:     used,
			diskFree:     free,
		}
	}
}

// ── computed data ─────────────────────────────────────────────────────────────

func (m *AppModel) computeScanData() {
	if m.scanResult == nil {
		return
	}
	sorted := make([]scan.CandidateItem, len(m.scanResult.Items))
	copy(sorted, m.scanResult.Items)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SizeBytes > sorted[j].SizeBytes
	})
	m.sortedItems = sorted

	var dockerTotal int64
	if m.dockerResult != nil {
		dockerTotal = m.dockerResult.UnusedImageBytes +
			m.dockerResult.UsedImageBytes + m.dockerResult.VolumeBytes
	}
	m.dockerTotal = dockerTotal

	var maxBytes int64
	for _, item := range sorted {
		if item.SizeBytes > maxBytes {
			maxBytes = item.SizeBytes
		}
	}
	if dockerTotal > maxBytes {
		maxBytes = dockerTotal
	}
	m.maxBytes = maxBytes

	var totalFound, safeBytes int64
	for _, item := range sorted {
		totalFound += item.SizeBytes
		if item.StalenessScore == scan.StalenessStale || item.StalenessScore == scan.StalenessUnused {
			safeBytes += item.SizeBytes
		}
	}
	totalFound += dockerTotal
	if m.dockerResult != nil {
		safeBytes += m.dockerResult.UnusedImageBytes
	}
	m.totalFound = totalFound
	m.safeBytes = safeBytes
}

func (m AppModel) visibleRows() int {
	// total height minus: 1 blank top + 1 header + 1 disk stats + 1 blank + 2 table header
	// + 1 docker + 1 blank + 1 summary + 1 blank + 1 sep + 1 hints = 12 overhead
	v := m.height - 14
	if v < 3 {
		v = 3
	}
	return v
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m AppModel) View() string {
	switch m.view {
	case viewScanning:
		return m.renderScanning()
	case viewScanResult:
		return m.renderScanResult()
	default:
		return m.renderHome()
	}
}

// ── home view ─────────────────────────────────────────────────────────────────

func (m AppModel) renderHome() string {
	const sepW = 52
	sep := strings.Repeat("─", sepW)

	if !m.caps.Color {
		var s strings.Builder
		s.WriteString("\n  diskwhy")
		if m.version != "" {
			s.WriteString("  v" + m.version)
		}
		s.WriteString("\n  your disk is full. but why?\n\n")
		s.WriteString("  " + strings.Repeat("-", sepW) + "\n\n")
		choices := homeChoices()
		for i, c := range choices {
			if m.cursor == i {
				s.WriteString(fmt.Sprintf("  ► %s.  %-14s %s\n", c[0], c[1], c[2]))
			} else {
				s.WriteString(fmt.Sprintf("     %s.  %-14s %s\n", c[0], c[1], c[2]))
			}
		}
		if m.scanErr != "" {
			s.WriteString("\n  error: " + m.scanErr + "\n")
		}
		s.WriteString("\n  " + strings.Repeat("-", sepW) + "\n")
		s.WriteString("  ↑↓ Navigate    Enter Select    Q Quit\n")
		return s.String()
	}

	brand := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
	numC := lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa"))
	activeLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#f1f5f9")).Bold(true)
	activeDesc := lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))
	inactiveLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	inactiveDesc := lipgloss.NewStyle().Foreground(lipgloss.Color("#374151"))
	cursorC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	hintKey := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e2e2"))
	rose := lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))

	dimSep := dim.Render("  " + sep)

	var s strings.Builder
	s.WriteString("\n")
	s.WriteString("  " + brand.Render("disk") + "why")
	if m.version != "" {
		s.WriteString("  " + dim.Render("v"+m.version))
	}
	s.WriteString("\n  " + dim.Render("your disk is full. but why?") + "\n")
	s.WriteString("\n" + dimSep + "\n\n")

	choices := homeChoices()
	for i, c := range choices {
		if m.cursor == i {
			s.WriteString(cursorC.Render("  ► "))
			s.WriteString(numC.Render(c[0] + ".  "))
			s.WriteString(activeLabel.Render(fmt.Sprintf("%-14s", c[1])))
			s.WriteString(activeDesc.Render(c[2]))
		} else {
			s.WriteString(dim.Render("     "+c[0]+".  "))
			s.WriteString(inactiveLabel.Render(fmt.Sprintf("%-14s", c[1])))
			s.WriteString(inactiveDesc.Render(c[2]))
		}
		s.WriteString("\n")
	}

	if m.scanErr != "" {
		s.WriteString("\n  " + rose.Render("error: "+m.scanErr) + "\n")
	}

	s.WriteString("\n" + dimSep + "\n")
	s.WriteString("  " +
		hintKey.Render("↑ ↓") + dim.Render(" Navigate") +
		dim.Render("  |  ") +
		hintKey.Render("Enter") + dim.Render(" Select") +
		dim.Render("  |  ") +
		hintKey.Render("Q") + dim.Render(" Quit") +
		"\n")

	return s.String()
}

func homeChoices() [][3]string {
	return [][3]string{
		{"1", "Scan", "Find space hogs on your disk"},
		{"2", "Deep Scan", "Full recursive scan  (~15s)"},
		{"3", "Clean", "Remove found junk safely"},
		{"4", "Quit", "Exit diskwhy"},
	}
}

// ── scanning view ─────────────────────────────────────────────────────────────

func (m AppModel) renderScanning() string {
	label := "Scanning your disk..."
	if m.deep {
		label = "Deep scanning your disk  (~15s)..."
	}

	if !m.caps.Color {
		frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
		return fmt.Sprintf("\n\n  %s  %s\n\n  Esc  Cancel\n", frame, label)
	}

	brand := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
	spin := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)

	frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
	return fmt.Sprintf("\n\n  %s  %s\n\n  %s\n",
		spin.Render(frame),
		brand.Render(label),
		dim.Render("Esc  Cancel"),
	)
}

// ── scan result view ──────────────────────────────────────────────────────────

func (m AppModel) renderScanResult() string {
	var s strings.Builder

	// header
	s.WriteString("\n")
	s.WriteString(Header(m.scanResult.Header, m.caps) + "\n")
	if m.diskTotal > 0 {
		s.WriteString(DiskStatsLine(m.diskTotal, m.diskUsed, m.diskFree, m.caps) + "\n")
	}
	s.WriteString("\n")

	if len(m.sortedItems) == 0 && m.dockerTotal == 0 {
		s.WriteString("  Nothing significant found.\n")
		s.WriteString("  Try option 2 (Deep Scan) for a full recursive scan.\n")
	} else {
		// table header
		s.WriteString(ScanTableHeader(m.caps) + "\n")

		// item rows with scroll
		end := m.scrollTop + m.visibleRows()
		if end > len(m.sortedItems) {
			end = len(m.sortedItems)
		}
		for _, item := range m.sortedItems[m.scrollTop:end] {
			label := CategoryLabel[item.Category]
			if label == "" {
				label = item.Category
			}
			s.WriteString(ItemLine(label, item.Path, item.SizeBytes, m.maxBytes, item.StalenessScore, "", m.caps) + "\n")
		}

		// scroll indicator
		if len(m.sortedItems) > m.visibleRows() {
			hidden := len(m.sortedItems) - m.visibleRows()
			if m.caps.Color {
				dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
				s.WriteString(dim.Render(fmt.Sprintf("\n  ↑↓ scroll  (%d more items)", hidden)) + "\n")
			} else {
				s.WriteString(fmt.Sprintf("\n  ↑↓ scroll  (%d more items)\n", hidden))
			}
		}

		// docker row
		if m.dockerResult != nil && m.dockerTotal > 0 {
			staleness := scan.StalenessActive
			if m.dockerResult.UnusedImageCount > 0 {
				staleness = scan.StalenessUnused
			}
			note := ""
			if m.dockerResult.UnusedImageCount > 0 {
				note = fmt.Sprintf("%d unused", m.dockerResult.UnusedImageCount)
			}
			s.WriteString(ItemLine("docker", "images & volumes", m.dockerTotal, m.maxBytes, staleness, note, m.caps) + "\n")
		}

		s.WriteString("\n")
		s.WriteString(ScanSummary(m.totalFound, m.safeBytes, m.caps) + "\n")
	}

	// footer hints
	const sepW = 87
	if m.caps.Color {
		dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
		hintKey := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e2e2"))
		s.WriteString("\n" + dim.Render("  "+strings.Repeat("─", sepW)) + "\n")
		s.WriteString("  " +
			hintKey.Render("B") + dim.Render(" Back") +
			dim.Render("  |  ") +
			hintKey.Render("C") + dim.Render(" Clean now") +
			dim.Render("  |  ") +
			hintKey.Render("Q") + dim.Render(" Quit") +
			"\n")
	} else {
		s.WriteString("\n  " + strings.Repeat("-", sepW) + "\n")
		s.WriteString("  B Back    C Clean now    Q Quit\n")
	}

	return s.String()
}
