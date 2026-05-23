package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/dhananjay6561/diskwhy/internal/tui"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open interactive TUI",
	Long:  `Open the full-screen interactive diskwhy TUI.`,
	RunE:  runShell,
}

func runShell(cmd *cobra.Command, _ []string) error {
	if GlobalConfig != nil && GlobalConfig.JSON {
		return fmt.Errorf("shell mode is interactive and does not support --json")
	}

	noColor := false
	if GlobalConfig != nil {
		noColor = GlobalConfig.NoColor
	}
	if flag, err := cmd.Root().PersistentFlags().GetBool("no-color"); err == nil && flag {
		noColor = true
	}

	caps := tui.Detect(noColor)
	version := cmd.Root().Version

	p := tea.NewProgram(
		tui.NewAppModel(version, caps),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
