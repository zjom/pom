package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/zjom/pom/internal/config"
	"github.com/zjom/pom/internal/storage"
	"github.com/zjom/pom/internal/tui"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a pomodoro session",
	RunE:  runStart,
}

var (
	flagName    string
	flagSession string
	flagSBreak  string
	flagLBreak  string
	flagNBreak  int
)

func init() {
	startCmd.Flags().StringVarP(&flagName, "name", "n", "", "optional session label")
	startCmd.Flags().StringVarP(&flagSession, "session", "s", "25m", "focus duration")
	startCmd.Flags().StringVar(&flagSBreak, "sbreak", "5m", "short break duration")
	startCmd.Flags().StringVar(&flagLBreak, "lbreak", "15m", "long break duration")
	startCmd.Flags().IntVar(&flagNBreak, "nbreak", 4, "sessions before a long break")

	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	sessionDur, err := time.ParseDuration(flagSession)
	if err != nil {
		return fmt.Errorf("invalid session duration: %w", err)
	}
	shortDur, err := time.ParseDuration(flagSBreak)
	if err != nil {
		return fmt.Errorf("invalid short break duration: %w", err)
	}
	longDur, err := time.ParseDuration(flagLBreak)
	if err != nil {
		return fmt.Errorf("invalid long break duration: %w", err)
	}

	cfg := config.Config{
		SessionName:     flagName,
		SessionDuration: sessionDur,
		ShortBreak:      shortDur,
		LongBreak:       longDur,
		SessionsToLong:  flagNBreak,
	}

	dbPath, err := storage.DefaultDBPath()
	if err != nil {
		return fmt.Errorf("determine database path: %w", err)
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer store.Close()

	m := tui.NewModel(cfg, store)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalState, err := p.Run()
	if err != nil {
		return fmt.Errorf("run timer: %w", err)
	}

	if fm, ok := finalState.(tui.Model); ok {
		result := struct {
			Name              string    `json:"name,omitempty"`
			CompletedSessions int       `json:"completedSessions"`
			StartTime         time.Time `json:"startTime"`
			EndTime           time.Time `json:"endTime"`
		}{
			Name:              fm.Cfg.SessionName,
			CompletedSessions: fm.SessionsDone,
			StartTime:         fm.StartTime,
			EndTime:           time.Now(),
		}
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal JSON: %v", err)
		} else {
			fmt.Println(string(data))
		}
	}

	return nil
}
