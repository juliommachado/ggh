package theme

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type Print int

const (
	PrintConfig = iota
	PrintHistory
)

func PrintTable(rows []table.Row, p Print) string {
	var columns []table.Column
	if p == PrintHistory {
		columns = []table.Column{
			{Title: "Name", Width: 8},
			{Title: "Nickname", Width: 10},
			{Title: "Host", Width: 12},
			{Title: "Port", Width: 4},
			{Title: "User", Width: 10},
			{Title: "Key", Width: 10},
			{Title: "Last login", Width: 15},
		}
	} else { // For PrintConfig
		columns = []table.Column{
			{Title: "Name", Width: 10},
			{Title: "Host", Width: 15},
			{Title: "Port", Width: 10},
			{Title: "User", Width: 10},
			{Title: "Key", Width: 10},
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithStyles(table.Styles{
			Header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
			Selected: lipgloss.NewStyle(),
		}),
		table.WithHeight(len(rows)+1),
	)

	return BaseStyle.Render(t.View())
}
