package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	eventStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("#22d3ee"))

	validStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ade80")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f87171")).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748b"))
)

// Dashboard represents the live TUI dashboard for monitoring telemetry events.
type Dashboard struct {
	events  []EventMsg
	Updates chan EventMsg
}

// NewDashboard initializes a new TUI dashboard model.
func NewDashboard(updates chan EventMsg) Dashboard {
	return Dashboard{
		events:  []EventMsg{},
		Updates: updates,
	}
}

func WaitForUpdates(updates chan EventMsg) tea.Cmd {
	return func() tea.Msg {
		return <-updates
	}
}

func (m Dashboard) Init() tea.Cmd {
	return WaitForUpdates(m.Updates)
}

// EventMsg represents a validation event message received by the dashboard.
type EventMsg struct {
	Name    string
	IsValid bool
	Errors  []string
}

func (m Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EventMsg:
		m.events = append(m.events, msg)
		return m, WaitForUpdates(m.Updates)
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Dashboard) View() string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("EventCanvas: Live Telemetry Dashboard"))
	sb.WriteString("\n")

	if len(m.events) == 0 {
		sb.WriteString(mutedStyle.Render("Waiting for events... (Press q to quit)"))
	} else {
		for _, e := range m.events {
			status := validStyle.Render("✓ VALID")
			if !e.IsValid {
				status = errorStyle.Render(fmt.Sprintf("✗ INVALID (%v)", e.Errors))
			}
			sb.WriteString(fmt.Sprintf("%s %s\n", eventStyle.Render(e.Name), status))
		}
		sb.WriteString("\n" + mutedStyle.Render(fmt.Sprintf("Total events: %d | Press q to quit", len(m.events))))
	}

	return sb.String() + "\n"
}
