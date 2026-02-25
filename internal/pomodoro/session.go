package pomodoro

import "time"

type SessionType string

const (
	Focus      SessionType = "Focus Session"
	ShortBreak SessionType = "Short Break"
	LongBreak  SessionType = "Long Break"
)

// SessionResult represents a completed pomodoro session or break.
type SessionResult struct {
	Name        string      `json:"name,omitempty"`
	SessionType SessionType `json:"sessionType"`
	Duration    int         `json:"durationSeconds"`
	StartedAt   time.Time   `json:"startedAt"`
	CompletedAt time.Time   `json:"completedAt"`
}
