package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
	title  lipgloss.Style
	logo   lipgloss.Style
	sub    lipgloss.Style
	box    lipgloss.Style
	prompt lipgloss.Style
	good   lipgloss.Style
	bad    lipgloss.Style
	info   lipgloss.Style
}

func newShellUI(noColor bool) shellUI {
	if noColor {
		return shellUI{
			title:  lipgloss.NewStyle().Bold(true),
			logo:   lipgloss.NewStyle().Bold(true),
			sub:    lipgloss.NewStyle(),
			box:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2),
			prompt: lipgloss.NewStyle().Bold(true),
			good:   lipgloss.NewStyle(),
			bad:    lipgloss.NewStyle(),
			info:   lipgloss.NewStyle(),
		}
	}

	if isDarkTerminal() {
		return shellUI{
			title: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
			logo:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")),
			sub:   lipgloss.NewStyle().Foreground(lipgloss.Color("111")),
			box: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(1, 2),
			prompt: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")),
			good:   lipgloss.NewStyle().Foreground(lipgloss.Color("78")),
			bad:    lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
			info:   lipgloss.NewStyle().Foreground(lipgloss.Color("117")),
		}
	}

	return shellUI{
		title: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("25")),
		logo:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("28")),
		sub:   lipgloss.NewStyle().Foreground(lipgloss.Color("24")),
		box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("27")).
			Padding(1, 2),
		prompt: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("24")),
		good:   lipgloss.NewStyle().Foreground(lipgloss.Color("28")),
		bad:    lipgloss.NewStyle().Foreground(lipgloss.Color("160")),
		info:   lipgloss.NewStyle().Foreground(lipgloss.Color("31")),
	}
}

func (u shellUI) clearScreen() {
	fmt.Fprint(os.Stdout, "\033[2J\033[H")
}

func (u shellUI) renderHome(version string) {
	u.clearScreen()
	width := terminalWidth()
	logo := diskwhyASCIIWide()
	buddy := diskBuddyASCII()
	if width < 130 {
		logo = diskwhyASCIICompact()
	}
	if width < 90 {
		buddy = "Disk Buddy says: type /scan to start"
	}

	banner := strings.Join([]string{
		u.logo.Render(logo),
		u.sub.Render("Your disk is full. But why?"),
		u.sub.Render("Made by DJ"),
		u.info.Render(buddy),
		u.info.Render("Version: " + version),
	}, "\n")

	menu := strings.Join([]string{
		"Commands:",
		"  /scan [--deep] [--path <dir>]",
		"  /clean [--all|--node|--cache|--git|--logs] [--dry-run] [--trash] [--yes]",
		"  /version    /help    /home    /clear    /exit",
		"",
		"Try now:",
		"  /scan --deep --path /Users/dj",
	}, "\n")

	fmt.Fprintln(os.Stdout, u.responsiveBox().Render(banner+"\n\n"+menu))
}

func (u shellUI) renderHelp() {
	help := strings.Join([]string{
		"Slash commands:",
		"  /scan [--deep] [--path <dir>]  Run scan",
		"  /clean [flags]                 Run clean",
		"  /version                       Show version",
		"  /home                          Show branded home screen",
		"  /clear                         Clear terminal",
		"  /help                          Show help",
		"  /exit                          Quit shell",
		"",
		"Examples:",
		"  /scan",
		"  /scan --deep --path /Users/dj",
		"  /clean --all --dry-run --yes",
	}, "\n")
	fmt.Fprintln(os.Stdout, u.responsiveBox().Render(help))
}

func (u shellUI) responsiveBox() lipgloss.Style {
	b := u.box
	width := terminalWidth()
	contentWidth := width - b.GetHorizontalFrameSize()
	if contentWidth < 48 {
		return b
	}
	return b.Width(contentWidth)
}

func terminalWidth() int {
	if col := os.Getenv("COLUMNS"); col != "" {
		if v, err := strconv.Atoi(col); err == nil && v > 0 {
			return v
		}
	}
	return 120
}

func isDarkTerminal() bool {
	v := strings.TrimSpace(os.Getenv("COLORFGBG"))
	if v == "" {
		return true
	}
	parts := strings.Split(v, ";")
	bgStr := strings.TrimSpace(parts[len(parts)-1])
	bg, err := strconv.Atoi(bgStr)
	if err != nil {
		return true
	}
	if bg >= 0 && bg <= 7 {
		return true
	}
	if bg >= 15 {
		return false
	}
	return true
}

func diskwhyASCIIWide() string {
	return strings.TrimSpace(`
/$$$$$$$  /$$$$$$  /$$$$$$  /$$   /$$ /$$      /$$ /$$   /$$ /$$     /$$        /$$$$$$  /$$       /$$$$$$
| $$__  $$|_  $$_/ /$$__  $$| $$  /$$/| $$  /$ | $$| $$  | $$|  $$   /$$/       /$$__  $$| $$      |_  $$_/
| $$  \ $$  | $$  | $$  \__/| $$ /$$/ | $$ /$$$| $$| $$  | $$ \  $$ /$$/       | $$  \__/| $$        | $$
| $$  | $$  | $$  |  $$$$$$ | $$$$$/  | $$/$$ $$ $$| $$$$$$$$  \  $$$$/        | $$      | $$        | $$
| $$  | $$  | $$   \____  $$| $$  $$  | $$$$_  $$$$| $$__  $$   \  $$/         | $$      | $$        | $$
| $$  | $$  | $$   /$$  \ $$| $$\  $$ | $$$/ \  $$$| $$  | $$    | $$          | $$    $$| $$        | $$
| $$$$$$$/ /$$$$$$|  $$$$$$/| $$ \  $$| $$/   \  $$| $$  | $$    | $$          |  $$$$$$/| $$$$$$$$ /$$$$$$
|_______/ |______/ \______/ |__/  \__/|__/     \__/|__/  |__/    |__/           \______/ |________/|______/
`)
}


func diskwhyASCIICompact() string {
	return strings.TrimSpace(`
/$$$$$$$  /$$$$$$  /$$$$$$  /$$   /$$
| $$__  $$|_  $$_/ /$$__  $$| $$  /$$/
| $$  \ $$  | $$  | $$  \__/| $$ /$$/
| $$  | $$  | $$  |  $$$$$$ | $$$$$/
| $$  | $$  | $$   \____  $$| $$  $$
| $$  | $$  | $$   /$$  \ $$| $$\  $$
| $$$$$$$/ /$$$$$$|  $$$$$$/| $$ \  $$
|_______/ |______/ \______/ |__/  \__/

/$$      /$$ /$$   /$$ /$$     /$$
| $$  /$ | $$| $$  | $$|  $$   /$$/
| $$ /$$$| $$| $$  | $$ \  $$ /$$/
| $$/$$ $$ $$| $$$$$$$$  \  $$$$/
| $$$$_  $$$$| $$__  $$   \  $$/
| $$$/ \  $$$| $$  | $$    | $$
| $$/   \  $$| $$  | $$    | $$
|__/     \__/|__/  |__/    |__/`)
}

func diskBuddyASCII() string {
	return strings.TrimSpace(`
  .-.
 (o o)   Disk Buddy says: type /scan to start
 | O \
  \   \
   '~~~'
`)
}
