package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/dhananjay6561/diskwhy/internal/tui"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open interactive menu",
	Long: `Open the interactive diskwhy menu.

Navigate with ↑↓ arrows or number keys. Press Enter to select.`,
	RunE: runShell,
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

	for {
		p := tea.NewProgram(
			tui.NewMenuModel(version, caps),
			tea.WithAltScreen(),
		)
		m, err := p.Run()
		if err != nil {
			return err
		}

		result := m.(tui.MenuModel).Selected()

		switch result {
		case "scan":
			fmt.Fprintln(os.Stdout)
			if err := runShellScan(cmd, nil); err != nil {
				printShellError(err, caps)
			}
			pressEnter()

		case "deep":
			fmt.Fprintln(os.Stdout)
			if err := runShellScan(cmd, []string{"--deep"}); err != nil {
				printShellError(err, caps)
			}
			pressEnter()

		case "clean":
			fmt.Fprintln(os.Stdout)
			if err := runShellClean(cmd, []string{"--node", "--cache", "--git", "--logs"}); err != nil {
				printShellError(err, caps)
			}
			pressEnter()

		case "quit", "":
			return nil
		}
	}
}

func pressEnter() {
	fmt.Fprint(os.Stdout, "\n  Press Enter to return to menu...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func printShellError(err error, caps tui.Caps) {
	bad := lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))
	if caps.Color {
		fmt.Fprintln(os.Stderr, bad.Render("  error: "+err.Error()))
	} else {
		fmt.Fprintln(os.Stderr, "  error: "+err.Error())
	}
}

func runShellScan(parent *cobra.Command, args []string) error {
	c := &cobra.Command{Use: "scan"}
	c.SetContext(parent.Context())
	c.Flags().Bool("deep", false, "")
	c.Flags().StringP("path", "p", "", "")

	if err := c.Flags().Parse(args); err != nil {
		return err
	}
	if c.Flags().NArg() > 0 {
		return fmt.Errorf("scan: unexpected argument(s): %s", strings.Join(c.Flags().Args(), " "))
	}

	return runScan(c, nil)
}

func runShellClean(parent *cobra.Command, args []string) error {
	c := &cobra.Command{Use: "clean"}
	c.SetContext(parent.Context())
	c.Flags().Bool("docker", false, "")
	c.Flags().Bool("node", false, "")
	c.Flags().Bool("cache", false, "")
	c.Flags().Bool("git", false, "")
	c.Flags().Bool("logs", false, "")
	c.Flags().Bool("trash", false, "")
	c.Flags().Bool("all", false, "")
	c.Flags().Bool("dry-run", false, "")
	c.Flags().BoolP("yes", "y", false, "")
	c.Flags().Bool("no-trash", false, "")

	if err := c.Flags().Parse(args); err != nil {
		return err
	}
	if c.Flags().NArg() > 0 {
		return fmt.Errorf("clean: unexpected argument(s): %s", strings.Join(c.Flags().Args(), " "))
	}

	return runClean(c, nil)
}
