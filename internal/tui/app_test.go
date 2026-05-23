package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dhananjay6561/diskwhy/internal/clean"
	"github.com/dhananjay6561/diskwhy/internal/scan"
)

func plainCaps() Caps { return Caps{Color: false, Emoji: false, IsTTY: false} }

// ── helpers ──────────────────────────────────────────────────────────────────

func TestNewAppModel(t *testing.T) {
	m := NewAppModel("1.0.0", plainCaps())
	if m.version != "1.0.0" {
		t.Errorf("version = %q, want '1.0.0'", m.version)
	}
	if m.width == 0 || m.height == 0 {
		t.Error("NewAppModel should set default dimensions")
	}
}

func TestPlainPad_short(t *testing.T) {
	got := plainPad("hi", 10)
	if len(got) != 10 {
		t.Errorf("plainPad len = %d, want 10", len(got))
	}
	if !strings.HasPrefix(got, "hi") {
		t.Error("plainPad should preserve original string")
	}
}

func TestPlainPad_exact(t *testing.T) {
	got := plainPad("hello", 5)
	if got != "hello" {
		t.Errorf("exact-length plainPad = %q, want 'hello'", got)
	}
}

func TestPlainPad_longer(t *testing.T) {
	got := plainPad("toolongstring", 4)
	if got != "toolongstring" {
		t.Error("plainPad should not truncate longer string")
	}
}

func TestAnsiPad_pads(t *testing.T) {
	got := ansiPad("hi", 10)
	if len(got) < 10 {
		t.Errorf("ansiPad result len = %d, want >= 10", len(got))
	}
}

func TestAnsiPad_noTruncate(t *testing.T) {
	long := strings.Repeat("x", 30)
	got := ansiPad(long, 5)
	if got != long {
		t.Error("ansiPad should not truncate wider string")
	}
}

func TestHomeChoices(t *testing.T) {
	choices := homeChoices()
	if len(choices) != 4 {
		t.Fatalf("homeChoices len = %d, want 4", len(choices))
	}
	wantKeys := []string{"1", "2", "3", "4"}
	for i, c := range choices {
		if c[0] != wantKeys[i] {
			t.Errorf("choice[%d] key = %q, want %q", i, c[0], wantKeys[i])
		}
	}
}

func TestHomeUsername_nonEmpty(t *testing.T) {
	u := homeUsername()
	if u == "" {
		t.Error("homeUsername should return non-empty string")
	}
}

func TestHomeShortPath_nonEmpty(t *testing.T) {
	if homeShortPath(50) == "" {
		t.Error("homeShortPath(50) should be non-empty")
	}
	if homeShortPath(5) == "" {
		t.Error("homeShortPath(5) should be non-empty")
	}
}

func TestBuildHomePlainRows(t *testing.T) {
	m := NewAppModel("", plainCaps())
	rows := buildHomePlainRows(m, 26, 40, "alice", homeChoices())
	if len(rows) == 0 {
		t.Fatal("buildHomePlainRows returned no rows")
	}
	for i, r := range rows {
		if len(r) != 2 {
			t.Errorf("row[%d] has %d columns, want 2", i, len(r))
		}
	}
}

func TestBuildHomePlainRows_cursor(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.cursor = 1
	rows := buildHomePlainRows(m, 26, 40, "alice", homeChoices())
	combined := ""
	for _, r := range rows {
		combined += r[0] + r[1]
	}
	if !strings.Contains(combined, "► ") {
		t.Error("cursor row should contain ► ")
	}
}

// ── computeScanData ───────────────────────────────────────────────────────────

func TestComputeScanData_nil(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.computeScanData()
	if m.totalFound != 0 {
		t.Errorf("totalFound = %d with nil scanResult, want 0", m.totalFound)
	}
}

