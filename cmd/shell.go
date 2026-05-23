package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open interactive slash-command shell",
	Long: `Open an interactive shell similar to chat-style CLIs.

Examples:
  /scan
  /scan --deep --path /Users/dj
  /clean --all --dry-run --yes
  /help
  /exit`,
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

	ui := newShellUI(noColor)
	ui.renderHome(cmd.Root().Version)

	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(os.Stdout, "\n"+ui.prompt.Render("diskwhy> "))
		if !reader.Scan() {
			if err := reader.Err(); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "\n"+ui.good.Render("bye"))
			return nil
		}

		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "/") {
			fmt.Fprintln(os.Stdout, "Commands should start with '/'. Try /help")
			continue
		}

		tokens := strings.Fields(strings.TrimPrefix(line, "/"))
		if len(tokens) == 0 {
			continue
		}

		name := tokens[0]
		args := tokens[1:]

		switch name {
		case "help":
			ui.renderHelp()
		case "home":
			ui.renderHome(cmd.Root().Version)
		case "clear":
			ui.clearScreen()
		case "exit", "quit", "q":
			fmt.Fprintln(os.Stdout, ui.good.Render("bye"))
			return nil
		case "version":
			fmt.Fprintln(os.Stdout, ui.info.Render("diskwhy "+cmd.Root().Version))
		case "scan":
			if err := runShellScan(cmd, args); err != nil {
				fmt.Fprintln(os.Stderr, ui.bad.Render(err.Error()))
			}
		case "clean":
			if err := runShellClean(cmd, args); err != nil {
				fmt.Fprintln(os.Stderr, ui.bad.Render(err.Error()))
			}
		default:
			fmt.Fprintln(os.Stdout, ui.bad.Render(fmt.Sprintf("Unknown command '/%s'. Try /help", name)))
		}
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

type shellUI struct {
	prompt  lipgloss.Style
	good    lipgloss.Style
	bad     lipgloss.Style
	info    lipgloss.Style
	noColor bool
}

func newShellUI(noColor bool) shellUI {
	if noColor {
		return shellUI{
			prompt:  lipgloss.NewStyle().Bold(true),
			good:    lipgloss.NewStyle(),
			bad:     lipgloss.NewStyle(),
			info:    lipgloss.NewStyle(),
			noColor: true,
		}
	}
	return shellUI{
		prompt:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#22c55e")),
		good:    lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")),
		bad:     lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171")),
		info:    lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8")),
		noColor: false,
	}
}

func (u shellUI) clearScreen() {
	fmt.Fprint(os.Stdout, "\033[2J\033[H")
}

func (u shellUI) renderHome(version string) {
	u.clearScreen()

	if u.noColor {
		fmt.Fprintln(os.Stdout, "\n  diskwhy  ·  your disk is full. but why?")
		if version != "" {
			fmt.Fprintf(os.Stdout, "  v%s\n", version)
		}
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "  /scan   [--deep] [--path <dir>]")
		fmt.Fprintln(os.Stdout, "  /clean  [--all] [--node] [--cache] [--git] [--logs] [--dry-run] [--yes]")
		fmt.Fprintln(os.Stdout, "  /help  ·  /version  ·  /clear  ·  /exit")
		fmt.Fprintln(os.Stdout)
		return
	}

	brand := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
	cmd := lipgloss.NewStyle().Foreground(lipgloss.Color("#f1f5f9"))

	fmt.Fprintln(os.Stdout)
	fmt.Fprintf(os.Stdout, "  %s%s", brand.Render("disk"), "why")
	if version != "" {
		fmt.Fprintf(os.Stdout, "  %s", dim.Render("v"+version))
	}
	fmt.Fprintln(os.Stdout, "  "+dim.Render("·")+"  "+dim.Render("your disk is full. but why?"))
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, cmd.Render("  /scan   [--deep] [--path <dir>]"))
	fmt.Fprintln(os.Stdout, cmd.Render("  /clean  [--all] [--node] [--cache] [--git] [--logs] [--dry-run] [--yes]"))
	fmt.Fprintln(os.Stdout, dim.Render("  /help  ·  /version  ·  /clear  ·  /exit"))
	fmt.Fprintln(os.Stdout)
}

func (u shellUI) renderHelp() {
	if u.noColor {
		fmt.Fprintln(os.Stdout, "\n  /scan   [--deep] [--path <dir>]")
		fmt.Fprintln(os.Stdout, "  /clean  [--all] [--node] [--cache] [--git] [--logs] [--dry-run] [--yes]")
		fmt.Fprintln(os.Stdout, "  /version  /home  /clear  /help  /exit")
		return
	}

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
	cmd := lipgloss.NewStyle().Foreground(lipgloss.Color("#f1f5f9"))

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, cmd.Render("  /scan   [--deep] [--path <dir>]"))
	fmt.Fprintln(os.Stdout, cmd.Render("  /clean  [--all] [--node] [--cache] [--git] [--logs] [--dry-run] [--yes]"))
	fmt.Fprintln(os.Stdout, dim.Render("  /version  /home  /clear  /help  /exit"))
}
