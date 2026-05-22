package tui

import (
	"strings"
	"testing"
)

func TestBar_Full(t *testing.T) {
	caps := Caps{Color: false}
	got := Bar(100, 100, caps)
	if len(got) != barWidth {
		t.Errorf("bar length = %d, want %d", len(got), barWidth)
	}
	if !strings.Contains(got, strings.Repeat(asciiFill, barWidth)) {
		t.Errorf("full bar should be all fill chars, got %q", got)
	}
}

func TestBar_Empty(t *testing.T) {
	caps := Caps{Color: false}
	got := Bar(0, 100, caps)
	if len(got) != barWidth {
		t.Errorf("bar length = %d, want %d", len(got), barWidth)
	}
	if !strings.Contains(got, strings.Repeat(asciiEmpty, barWidth)) {
		t.Errorf("zero bar should be all empty chars, got %q", got)
	}
}

func TestBar_Half(t *testing.T) {
	caps := Caps{Color: false}
	got := Bar(50, 100, caps)
	if len(got) != barWidth {
		t.Errorf("half bar length = %d, want %d", len(got), barWidth)
	}
}

func TestBar_ZeroMax(t *testing.T) {
	caps := Caps{Color: false}
	got := Bar(0, 0, caps)
	if len(got) != barWidth {
		t.Errorf("zero-max bar length = %d, want %d", len(got), barWidth)
	}
}

func TestBar_SmallNonZero(t *testing.T) {
	caps := Caps{Color: false}
	// 1 byte out of 1 GB — ratio rounds to 0 but should show minimum 1 fill
	got := Bar(1, 1<<30, caps)
	if !strings.Contains(got, string([]byte(asciiFill)[0:1])) {
		t.Errorf("tiny bar should show at least one fill char, got %q", got)
	}
}

func TestSizeStr_1GB(t *testing.T) {
	caps := Caps{Color: false}
	got := SizeStr(1<<30, caps)
	if !strings.Contains(got, "1.0 GB") {
		t.Errorf("1 GB SizeStr = %q, want to contain '1.0 GB'", got)
	}
}

func TestSizeStr_10GB(t *testing.T) {
	caps := Caps{Color: false}
	got := SizeStr(10<<30, caps)
	if !strings.Contains(got, "10.0 GB") {
		t.Errorf("10 GB SizeStr = %q, want to contain '10.0 GB'", got)
	}
}

func TestSizeStr_Color(t *testing.T) {
	caps := Caps{Color: true}
	// With color the string still contains the numeric portion.
	got := SizeStr(1<<30, caps)
	if !strings.Contains(got, "1.0") {
		t.Errorf("colored SizeStr = %q, should still contain '1.0'", got)
	}
}

func TestHeader_NoEmoji(t *testing.T) {
	caps := Caps{Color: false, Emoji: false}
	got := Header("test header", caps)
	if !strings.Contains(got, "diskwhy") {
		t.Errorf("Header = %q, should contain 'diskwhy'", got)
	}
	if !strings.Contains(got, "[disk]") {
		t.Errorf("Header without emoji = %q, should contain '[disk]'", got)
	}
}

func TestHeader_Emoji(t *testing.T) {
	caps := Caps{Color: false, Emoji: true}
	got := Header("test header", caps)
	if !strings.Contains(got, "💽") {
		t.Errorf("Header with emoji = %q, should contain disk emoji", got)
	}
}

func TestDiskStatsLine(t *testing.T) {
	caps := Caps{Color: false}
	total := int64(100 << 30)
	used := int64(75 << 30)
	free := int64(25 << 30)
	got := DiskStatsLine(total, used, free, caps)
	if !strings.Contains(got, "100") {
		t.Errorf("DiskStatsLine = %q, should contain total GB", got)
	}
}

func TestCategoryLine(t *testing.T) {
	caps := Caps{Color: false, Emoji: false}
	got := CategoryLine("node_modules", "", 5<<30, 10<<30, 3, "stale", caps)
	if !strings.Contains(got, "node_modules") {
		t.Errorf("CategoryLine = %q, should contain label", got)
	}
	if !strings.Contains(got, "3 items") {
		t.Errorf("CategoryLine = %q, should contain item count", got)
	}
}

func TestSafeToCleanLine(t *testing.T) {
	caps := Caps{Color: false, Emoji: false}
	got := SafeToCleanLine(5<<30, caps)
	if !strings.Contains(got, "5.0 GB") {
		t.Errorf("SafeToCleanLine = %q, should contain GB amount", got)
	}
}

func TestSizeColor(t *testing.T) {
	// red > 10 GB
	if sizeColor(11.0) != colRed {
		t.Error("11 GB should be red")
	}
	// yellow 2-10 GB
	if sizeColor(5.0) != colYellow {
		t.Error("5 GB should be yellow")
	}
	// green < 2 GB
	if sizeColor(1.0) != colGreen {
		t.Error("1 GB should be green")
	}
}
