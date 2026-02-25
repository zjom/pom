package tui

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/beeep"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func notifyCmd(title, message string) tea.Cmd {
	return func() tea.Msg {
		if err := beeep.Notify(title, message, ""); err != nil {
			log.Printf("Failed to send notification: %v", err)
		}
		return nil
	}
}