func TestComputeScanData_items(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.scanResult = &scan.Result{
		Items: []scan.CandidateItem{
			{Path: "/a", SizeBytes: 2 << 30, StalenessScore: scan.StalenessUnused},
			{Path: "/b", SizeBytes: 1 << 30, StalenessScore: scan.StalenessActive},
		},
	}
	m.computeScanData()
	if m.totalFound != 3<<30 {
		t.Errorf("totalFound = %d, want %d", m.totalFound, int64(3<<30))
	}
	if m.safeBytes != 2<<30 {
		t.Errorf("safeBytes = %d (only unused+stale), want %d", m.safeBytes, int64(2<<30))
	}
	if m.maxBytes != 2<<30 {
		t.Errorf("maxBytes = %d, want %d", m.maxBytes, int64(2<<30))
	}
	if len(m.sortedItems) != 2 {
		t.Fatalf("sortedItems len = %d, want 2", len(m.sortedItems))
	}
	if m.sortedItems[0].SizeBytes < m.sortedItems[1].SizeBytes {
		t.Error("sortedItems should be descending by size")
	}
}

func TestComputeScanData_staleCountsAsSafe(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.scanResult = &scan.Result{
		Items: []scan.CandidateItem{
			{SizeBytes: 1 << 20, StalenessScore: scan.StalenessStale},
			{SizeBytes: 1 << 20, StalenessScore: scan.StalenessRecent},
		},
	}
	m.computeScanData()
	if m.safeBytes != 1<<20 {
		t.Errorf("only stale should count as safe, safeBytes = %d", m.safeBytes)
	}
}

// ── visibleRows ───────────────────────────────────────────────────────────────

func TestVisibleRows_normal(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.height = 24
	v := m.visibleRows()
	if v < 3 {
		t.Errorf("visibleRows = %d, want >= 3", v)
	}
}

func TestVisibleRows_minimum(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.height = 1
	if m.visibleRows() != 3 {
		t.Errorf("visibleRows minimum should be 3, got %d", m.visibleRows())
	}
}

// ── render functions (plain mode) ─────────────────────────────────────────────

func TestRenderHome_plain(t *testing.T) {
	m := NewAppModel("1.2.3", plainCaps())
	out := m.renderHome()
	if !strings.Contains(out, "diskwhy") {
		t.Error("renderHome should contain 'diskwhy'")
	}
	if !strings.Contains(out, "Quick start") {
		t.Error("renderHome should contain 'Quick start'")
	}
	if !strings.Contains(out, "Scan") {
		t.Error("renderHome should contain 'Scan'")
	}
}

func TestRenderHome_cursor(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.cursor = 2
	out := m.renderHome()
	if !strings.Contains(out, "► ") {
		t.Error("active cursor should render ► ")
	}
}

func TestRenderHome_scanErr(t *testing.T) {
	m := NewAppModel("", Caps{Color: true})
	m.scanErr = "disk unavailable"
	out := m.renderHome()
	if !strings.Contains(out, "disk unavailable") {
		t.Error("scanErr should appear in colored renderHome")
	}
}

func TestRenderHome_colored(t *testing.T) {
	caps := Caps{Color: true, Emoji: false, IsTTY: true}
	m := NewAppModel("dev", caps)
	out := m.renderHome()
	if out == "" {
		t.Error("colored renderHome should not be empty")
	}
}

func TestRenderScanning_shallow(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.spinnerFrame = 0
	out := m.renderScanning()
	if !strings.Contains(out, "Scanning") {
		t.Error("renderScanning should mention Scanning")
	}
	if !strings.Contains(out, "Esc") {
		t.Error("renderScanning should show Esc hint")
	}
}

func TestRenderScanning_deep(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.deep = true
	out := m.renderScanning()
	if !strings.Contains(out, "Deep") {
		t.Error("deep renderScanning should mention Deep")
	}
}

func TestRenderScanning_colored(t *testing.T) {
	m := NewAppModel("", Caps{Color: true})
	m.spinnerFrame = 2
	if m.renderScanning() == "" {
		t.Error("colored renderScanning should not be empty")
	}
}

func TestRenderCleaning_plain(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.spinnerFrame = 1
	out := m.renderCleaning()
	if !strings.Contains(out, "cleaning") {
		t.Error("renderCleaning should mention 'cleaning'")
	}
	if !strings.Contains(out, "Ctrl+C") {
		t.Error("renderCleaning should show Ctrl+C hint")
	}
}

func TestRenderCleaning_colored(t *testing.T) {
	m := NewAppModel("", Caps{Color: true})
	if m.renderCleaning() == "" {
		t.Error("colored renderCleaning should not be empty")
	}
}

