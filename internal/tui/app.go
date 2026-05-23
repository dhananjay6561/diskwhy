package tui

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dhananjay6561/diskwhy/internal/clean"
	"github.com/dhananjay6561/diskwhy/internal/docker"
	"github.com/dhananjay6561/diskwhy/internal/scan"
)

// ── view states ───────────────────────────────────────────────────────────────

type appView int

const (
	viewHome appView = iota
	viewScanning
	viewScanResult
	viewCleanConfirm
	viewCleaning
	viewCleanDone
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

type cleanDoneMsg struct {
	results []clean.ItemResult
	freed   int64
	err     error
}

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

	// clean flow
	cleanCursor  int
	cleanToggle  [4]bool
	cleanResults []clean.ItemResult
	cleanFreed   int64
	cleanErr     string
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
		if m.view == viewScanning || m.view == viewCleaning {
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
		m.cleanResults = msg.results
		m.cleanFreed = msg.freed
		if msg.err != nil {
			m.cleanErr = msg.err.Error()
		}
		m.view = viewCleanDone

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

	case viewCleanConfirm:
		switch key {
		case "ctrl+c", "q", "Q":
			return m, tea.Quit
		case "up", "k":
			if m.cleanCursor > 0 {
				m.cleanCursor--
			}
		case "down", "j":
			if m.cleanCursor < 3 {
				m.cleanCursor++
			}
		case " ":
			m.cleanToggle[m.cleanCursor] = !m.cleanToggle[m.cleanCursor]
		case "enter":
			if m.cleanToggle[0] || m.cleanToggle[1] || m.cleanToggle[2] || m.cleanToggle[3] {
				return m.startClean()
			}
		case "esc", "b", "B", "n", "N":
			m.view = viewHome
		}

	case viewCleaning:
		if key == "ctrl+c" {
			return m, tea.Quit
		}

	case viewCleanDone:
		switch key {
		case "ctrl+c", "q", "Q":
			return m, tea.Quit
		default:
			m.view = viewHome
			m.cursor = 0
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
	m.view = viewCleanConfirm
	m.cleanErr = ""
	m.cleanCursor = 0
	m.cleanToggle = [4]bool{true, true, true, true}
	return m, nil
}

func (m AppModel) startClean() (AppModel, tea.Cmd) {
	m.view = viewCleaning
	m.spinnerFrame = 0
	return m, tea.Batch(doTick(), doCleanCmd(m.cleanToggle))
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

func doCleanCmd(toggle [4]bool) tea.Cmd {
	return func() tea.Msg {
		home, _ := os.UserHomeDir()
		var categories []string
		if toggle[0] {
			categories = append(categories, scan.CatNodeModules)
		}
		if toggle[1] {
			categories = append(categories, scan.CatBrewCache, scan.CatPipCache, scan.CatNpmCache, scan.CatAptCache, scan.CatXcodeDerived, scan.CatPycache)
		}
		if toggle[2] {
			categories = append(categories, scan.CatGitObjects)
		}
		if toggle[3] {
			categories = append(categories, scan.CatLogs, scan.CatJournald)
		}
		catSet := make(map[string]bool)
		for _, c := range categories {
			catSet[c] = true
		}
		cfg := scan.Config{Root: "", Deep: true, StaleDays: 90, Workers: 4}
		result, err := scan.Scan(context.Background(), cfg)
		if err != nil {
			return cleanDoneMsg{err: err}
		}
		var toClean []scan.CandidateItem
		for _, item := range result.Items {
			if catSet[item.Category] {
				toClean = append(toClean, item)
			}
		}
		if len(toClean) == 0 {
			return cleanDoneMsg{}
		}
		cleanCfg := clean.Config{
			Categories:     categories,
			DryRun:         false,
			UseTrash:       false,
			GitTimeoutSecs: 30,
			Home:           home,
		}
		results := clean.Run(context.Background(), cleanCfg, toClean)
		var freed int64
		for _, r := range results {
			freed += r.BytesDelta
		}
		return cleanDoneMsg{results: results, freed: freed}
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
	case viewCleanConfirm:
		return m.renderCleanConfirm()
	case viewCleaning:
		return m.renderCleaning()
	case viewCleanDone:
		return m.renderCleanDone()
	default:
		return m.renderHome()
	}
}

// ── home view ─────────────────────────────────────────────────────────────────

func (m AppModel) renderHome() string {
	termW := m.width
	if termW < 4 {
		termW = 80
	}
	boxW := termW - 2
	if boxW > 118 {
		boxW = 118
	}
	if boxW < 58 {
		boxW = 58
	}

	leftInner := 26
	rightInner := boxW - 2 - leftInner - 3 // 2 for outer borders, 3 for " │ "
	if rightInner < 28 {
		rightInner = 28
	}

	versionStr := ""
	if m.version != "" {
		versionStr = " v" + m.version
	}
	username := homeUsername()
	choices := homeChoices()

	if !m.caps.Color {
		// plain fallback
		var s strings.Builder
		title := "diskwhy" + versionStr
		topFill := boxW - 2 - len(title) - 2
		if topFill < 0 {
			topFill = 0
		}
		s.WriteString("╭─ " + title + " " + strings.Repeat("─", topFill) + "╮\n")
		rows := buildHomePlainRows(m, leftInner, rightInner, username, choices)
		for _, r := range rows {
			l := plainPad(r[0], leftInner)
			ri := plainPad(r[1], rightInner)
			s.WriteString("│ " + l + " │ " + ri + " │\n")
		}
		s.WriteString("╰" + strings.Repeat("─", leftInner+2) + "┴" + strings.Repeat("─", rightInner+2) + "╯\n")
		return "\n" + s.String()
	}

	// Colors
	borderC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e"))
	titleC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	dimC := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
	brandC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	headerC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	cursorC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	numC := lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa"))
	activeLabelC := lipgloss.NewStyle().Foreground(lipgloss.Color("#f1f5f9")).Bold(true)
	activeDescC := lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))
	inactiveLabelC := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	inactiveDescC := lipgloss.NewStyle().Foreground(lipgloss.Color("#374151"))
	mascotC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e"))
	roseC := lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))
	hintKeyC := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e2e2"))

	// Left column lines
	left := []string{
		"",
		mascotC.Render("   ┌──────────┐"),
		mascotC.Render("   │ ◉  ────  │"),
		mascotC.Render("   │    ────  │"),
		mascotC.Render("   │    ────  │"),
		mascotC.Render("   └──────────┘"),
		"",
		"   Welcome back, " + brandC.Render(username) + "!",
		dimC.Render("   your disk is full."),
		dimC.Render("   but why?"),
		"",
	}

	// Right column lines
	right := []string{
		"",
		headerC.Render("  Quick start"),
		dimC.Render("  Select an option to get started."),
		"",
	}
	for i, c := range choices {
		var row string
		if m.cursor == i {
			row = "  " + cursorC.Render("► ") + numC.Render(c[0]+"  ") +
				activeLabelC.Render(plainPad(c[1], 14)) + activeDescC.Render(c[2])
		} else {
			row = "  " + dimC.Render("  "+c[0]+"  ") +
				inactiveLabelC.Render(plainPad(c[1], 14)) + inactiveDescC.Render(c[2])
		}
		right = append(right, row)
	}
	if m.scanErr != "" {
		right = append(right, "", roseC.Render("  error: "+m.scanErr))
	}
	right = append(right,
		"",
		headerC.Render("  Navigation"),
		"  "+hintKeyC.Render("↑ ↓")+" "+dimC.Render("Move")+
			dimC.Render("    ")+hintKeyC.Render("Enter / 1–4")+" "+dimC.Render("Select")+
			dimC.Render("    ")+hintKeyC.Render("Q")+" "+dimC.Render("Quit"),
		"",
	)

	// Equalize heights
	for len(left) < len(right) {
		left = append(left, "")
	}
	for len(right) < len(left) {
		right = append(right, "")
	}

	// Build title bar
	titleText := " diskwhy" + versionStr + " "
	topFill := boxW - 2 - len(titleText) - 2
	if topFill < 0 {
		topFill = 0
	}
	topBar := borderC.Render("╭─") + titleC.Render(titleText) + borderC.Render(strings.Repeat("─", topFill)+"╮")
	botBar := borderC.Render("╰" + strings.Repeat("─", leftInner+2) + "┴" + strings.Repeat("─", rightInner+2) + "╯")

	bord := borderC.Render("│")
	div := borderC.Render("│")

	var sb strings.Builder
	sb.WriteString("\n" + topBar + "\n")
	for i := range left {
		l := ansiPad(left[i], leftInner)
		r := ansiPad(right[i], rightInner)
		sb.WriteString(bord + " " + l + " " + div + " " + r + " " + bord + "\n")
	}
	sb.WriteString(botBar + "\n")
	return sb.String()
}

