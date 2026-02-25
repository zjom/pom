package storage

import (
	"time"

	"github.com/zjom/pom/internal/pomodoro"
)

// QueryFilter constrains which sessions are returned by List or Statistics queries.
type QueryFilter struct {
	Name        string
	SessionType *pomodoro.SessionType
	From        *time.Time
	To          *time.Time
	Limit       int
}

// Statistics holds aggregated session data.
type Statistics struct {
	TotalSessions   int            `json:"totalSessions"`
	TotalTime       time.Duration  `json:"totalTime"`
	AverageDuration time.Duration  `json:"averageDuration"`
	ByType          map[string]int `json:"byType"`
}
