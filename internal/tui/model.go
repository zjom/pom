package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/zjom/pom/internal/config"
	"github.com/zjom/pom/internal/pomodoro"
	"github.com/zjom/pom/internal/storage"
)

// Model holds all TUI state for the pomodoro timer.
type Model struct {
	Cfg           config.Config
	Store         storage.Store
	CurrentType   pomodoro.SessionType
	TargetTime    time.Time
	StartTime     time.Time
	SessionStart  time.Time // start of the current individual session/break
	TotalDuration time.Duration
	SessionsDone  int
	Quitting      bool

	width  int
	height int

	IsPaused   bool
	IsRenaming bool
	ShowHelp   bool
	TimeLeft   time.Duration
	TextInput  textinput.Model
	Progress   progress.Model
}

func NewModel(cfg config.Config, store storage.Store) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter new session name"
	ti.CharLimit = 50
	ti.Width = 30

	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40

	now := time.Now()
	return Model{
		Cfg:           cfg,
		Store:         store,
		CurrentType:   pomodoro.Focus,
		StartTime:     now,
		SessionStart:  now,
		TargetTime:    now.Add(cfg.SessionDuration),
		TotalDuration: cfg.SessionDuration,
		TextInput:     ti,
		Progress:      prog,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), textinput.Blink)
}
