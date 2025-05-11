package migration

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding  = 2
	maxWidth = 80
)

type errMsg error

var infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render

var successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render

func handleQuitKeyPress(msg tea.KeyMsg, m tea.Model) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyCtrlQ:
		return m, tea.Quit
	default:
		return m, nil
	}
}
