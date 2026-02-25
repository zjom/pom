package storage

import (
	"context"

	"github.com/zjom/pom/internal/pomodoro"
)

// Store defines the interface for session persistence.
type Store interface {
	SaveSession(ctx context.Context, s pomodoro.SessionResult) error
	ListSessions(ctx context.Context, f QueryFilter) ([]pomodoro.SessionResult, error)
	GetStatistics(ctx context.Context, f QueryFilter) (*Statistics, error)
	Close() error
}
