package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

// Config holds our parsed flags.
type Config struct {
	SessionName     string
	SessionDuration time.Duration
	ShortBreak      time.Duration
	LongBreak       time.Duration
	SessionsToLong  int
}

// SessionResult represents the final stats outputted as JSON.
type SessionResult struct {
	Name              string    `json:"name,omitempty"`
	CompletedSessions int       `json:"completedSessions"`
	StartTime         time.Time `json:"startTime"`
	EndTime           time.Time `json:"endTime"`
}

type SessionType string

const (
	TypeSession    SessionType = "Focus Session"
	TypeShortBreak SessionType = "Short Break"
	TypeLongBreak  SessionType = "Long Break"
)

type tickMsg time.Time

type model struct {
	cfg           Config
	currentType   SessionType
	targetTime    time.Time
	startTime     time.Time
	totalDuration time.Duration // Tracks the total time of the current block for the progress bar
	sessionsDone  int
	quitting      bool

	// Dimensions
	width  int
	height int

	// Feature states
	isPaused   bool
	isRenaming bool
	showHelp   bool
	timeLeft   time.Duration
	textInput  textinput.Model
	progress   progress.Model // Our new progress bar component
}

func initialModel(cfg Config) model {
	ti := textinput.New()
	ti.Placeholder = "Enter new session name"
	ti.CharLimit = 50
	ti.Width = 30

	// Initialize the progress bar with a gradient that matches our theme
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40 // Set a fixed comfortable width for the bar

	return model{
		cfg:           cfg,
		currentType:   TypeSession,
		startTime:     time.Now(),
		targetTime:    time.Now().Add(cfg.SessionDuration),
		totalDuration: cfg.SessionDuration,
		sessionsDone:  0,
		textInput:     ti,
		progress:      prog,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), textinput.Blink)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func notifyCmd(title, message string) tea.Cmd {
	return func() tea.Msg {
		err := beeep.Notify(title, message, "")
		if err != nil {
			log.Printf("Failed to send notification: %v", err)
		}
		return nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Ensure the progress bar doesn't overflow small terminal windows
		if m.width < 50 {
			m.progress.Width = m.width - 10
		} else {
			m.progress.Width = 40
		}
		return m, nil

	case tea.KeyMsg:
		if m.isRenaming {
			switch msg.Type {
			case tea.KeyEnter:
				if m.textInput.Value() != "" {
					m.cfg.SessionName = m.textInput.Value()
				}
				m.isRenaming = false
				m.textInput.Blur()
				m.targetTime = time.Now().Add(m.timeLeft)
				return m, nil

			case tea.KeyEsc:
				m.isRenaming = false
				m.textInput.Blur()
				m.targetTime = time.Now().Add(m.timeLeft)
				return m, nil
			}

			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case " ", "p":
			m.isPaused = !m.isPaused
			if m.isPaused {
				m.timeLeft = time.Until(m.targetTime)
			} else {
				m.targetTime = time.Now().Add(m.timeLeft)
			}

		case "?":
			m.showHelp = !m.showHelp

		case "r":
			m.isRenaming = true
			m.timeLeft = time.Until(m.targetTime)
			m.textInput.SetValue(m.cfg.SessionName)
			m.textInput.Focus()
			return m, textinput.Blink
		}

	case tickMsg:
		if !m.isPaused && !m.isRenaming {
			if time.Now().After(m.targetTime) {
				var notify tea.Cmd
				m, notify = m.nextState()
				return m, tea.Batch(tickCmd(), notify)
			}
		}
		return m, tickCmd()
	}

	return m, nil
}

func (m model) nextState() (model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentType {
	case TypeSession:
		m.sessionsDone++
		if m.sessionsDone%m.cfg.SessionsToLong == 0 {
			m.currentType = TypeLongBreak
			m.totalDuration = m.cfg.LongBreak
			m.targetTime = time.Now().Add(m.cfg.LongBreak)
			cmd = notifyCmd("Pomodoro", "Focus session complete! Time for a long break.")
		} else {
			m.currentType = TypeShortBreak
			m.totalDuration = m.cfg.ShortBreak
			m.targetTime = time.Now().Add(m.cfg.ShortBreak)
			cmd = notifyCmd("Pomodoro", "Focus session complete! Take a quick breather.")
		}
	case TypeShortBreak, TypeLongBreak:
		m.currentType = TypeSession
		m.totalDuration = m.cfg.SessionDuration
		m.targetTime = time.Now().Add(m.cfg.SessionDuration)
		cmd = notifyCmd("Pomodoro", "Break is over. Time to get back to focus!")
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var displayTime time.Duration
	if m.isPaused || m.isRenaming {
		displayTime = m.timeLeft
	} else {
		displayTime = time.Until(m.targetTime)
	}

	if displayTime < 0 {
		displayTime = 0
	}

	mins := int(displayTime.Minutes())
	secs := int(displayTime.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

	// Calculate Progress Percentage
	percent := 1.0 - (float64(displayTime) / float64(m.totalDuration))
	if percent < 0 {
		percent = 0
	} else if percent > 1.0 {
		percent = 1.0
	}

	titleText := "🍅 Pomodoro Timer"
	if m.cfg.SessionName != "" {
		titleText += fmt.Sprintf(" - %s", m.cfg.SessionName)
	}

	statusText := string(m.currentType)
	if m.isPaused {
		statusText += " (PAUSED)"
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	timerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("43")).Padding(0, 1)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(2)

	// Build the main UI block
	ui := fmt.Sprintf(
		"%s\nStatus: %s\nTime: %s\n\n%s\n\nSessions Completed: %d\n",
		titleStyle.Render(titleText),
		statusStyle.Render(statusText),
		timerStyle.Render(timeStr),
		m.progress.ViewAs(percent), // Inject the progress bar here
		m.sessionsDone,
	)

	// Append contextual menus
	if m.isRenaming {
		ui += fmt.Sprintf("\nRename Session:\n%s\n\n%s",
			m.textInput.View(),
			helpStyle.Render("(Enter to save, Esc to cancel)"),
		)
	} else {
		if m.showHelp {
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

func main() {
	nameFlag := flag.String("name", "", "optional name for the session")
	sessionFlag := flag.String("session", "25m", "duration of session")
	sbreakFlag := flag.String("sbreak", "5m", "duration of short break")
	lbreakFlag := flag.String("lbreak", "15m", "duration of long break")
	nbreakFlag := flag.Int("nbreak", 4, "n sessions before long break")

	flag.Parse()

	sessionDuration, err := time.ParseDuration(*sessionFlag)
	if err != nil {
		log.Fatalf("Invalid session duration: %v", err)
	}
	shortBreakDuration, err := time.ParseDuration(*sbreakFlag)
	if err != nil {
		log.Fatalf("Invalid short break duration: %v", err)
	}
	longBreakDuration, err := time.ParseDuration(*lbreakFlag)
	if err != nil {
		log.Fatalf("Invalid long break duration: %v", err)
	}

	cfg := Config{
		SessionName:     *nameFlag,
		SessionDuration: sessionDuration,
		ShortBreak:      shortBreakDuration,
		LongBreak:       longBreakDuration,
		SessionsToLong:  *nbreakFlag,
	}

	p := tea.NewProgram(initialModel(cfg), tea.WithAltScreen())

	finalState, err := p.Run()
	if err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}

	if m, ok := finalState.(model); ok {
		result := SessionResult{
			Name:              m.cfg.SessionName,
			CompletedSessions: m.sessionsDone,
			StartTime:         m.startTime,
			EndTime:           time.Now(),
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}

		fmt.Println(string(jsonData))
	}
}