func TestRenderCleanConfirm_allSelected(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.cleanToggle = [4]bool{true, true, true, true}
	out := m.renderCleanConfirm()
	if !strings.Contains(out, "[x]") {
		t.Error("selected items should show [x]")
	}
	if strings.Contains(out, "select at least one") {
		t.Error("should not warn when items selected")
	}
}

func TestRenderCleanConfirm_noneSelected(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.cleanToggle = [4]bool{false, false, false, false}
	out := m.renderCleanConfirm()
	if !strings.Contains(out, "[ ]") {
		t.Error("unselected items should show [ ]")
	}
	if !strings.Contains(out, "select at least one") {
		t.Error("should warn when nothing selected")
	}
}

func TestRenderCleanConfirm_cursor(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.cleanCursor = 1
	m.cleanToggle = [4]bool{true, true, false, false}
	out := m.renderCleanConfirm()
	if !strings.Contains(out, "► ") {
		t.Error("cursor row should show ► ")
	}
}

func TestRenderCleanConfirm_error(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.cleanErr = "scan failed badly"
	out := m.renderCleanConfirm()
	if !strings.Contains(out, "scan failed badly") {
		t.Error("cleanErr should appear in renderCleanConfirm")
	}
}

func TestRenderCleanConfirm_colored(t *testing.T) {
	m := NewAppModel("", Caps{Color: true})
	m.cleanToggle = [4]bool{true, false, true, false}
	if m.renderCleanConfirm() == "" {
		t.Error("colored renderCleanConfirm should not be empty")
	}
}

func TestRenderCleanDone_noResults(t *testing.T) {
	m := NewAppModel("", plainCaps())
	out := m.renderCleanDone()
	if !strings.Contains(out, "Nothing to clean") {
		t.Error("no results should show 'Nothing to clean'")
	}
}

func TestRenderCleanDone_withResults(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.cleanResults = []clean.ItemResult{
		{Outcome: clean.OutcomeDeleted, BytesDelta: 1 << 30},
		{Outcome: clean.OutcomeSkipped},
	}
	m.cleanFreed = 1 << 30
	out := m.renderCleanDone()
	if !strings.Contains(out, "freed") {
		t.Error("results should show 'freed'")
	}
}

func TestRenderCleanDone_withPartial(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.cleanResults = []clean.ItemResult{
		{Outcome: clean.OutcomePartial, FilesRemoved: 3, FilesTotal: 10},
	}
	out := m.renderCleanDone()
	if !strings.Contains(out, "done") {
		t.Error("renderCleanDone should contain 'done'")
	}
}

func TestRenderCleanDone_withError(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.cleanErr = "permission denied"
	out := m.renderCleanDone()
	if !strings.Contains(out, "permission denied") {
		t.Error("cleanErr should appear in renderCleanDone")
	}
}

func TestRenderCleanDone_colored(t *testing.T) {
	m := NewAppModel("", Caps{Color: true})
	m.cleanResults = []clean.ItemResult{
		{Outcome: clean.OutcomeGCRun, BytesDelta: 500 << 20},
	}
	m.cleanFreed = 500 << 20
	if m.renderCleanDone() == "" {
		t.Error("colored renderCleanDone should not be empty")
	}
}

func TestRenderScanResult_empty(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.scanResult = &scan.Result{Header: "scan header"}
	m.computeScanData()
	out := m.renderScanResult()
	if !strings.Contains(out, "Nothing significant found") {
		t.Error("empty scan should say 'Nothing significant found'")
	}
}

func TestRenderScanResult_withItems(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.height = 40
	m.scanResult = &scan.Result{
		Header: "test",
		Items: []scan.CandidateItem{
			{
				Path:           "/home/user/project/node_modules",
				Category:       scan.CatNodeModules,
				SizeBytes:      500 << 20,
				StalenessScore: scan.StalenessStale,
			},
		},
	}
	m.computeScanData()
	out := m.renderScanResult()
	if !strings.Contains(out, "node_modules") {
		t.Error("scan result should show node_modules category")
	}
	if !strings.Contains(out, "B Back") || !strings.Contains(out, "Clean now") {
		t.Error("scan result should show navigation hints")
	}
}

