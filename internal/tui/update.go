package tui

import (
	"context"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/zjom/pom/internal/pomodoro"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.width < 50 {
			m.Progress.Width = m.width - 10
		} else {
			m.Progress.Width = 40
		}
		return m, nil

	case tea.KeyMsg:
		if m.IsRenaming {
			switch msg.Type {
			case tea.KeyEnter:
				if m.TextInput.Value() != "" {
					m.Cfg.SessionName = m.TextInput.Value()
				}
				m.IsRenaming = false
				m.TextInput.Blur()
				m.TargetTime = time.Now().Add(m.TimeLeft)
				return m, nil
			case tea.KeyEsc:
				m.IsRenaming = false
				m.TextInput.Blur()
				m.TargetTime = time.Now().Add(m.TimeLeft)
				return m, nil
			}
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit
		case " ", "p":
			m.IsPaused = !m.IsPaused
			if m.IsPaused {
				m.TimeLeft = time.Until(m.TargetTime)
			} else {
				m.TargetTime = time.Now().Add(m.TimeLeft)
			}
		case "?":
			m.ShowHelp = !m.ShowHelp
		case "r":
			m.IsRenaming = true
			m.TimeLeft = time.Until(m.TargetTime)
			m.TextInput.SetValue(m.Cfg.SessionName)
			m.TextInput.Focus()
			return m, textinput.Blink
		}

	case tickMsg:
		if !m.IsPaused && !m.IsRenaming {
			if time.Now().After(m.TargetTime) {
				var notify tea.Cmd
				m, notify = m.nextState()
				return m, tea.Batch(tickCmd(), notify)
			}
		}
		return m, tickCmd()
	}

	return m, nil
}

func (m Model) nextState() (Model, tea.Cmd) {
	now := time.Now()

	// Persist the completed session/break.
	if m.Store != nil {
		sr := pomodoro.SessionResult{
			Name:        m.Cfg.SessionName,
			SessionType: m.CurrentType,
			Duration:    int(m.TotalDuration.Seconds()),
			StartedAt:   m.SessionStart,
			CompletedAt: now,
		}
		if err := m.Store.SaveSession(context.Background(), sr); err != nil {
			log.Printf("Failed to save session: %v", err)
		}
	}

	var cmd tea.Cmd
	nextType, nextDur, done := pomodoro.NextSession(m.CurrentType, m.SessionsDone, m.Cfg)
	m.SessionsDone = done
	m.CurrentType = nextType
	m.TotalDuration = nextDur
	m.TargetTime = now.Add(nextDur)
	m.SessionStart = now

	switch nextType {
	case pomodoro.ShortBreak:
		cmd = notifyCmd("Pomodoro", "Focus session complete! Take a quick breather.")
	case pomodoro.LongBreak:
		cmd = notifyCmd("Pomodoro", "Focus session complete! Time for a long break.")
	case pomodoro.Focus:
		cmd = notifyCmd("Pomodoro", "Break is over. Time to get back to focus!")
	}

	return m, cmd
}