func homeChoices() [][3]string {
	return [][3]string{
		{"1", "Scan", "Find space hogs on your disk"},
		{"2", "Deep Scan", "Full recursive scan  (~15s)"},
		{"3", "Clean", "Remove found junk safely"},
		{"4", "Quit", "Exit diskwhy"},
	}
}

func buildHomePlainRows(m AppModel, leftW, rightW int, username string, choices [][3]string) [][2]string {
	left := []string{
		"",
		"   ┌──────────┐",
		"   │ ◉  ────  │",
		"   │    ────  │",
		"   │    ────  │",
		"   └──────────┘",
		"",
		"   Welcome back, " + username + "!",
		"   your disk is full.",
		"   but why?",
		"",
	}
	right := []string{"", "  Quick start", "  Select an option to get started.", ""}
	for i, c := range choices {
		pfx := "    "
		if m.cursor == i {
			pfx = "  ► "
		}
		right = append(right, pfx+c[0]+"  "+plainPad(c[1], 14)+c[2])
	}
	right = append(right, "", "  Navigation", "  ↑↓ Move    Enter/1-4 Select    Q Quit", "")
	for len(left) < len(right) {
		left = append(left, "")
	}
	for len(right) < len(left) {
		right = append(right, "")
	}
	rows := make([][2]string, len(left))
	for i := range left {
		rows[i] = [2]string{left[i], right[i]}
	}
	return rows
}