func TestRenderScanResult_scroll(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.height = 20
	items := make([]scan.CandidateItem, 30)
	for i := range items {
		items[i] = scan.CandidateItem{
			Path:     "/a",
			Category: scan.CatLogs,
			SizeBytes: int64(i+1) << 20,
		}
	}
	m.scanResult = &scan.Result{Header: "h", Items: items}
	m.computeScanData()
	out := m.renderScanResult()
	if !strings.Contains(out, "scroll") {
		t.Error("many items should show scroll indicator")
	}
}

func TestRenderScanResult_diskStats(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.scanResult = &scan.Result{Header: "h"}
	m.diskTotal = 500 << 30
	m.diskUsed = 300 << 30
	m.diskFree = 200 << 30
	m.computeScanData()
	out := m.renderScanResult()
	if !strings.Contains(out, "GB") {
		t.Error("disk stats line should show GB")
	}
}

// ── View dispatch ─────────────────────────────────────────────────────────────

func TestView_allViews(t *testing.T) {
	type tc struct {
		view appView
		name string
	}
	cases := []tc{
		{viewHome, "viewHome"},
		{viewScanning, "viewScanning"},
		{viewCleanConfirm, "viewCleanConfirm"},
		{viewCleaning, "viewCleaning"},
		{viewCleanDone, "viewCleanDone"},
	}
	for _, c := range cases {
		m := NewAppModel("", plainCaps())
		m.view = c.view
		out := m.View()
		if out == "" {
			t.Errorf("View() for %s returned empty string", c.name)
		}
	}
	// viewScanResult needs scanResult non-nil
	m := NewAppModel("", plainCaps())
	m.view = viewScanResult
	m.scanResult = &scan.Result{Header: "h"}
	m.computeScanData()
	if m.View() == "" {
		t.Error("View() for viewScanResult returned empty string")
	}
}

// ── handleKey ────────────────────────────────────────────────────────────────

func TestHandleKey_home_upDown(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewHome
	m.cursor = 0

	m2, _ := m.handleKey("down")
	if m2.cursor != 1 {
		t.Errorf("down from 0: cursor = %d, want 1", m2.cursor)
	}
	m3, _ := m2.handleKey("up")
	if m3.cursor != 0 {
		t.Errorf("up from 1: cursor = %d, want 0", m3.cursor)
	}
	_, _ = m3.handleKey("up") // at 0, should not go below
	m4 := m
	m4.cursor = 0
	m5, _ := m4.handleKey("up")
	if m5.cursor != 0 {
		t.Error("up at 0 should stay at 0")
	}
	m6 := m
	m6.cursor = 3
	m7, _ := m6.handleKey("down")
	if m7.cursor != 3 {
		t.Error("down at 3 should stay at 3")
	}
}

func TestHandleKey_home_jk(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewHome
	m.cursor = 1
	m2, _ := m.handleKey("k")
	if m2.cursor != 0 {
		t.Errorf("k from 1: cursor = %d, want 0", m2.cursor)
	}
	m3, _ := m2.handleKey("j")
	if m3.cursor != 1 {
		t.Errorf("j from 0: cursor = %d, want 1", m3.cursor)
	}
}

func TestHandleKey_home_numberKeys(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewHome

	m2, _ := m.handleKey("1")
	if m2.view != viewScanning {
		t.Errorf("key '1' should start scan, got view %d", m2.view)
	}
	m3, _ := m.handleKey("2")
	if m3.view != viewScanning || !m3.deep {
		t.Errorf("key '2' should start deep scan, view=%d deep=%v", m3.view, m3.deep)
	}
	m4, _ := m.handleKey("3")
	if m4.view != viewCleanConfirm {
		t.Errorf("key '3' should open clean confirm, got view %d", m4.view)
	}
}

func TestHandleKey_home_enter(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewHome
	m.cursor = 0
	m2, _ := m.handleKey("enter")
	if m2.view != viewScanning {
		t.Errorf("enter on cursor=0 should scan, got view %d", m2.view)
	}
	m.cursor = 2
	m3, _ := m.handleKey("enter")
	if m3.view != viewCleanConfirm {
		t.Errorf("enter on cursor=2 should open clean confirm, got view %d", m3.view)
	}
}

