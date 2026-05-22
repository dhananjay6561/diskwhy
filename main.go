package main

import "github.com/dhananjay6561/diskwhy/cmd"

// Injected at build time via -ldflags "-X main.version=... -X main.commit=..."
var (
	version = "dev"
	commit  = "none"
)

func main() {
	cmd.Execute(version, commit)
}
