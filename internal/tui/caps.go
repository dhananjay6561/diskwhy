package tui

import (
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// Caps describes what the current terminal supports. It is computed once at
// startup and passed through to all rendering functions so they never re-query
// the environment.
type Caps struct {
	IsTTY bool // stdout is an interactive terminal
	Color bool // ANSI 256-color output is safe to emit
	Emoji bool // the terminal locale is UTF-8 (emoji render correctly)
}

// Detect probes stdout for terminal capabilities.
//
// Precedence for disabling color:
//  1. noColorFlag   -- --no-color CLI flag
//  2. NO_COLOR      -- universal standard (any non-empty value)
//  3. DISKWHY_NO_COLOR -- project env var
//  4. not a TTY     -- piped or redirected output
func Detect(noColorFlag bool) Caps {
	isTTY := isatty.IsTerminal(os.Stdout.Fd()) ||
		isatty.IsCygwinTerminal(os.Stdout.Fd())

	noColor := noColorFlag ||
		os.Getenv("NO_COLOR") != "" ||
		os.Getenv("DISKWHY_NO_COLOR") != ""

	return Caps{
		IsTTY: isTTY,
		Color: isTTY && !noColor,
		Emoji: isTTY && isUTF8Locale(),
	}
}

// isUTF8Locale reports whether the active locale signals UTF-8 support.
// Checked in priority order: LC_ALL, LC_CTYPE, LANG.
func isUTF8Locale() bool {
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		v := os.Getenv(key)
		if v == "" {
			continue
		}
		upper := strings.ToUpper(v)
		if strings.Contains(upper, "UTF-8") || strings.Contains(upper, "UTF8") {
			return true
		}
	}
	return false
}