func TestHandleKey_scanning_esc(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewScanning
	m2, _ := m.handleKey("esc")
	if m2.view != viewHome {
		t.Errorf("esc in viewScanning should go home, got view %d", m2.view)
	}
}

func TestHandleKey_cleanConfirm_nav(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewCleanConfirm
	m.cleanCursor = 0
	m.cleanToggle = [4]bool{true, true, true, true}

	m2, _ := m.handleKey("down")
	if m2.cleanCursor != 1 {
		t.Errorf("down: cleanCursor = %d, want 1", m2.cleanCursor)
	}
	m3, _ := m.handleKey("j")
	if m3.cleanCursor != 1 {
		t.Errorf("j: cleanCursor = %d, want 1", m3.cleanCursor)
	}
	m4, _ := m2.handleKey("up")
	if m4.cleanCursor != 0 {
		t.Errorf("up: cleanCursor = %d, want 0", m4.cleanCursor)
	}
	m5, _ := m4.handleKey("k")
	if m5.cleanCursor != 0 {
		t.Errorf("k at 0: cleanCursor = %d, want 0", m5.cleanCursor)
	}
	m6 := m
	m6.cleanCursor = 3
	m7, _ := m6.handleKey("down")
	if m7.cleanCursor != 3 {
		t.Error("down at 3 should stay at 3")
	}
}

func TestHandleKey_cleanConfirm_toggle(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewCleanConfirm
	m.cleanCursor = 0
	m.cleanToggle = [4]bool{true, true, true, true}

	m2, _ := m.handleKey(" ")
	if m2.cleanToggle[0] {
		t.Error("space should toggle off item 0")
	}
	m3, _ := m2.handleKey(" ")
	if !m3.cleanToggle[0] {
		t.Error("space again should toggle item 0 back on")
	}
}

func TestHandleKey_cleanConfirm_enterBlocked(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewCleanConfirm
	m.cleanToggle = [4]bool{false, false, false, false}
	m2, _ := m.handleKey("enter")
	if m2.view != viewCleanConfirm {
		t.Errorf("enter with none selected should stay on confirm, got view %d", m2.view)
	}
}

func TestHandleKey_cleanConfirm_esc(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewCleanConfirm
	m2, _ := m.handleKey("esc")
	if m2.view != viewHome {
		t.Errorf("esc should go home, got view %d", m2.view)
	}
}

func TestHandleKey_cleanConfirm_bGoesHome(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewCleanConfirm
	m2, _ := m.handleKey("b")
	if m2.view != viewHome {
		t.Errorf("b should go home, got view %d", m2.view)
	}
}

func TestHandleKey_cleanDone_anyKey(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewCleanDone
	m2, _ := m.handleKey("enter")
	if m2.view != viewHome {
		t.Errorf("any key in cleanDone should go home, got view %d", m2.view)
	}
	m3, _ := m.handleKey("x")
	if m3.view != viewHome {
		t.Errorf("x in cleanDone should go home, got view %d", m3.view)
	}
}

func TestHandleKey_scanResult_back(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewScanResult
	m2, _ := m.handleKey("b")
	if m2.view != viewHome {
		t.Errorf("b should go home, got view %d", m2.view)
	}
	m3, _ := m.handleKey("esc")
	if m3.view != viewHome {
		t.Errorf("esc should go home, got view %d", m3.view)
	}
}

func TestHandleKey_scanResult_clean(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewScanResult
	m2, _ := m.handleKey("c")
	if m2.view != viewCleanConfirm {
		t.Errorf("c in scanResult should open clean confirm, got view %d", m2.view)
	}
}

