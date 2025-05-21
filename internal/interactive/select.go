package interactive

import (
	"fmt"
	"github.com/byawitz/ggh/internal/config"
	"github.com/byawitz/ggh/internal/history"
	"github.com/byawitz/ggh/internal/theme"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Selecting int

const (
	SelectConfig Selecting = iota
	SelectHistory
)

type model struct {
	table         table.Model
	choice        config.SSHConfig
	what          Selecting
	exit          bool
	action        string // New field
	selectedIndex int    // New field
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			history.RemoveByIP(m.table.SelectedRow())

			rows := slices.Delete(m.table.Rows(), m.table.Cursor(), m.table.Cursor()+1)
			m.table.SetRows(rows)

			m.table, cmd = m.table.Update("") // Overrides default `d` behavior
			return m, cmd
		case "n": // New case for 'n'
			m.action = "edit"
			m.selectedIndex = m.table.Cursor()
			return m, tea.Quit
		case "q", "ctrl+c", "esc":
			m.action = "quit"
			m.exit = true
			return m, tea.Quit
		case "enter":
			m.action = "connect"
			m.choice = setConfig(m.table.SelectedRow(), m.what)
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func setConfig(row table.Row, what Selecting) config.SSHConfig {
	// Nickname is row[1], but not used in SSHConfig for generating command args
	return config.SSHConfig{
		Host: row[2],
		Port: row[3],
		User: row[4],
		Key:  row[5],
	}
}

func (m model) View() string {
	if m.choice.Host != "" || m.exit {
		return ""
	}
	return theme.BaseStyle.Render(m.table.View()) + "\n  " + m.HelpView() + "\n"
}

func Select(rows []table.Row, what Selecting) (config.SSHConfig, string, int) {
	var columns []table.Column
	if what == SelectConfig {
		columns = append(columns, []table.Column{
			{Title: "Name", Width: 15},
			{Title: "Host", Width: 15},
			{Title: "Port", Width: 10},
			{Title: "User", Width: 10},
			{Title: "Key", Width: 10},
		}...)
	}

	if what == SelectHistory {
		columns = append(columns, []table.Column{
			{Title: "Name", Width: 8},
			{Title: "Nickname", Width: 10},
			{Title: "Host", Width: 12},
			{Title: "Port", Width: 4},
			{Title: "User", Width: 10},
			{Title: "Key", Width: 10},
			{Title: "Last login", Width: 15},
		}...)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(int(math.Min(8, float64(len(rows)+1)))),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(false)
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)

	t.SetStyles(s)

	p := tea.NewProgram(model{table: t, what: what})
	m, err := p.Run()
	if err != nil {
		fmt.Println("error while running the interactive selector, ", err)
		os.Exit(1)
	}
	// Assert the final tea.Model to our local model and print the choice.
	if m, ok := m.(model); ok {
		switch m.action {
		case "connect":
			return m.choice, "connect", m.table.Cursor()
		case "edit":
			return config.SSHConfig{}, "edit", m.selectedIndex
		case "quit":
			return config.SSHConfig{}, "quit", -1
		}
		// Default case if action is not set, but exit is true (e.g. unexpected)
		if m.exit {
			return config.SSHConfig{}, "quit", -1
		}
	}
	// Should ideally not be reached if tea.Quit is always paired with an action
	return config.SSHConfig{}, "quit", -1
}
func (m model) HelpView() string {

	km := table.DefaultKeyMap()

	var b strings.Builder

	b.WriteString(generateHelpBlock(km.LineUp.Help().Key, km.LineUp.Help().Desc, true))
	b.WriteString(generateHelpBlock(km.LineDown.Help().Key, km.LineDown.Help().Desc, true))

	if m.what == SelectHistory {
		b.WriteString(generateHelpBlock("d", "delete", true))
		b.WriteString(generateHelpBlock("n", "nickname", true)) // Add help for 'n'
	}

	b.WriteString(generateHelpBlock("q/esc", "quit", false))

	return b.String()
}

func generateHelpBlock(key, desc string, withSep bool) string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	})

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#4A4A4A",
	})

	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#DDDADA",
		Dark:  "#3C3C3C",
	})

	sep := sepStyle.Inline(true).Render(" • ")

	str := keyStyle.Inline(true).Render(key) +
		" " +
		descStyle.Inline(true).Render(desc)

	if withSep {
		str += sep
	}

	return str
}