// ansiPad pads s to visible width w, accounting for ANSI escape codes.
func ansiPad(s string, w int) string {
	vis := lipgloss.Width(s)
	if vis >= w {
		return s
	}
	return s + strings.Repeat(" ", w-vis)
}

// plainPad pads a plain (no ANSI) string to width w.
func plainPad(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func homeUsername() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	if u := os.Getenv("USERNAME"); u != "" {
		return u
	}
	return "there"
}

func homeShortPath(maxLen int) string {
	home, _ := os.UserHomeDir()
	if len(home) <= maxLen {
		return home
	}
	parts := strings.Split(home, "/")
	if len(parts) > 0 {
		short := "~/" + parts[len(parts)-1]
		if len(short) <= maxLen {
			return short
		}
	}
	if len(home) > maxLen {
		return "..." + home[len(home)-(maxLen-3):]
	}
	return home
}

// ── clean confirm view ────────────────────────────────────────────────────────

var cleanCategories = [4][2]string{
	{"node_modules", "stale or unused project folders"},
	{"caches", "npm, pip, brew, yarn, xcode derived"},
	{"git objects", "loose pack files  (git gc)"},
	{"logs", "compressed logs older than 7 days"},
}

func (m AppModel) renderCleanConfirm() string {
	anySelected := m.cleanToggle[0] || m.cleanToggle[1] || m.cleanToggle[2] || m.cleanToggle[3]

	if !m.caps.Color {
		var s strings.Builder
		s.WriteString("\n  diskwhy  clean\n\n")
		s.WriteString("  Select categories to clean:\n\n")
		for i, cat := range cleanCategories {
			pfx := "     "
			if m.cleanCursor == i {
				pfx = "  ►  "
			}
			box := "[ ]"
			if m.cleanToggle[i] {
				box = "[x]"
			}
			s.WriteString(pfx + box + "  " + plainPad(cat[0], 14) + cat[1] + "\n")
		}
		s.WriteString("\n  Space Toggle    Enter Proceed")
		if !anySelected {
			s.WriteString("  (select at least one)")
		}
		s.WriteString("    Esc Cancel\n")
		if m.cleanErr != "" {
			s.WriteString("\n  error: " + m.cleanErr + "\n")
		}
		return s.String()
	}

	headerC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	dimC := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
	hintKeyC := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e2e2"))
	roseC := lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))
	cursorC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	checkOnC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	checkOffC := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
	activeLabelC := lipgloss.NewStyle().Foreground(lipgloss.Color("#f1f5f9")).Bold(true)
	activeDescC := lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))
	inactiveLabelC := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	inactiveDescC := lipgloss.NewStyle().Foreground(lipgloss.Color("#374151"))
	warnC := lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b"))

	var s strings.Builder
	s.WriteString("\n  " + headerC.Render("diskwhy") + dimC.Render("  clean") + "\n\n")
	s.WriteString("  " + dimC.Render("Select categories to clean:") + "\n\n")

	for i, cat := range cleanCategories {
		active := m.cleanCursor == i
		var pfx string
		if active {
			pfx = "  " + cursorC.Render("► ") + " "
		} else {
			pfx = "      "
		}
		var box string
		if m.cleanToggle[i] {
			box = checkOnC.Render("[✓]")
		} else {
			box = checkOffC.Render("[ ]")
		}
		var label, desc string
		if active {
			label = activeLabelC.Render(plainPad(cat[0], 14))
			desc = activeDescC.Render(cat[1])
		} else {
			label = inactiveLabelC.Render(plainPad(cat[0], 14))
			desc = inactiveDescC.Render(cat[1])
		}
		s.WriteString(pfx + box + "  " + label + "  " + desc + "\n")
	}

	s.WriteString("\n  " +
		hintKeyC.Render("↑ ↓") + dimC.Render(" Move") +
		dimC.Render("   ") +
		hintKeyC.Render("Space") + dimC.Render(" Toggle") +
		dimC.Render("   ") +
		hintKeyC.Render("Enter") + dimC.Render(" Proceed") +
		dimC.Render("   ") +
		hintKeyC.Render("Esc") + dimC.Render(" Cancel") +
		"\n")
	if !anySelected {
		s.WriteString("  " + warnC.Render("select at least one category") + "\n")
	}
	if m.cleanErr != "" {
		s.WriteString("\n  " + roseC.Render("error: "+m.cleanErr) + "\n")
	}
	return s.String()
}