func TestHandleKey_scanResult_scroll(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewScanResult
	items := make([]scan.CandidateItem, 20)
	for i := range items {
		items[i] = scan.CandidateItem{SizeBytes: int64(i+1) << 20}
	}
	m.sortedItems = items
	m.height = 24

	m2, _ := m.handleKey("down")
	if m2.scrollTop != 1 {
		t.Errorf("down scroll: scrollTop = %d, want 1", m2.scrollTop)
	}
	m3, _ := m2.handleKey("up")
	if m3.scrollTop != 0 {
		t.Errorf("up scroll: scrollTop = %d, want 0", m3.scrollTop)
	}
	// up at 0 stays at 0
	m4, _ := m3.handleKey("up")
	if m4.scrollTop != 0 {
		t.Error("up at 0 scroll should stay at 0")
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUpdate_windowSize(t *testing.T) {
	m := NewAppModel("", plainCaps())
	model, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	am := model.(AppModel)
	if am.width != 200 || am.height != 50 {
		t.Errorf("width/height = %d/%d, want 200/50", am.width, am.height)
	}
}

func TestUpdate_scanDoneMsg_success(t *testing.T) {
	m := NewAppModel("", plainCaps())
	result := &scan.Result{Header: "ok", Items: []scan.CandidateItem{}}
	model, _ := m.Update(scanDoneMsg{result: result})
	am := model.(AppModel)
	if am.view != viewScanResult {
		t.Errorf("after scan done, view should be viewScanResult, got %d", am.view)
	}
}

func TestUpdate_scanDoneMsg_error(t *testing.T) {
	m := NewAppModel("", plainCaps())
	model, _ := m.Update(scanDoneMsg{err: errors.New("disk error")})
	am := model.(AppModel)
	if am.view != viewHome {
		t.Errorf("scan error should return to viewHome, got %d", am.view)
	}
	if am.scanErr == "" {
		t.Error("scanErr should be set on scan error")
	}
}

func TestUpdate_cleanDoneMsg(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewCleaning
	model, _ := m.Update(cleanDoneMsg{freed: 500 << 20})
	am := model.(AppModel)
	if am.view != viewCleanDone {
		t.Errorf("after clean done, view should be viewCleanDone, got %d", am.view)
	}
	if am.cleanFreed != 500<<20 {
		t.Errorf("cleanFreed = %d, want %d", am.cleanFreed, int64(500<<20))
	}
}

func TestUpdate_cleanDoneMsg_withError(t *testing.T) {
	m := NewAppModel("", plainCaps())
	model, _ := m.Update(cleanDoneMsg{err: errors.New("clean failed")})
	am := model.(AppModel)
	if am.cleanErr == "" {
		t.Error("cleanErr should be set when cleanDoneMsg has error")
	}
}

func TestUpdate_tickMsg_spinnerViews(t *testing.T) {
	for _, v := range []appView{viewScanning, viewCleaning} {
		m := NewAppModel("", plainCaps())
		m.view = v
		m.spinnerFrame = 0
		model, cmd := m.Update(tickMsg{})
		am := model.(AppModel)
		if am.spinnerFrame != 1 {
			t.Errorf("view %d: tickMsg should advance spinnerFrame to 1, got %d", v, am.spinnerFrame)
		}
		if cmd == nil {
			t.Errorf("view %d: tickMsg should return next tick cmd", v)
		}
	}
}

func TestUpdate_tickMsg_nonSpinnerView(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m.view = viewHome
	m.spinnerFrame = 5
	model, _ := m.Update(tickMsg{})
	am := model.(AppModel)
	if am.spinnerFrame != 5 {
		t.Error("tickMsg in viewHome should not advance spinner")
	}
}

// ── runClean / startScan ──────────────────────────────────────────────────────

func TestRunClean(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m2, _ := m.runClean()
	if m2.view != viewCleanConfirm {
		t.Errorf("runClean view = %d, want viewCleanConfirm", m2.view)
	}
	if m2.cleanToggle != ([4]bool{true, true, true, true}) {
		t.Error("runClean should reset all toggles to true")
	}
	if m2.cleanCursor != 0 {
		t.Error("runClean should reset cleanCursor to 0")
	}
	if m2.cleanErr != "" {
		t.Error("runClean should clear cleanErr")
	}
}

func TestStartScan(t *testing.T) {
	m := NewAppModel("", plainCaps())
	m2, cmd := m.startScan(false)
	if m2.view != viewScanning {
		t.Errorf("startScan view = %d, want viewScanning", m2.view)
	}
	if m2.deep {
		t.Error("startScan(false) should not set deep")
	}
	if cmd == nil {
		t.Error("startScan should return a command")
	}
	m3, _ := m.startScan(true)
	if !m3.deep {
		t.Error("startScan(true) should set deep=true")
	}
}

func TestInit(t *testing.T) {
	m := NewAppModel("", plainCaps())
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil cmd")
	}
}
