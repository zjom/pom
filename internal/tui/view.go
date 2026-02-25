package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Quitting {
		return ""
	}

	var displayTime time.Duration
	if m.IsPaused || m.IsRenaming {
		displayTime = m.TimeLeft
	} else {
		displayTime = time.Until(m.TargetTime)
	}
	if displayTime < 0 {
		displayTime = 0
	}

	mins := int(displayTime.Minutes())
	secs := int(displayTime.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

	percent := 1.0 - (float64(displayTime) / float64(m.TotalDuration))
	if percent < 0 {
		percent = 0
	} else if percent > 1.0 {
		percent = 1.0
	}

	titleText := "🍅 Pomodoro Timer"
	if m.Cfg.SessionName != "" {
		titleText += fmt.Sprintf(" - %s", m.Cfg.SessionName)
	}

	statusText := string(m.CurrentType)
	if m.IsPaused {
		statusText += " (PAUSED)"
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	timerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("43")).Padding(0, 1)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(2)

	ui := fmt.Sprintf(
		"%s\nStatus: %s\nTime: %s\n\n%s\n\nSessions Completed: %d\n",
		titleStyle.Render(titleText),
		statusStyle.Render(statusText),
		timerStyle.Render(timeStr),
		m.Progress.ViewAs(percent),
		m.SessionsDone,
	)

	if m.IsRenaming {
		ui += fmt.Sprintf("\nRename Session:\n%s\n\n%s",
			m.TextInput.View(),
			helpStyle.Render("(Enter to save, Esc to cancel)"),
		)
	} else {
		if m.ShowHelp {
			helpText := "Shortcuts:\n" +
				"  [space/p] Pause / Resume\n" +
				"  [r]       Rename Session\n" +
				"  [?]       Hide Help\n" +
				"  [q]       Quit"
			ui += helpStyle.Render(helpText)
		} else {
			ui += helpStyle.Render("[?] show help • [q] quit")
		}
	}

	uiBox := lipgloss.NewStyle().Align(lipgloss.Center).Render(ui)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, uiBox)
}
