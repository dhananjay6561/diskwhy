package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuChoice is one item in the home menu.
type MenuChoice struct {
	Num   string
	Label string
	Desc  string
	Key   string
}

// DefaultMenuChoices is the ordered list shown on the home screen.
var DefaultMenuChoices = []MenuChoice{
	{"1", "Scan", "Find space hogs on your disk", "scan"},
	{"2", "Deep Scan", "Full recursive scan  (~15s)", "deep"},
	{"3", "Clean", "Remove found junk safely", "clean"},
	{"4", "Quit", "Exit diskwhy", "quit"},
}

// MenuModel is the bubbletea model for the interactive home menu.
type MenuModel struct {
	choices  []MenuChoice
	cursor   int
	selected string
	version  string
	caps     Caps
}

// NewMenuModel returns an initialized MenuModel.
func NewMenuModel(version string, caps Caps) MenuModel {
	return MenuModel{
		choices: DefaultMenuChoices,
		version: version,
		caps:    caps,
	}
}

// Selected returns the key of the chosen item (empty string if none).
func (m MenuModel) Selected() string { return m.selected }

func (m MenuModel) Init() tea.Cmd { return nil }

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "Q":
			m.selected = "quit"
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0]-'1')
			if idx >= 0 && idx < len(m.choices) {
				m.selected = m.choices[idx].Key
				return m, tea.Quit
			}
		case "enter", " ":
			m.selected = m.choices[m.cursor].Key
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	const sepWidth = 52
	sep := strings.Repeat("─", sepWidth)

	if !m.caps.Color {
		var s strings.Builder
		s.WriteString("\n  diskwhy")
		if m.version != "" {
			s.WriteString("  v" + m.version)
		}
		s.WriteString("\n  your disk is full. but why?\n\n")
		s.WriteString("  " + strings.Repeat("-", sepWidth) + "\n\n")
		for i, c := range m.choices {
			if m.cursor == i {
				s.WriteString(fmt.Sprintf("  ► %s.  %-14s %s\n", c.Num, c.Label, c.Desc))
			} else {
				s.WriteString(fmt.Sprintf("     %s.  %-14s %s\n", c.Num, c.Label, c.Desc))
			}
		}
		s.WriteString("\n  " + strings.Repeat("-", sepWidth) + "\n")
		s.WriteString("  ↑↓ Navigate    Enter Select    Q Quit\n")
		return s.String()
	}

	brand := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563"))
	numStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa"))
	activeLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#f1f5f9")).Bold(true)
	activeDesc := lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))
	inactiveLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	inactiveDesc := lipgloss.NewStyle().Foreground(lipgloss.Color("#374151"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	hintKey := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e2e2"))
	hintSep := dim.Render("  |  ")

	dimSep := dim.Render("  " + sep)

	var s strings.Builder
	s.WriteString("\n")

	// Header
	s.WriteString("  " + brand.Render("disk") + "why")
	if m.version != "" {
		s.WriteString("  " + dim.Render("v"+m.version))
	}
	s.WriteString("\n  " + dim.Render("your disk is full. but why?") + "\n")
	s.WriteString("\n" + dimSep + "\n\n")

	// Menu items
	for i, c := range m.choices {
		if m.cursor == i {
			s.WriteString(cursorStyle.Render("  ► "))
			s.WriteString(numStyle.Render(c.Num + ".  "))
			s.WriteString(activeLabel.Render(fmt.Sprintf("%-14s", c.Label)))
			s.WriteString(activeDesc.Render(c.Desc))
		} else {
			s.WriteString(dim.Render("     " + c.Num + ".  "))
			s.WriteString(inactiveLabel.Render(fmt.Sprintf("%-14s", c.Label)))
			s.WriteString(inactiveDesc.Render(c.Desc))
		}
		s.WriteString("\n")
	}

	// Footer
	s.WriteString("\n" + dimSep + "\n")
	s.WriteString("  " +
		hintKey.Render("↑ ↓") + dim.Render(" Navigate") +
		hintSep +
		hintKey.Render("Enter") + dim.Render(" Select") +
		hintSep +
		hintKey.Render("Q") + dim.Render(" Quit") +
		"\n")

	return s.String()
}