// ── cleaning view ─────────────────────────────────────────────────────────────

func (m AppModel) renderCleaning() string {
	label := "Scanning and cleaning..."

	if !m.caps.Color {
		frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
		return fmt.Sprintf("\n\n  %s  %s\n\n  Ctrl+C  Quit\n", frame, label)
	}

	brand := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
	spin := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)

	frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
	return fmt.Sprintf("\n\n  %s  %s\n\n  %s\n",
		spin.Render(frame),
		brand.Render(label),
		dim.Render("Ctrl+C  Quit"),
	)
}

// ── clean done view ───────────────────────────────────────────────────────────

func (m AppModel) renderCleanDone() string {
	var ops, skipped, partial, errCount int
	for _, r := range m.cleanResults {
		switch r.Outcome {
		case clean.OutcomeDeleted, clean.OutcomeTrashed, clean.OutcomeGCRun:
			ops++
		case clean.OutcomeSkipped:
			skipped++
		case clean.OutcomePartial:
			partial++
		case clean.OutcomeError:
			errCount++
		}
	}

	if !m.caps.Color {
		var s strings.Builder
		s.WriteString("\n  diskwhy  clean  done\n\n")
		if m.cleanErr != "" {
			s.WriteString("  error: " + m.cleanErr + "\n\n")
		} else if len(m.cleanResults) == 0 {
			s.WriteString("  Nothing to clean.\n\n")
		} else {
			s.WriteString(CleanSummary(m.cleanFreed, ops, skipped, partial, errCount, m.caps) + "\n\n")
		}
		s.WriteString("  Press any key to return home\n")
		return s.String()
	}

	headerC := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	dimC := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
	hintKeyC := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e2e2"))
	roseC := lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))

	var s strings.Builder
	s.WriteString("\n  " + headerC.Render("diskwhy") + dimC.Render("  clean  done") + "\n\n")
	if m.cleanErr != "" {
		s.WriteString("  " + roseC.Render("error: "+m.cleanErr) + "\n\n")
	} else if len(m.cleanResults) == 0 {
		s.WriteString("  " + dimC.Render("Nothing to clean.") + "\n\n")
	} else {
		s.WriteString(CleanSummary(m.cleanFreed, ops, skipped, partial, errCount, m.caps) + "\n\n")
	}
	s.WriteString("  " + hintKeyC.Render("any key") + dimC.Render("  Return home") + "\n")
	return s.String()
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
