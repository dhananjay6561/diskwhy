package tui

import (
	"fmt"
	"os"
	"time"
)

// braille spinner frames for UTF-8 terminals.
var brailleFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ascii spinner frames for non-UTF-8 terminals.
var asciiSpinFrames = []string{"-", "\\", "|", "/"}

// StartSpinner writes an animated spinner to stderr while work is in progress.
// It returns a stop function that must be called to clear the spinner line.
// When caps.IsTTY is false the spinner is suppressed entirely — piped output
// must stay clean.
func StartSpinner(label string, caps Caps) func() {
	if !caps.IsTTY {
		return func() {}
	}

	frames := brailleFrames
	if !caps.Emoji {
		frames = asciiSpinFrames
	}

	done := make(chan struct{})
	go func() {
		i := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				fmt.Fprint(os.Stderr, "\r\033[K") // clear spinner line
				return
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "\r%s  %s", frames[i%len(frames)], label)
				i++
			}
		}
	}()

	return func() { close(done) }
}
