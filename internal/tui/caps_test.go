package tui

import (
	"os"
	"testing"
)

func TestIsUTF8Locale_ViaEnv(t *testing.T) {
	// Save and restore all locale vars.
	keys := []string{"LC_ALL", "LC_CTYPE", "LANG"}
	saved := make(map[string]string)
	for _, k := range keys {
		saved[k] = os.Getenv(k)
	}
	restore := func() {
		for _, k := range keys {
			os.Setenv(k, saved[k])
		}
	}
	defer restore()

	// Clear all, then set LC_ALL to UTF-8 — should be true.
	for _, k := range keys {
		os.Unsetenv(k)
	}
	os.Setenv("LC_ALL", "en_US.UTF-8")
	if !isUTF8Locale() {
		t.Error("isUTF8Locale should return true for en_US.UTF-8")
	}

	// Clear all, then set all to non-UTF-8 — should be false.
	for _, k := range keys {
		os.Setenv(k, "C")
	}
	if isUTF8Locale() {
		t.Error("isUTF8Locale should return false when all locales are C")
	}

	// Clear all, set only LANG — should detect UTF-8 via LANG.
	for _, k := range keys {
		os.Unsetenv(k)
	}
	os.Setenv("LANG", "en_US.UTF-8")
	if !isUTF8Locale() {
		t.Error("isUTF8Locale should return true when LANG is UTF-8")
	}
}

func TestDetect_NoColorFlag(t *testing.T) {
	caps := Detect(true)
	if caps.Color {
		t.Error("Color should be false when noColorFlag=true")
	}
}

func TestDetect_NoColorEnv(t *testing.T) {
	orig := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", orig)
	os.Setenv("NO_COLOR", "1")

	caps := Detect(false)
	if caps.Color {
		t.Error("Color should be false when NO_COLOR env is set")
	}
}
