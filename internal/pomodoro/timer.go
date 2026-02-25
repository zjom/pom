package pomodoro

import (
	"time"

	"github.com/zjom/pom/internal/config"
)

// NextSession returns the next session type and its duration given the current
// state. It also returns the updated count of completed focus sessions.
func NextSession(current SessionType, sessionsDone int, cfg config.Config) (SessionType, time.Duration, int) {
	switch current {
	case Focus:
		sessionsDone++
		if sessionsDone%cfg.SessionsToLong == 0 {
			return LongBreak, cfg.LongBreak, sessionsDone
		}
		return ShortBreak, cfg.ShortBreak, sessionsDone
	default: // ShortBreak, LongBreak
		return Focus, cfg.SessionDuration, sessionsDone
	}
}
