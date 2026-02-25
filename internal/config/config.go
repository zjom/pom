package config

import "time"

type Config struct {
	SessionName     string
	SessionDuration time.Duration
	ShortBreak      time.Duration
	LongBreak       time.Duration
	SessionsToLong  int
}

func Default() Config {
	return Config{
		SessionDuration: 25 * time.Minute,
		ShortBreak:      5 * time.Minute,
		LongBreak:       15 * time.Minute,
		SessionsToLong:  4,
	}
}
