package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/zjom/pom/internal/pomodoro"
)

const schema = `
CREATE TABLE IF NOT EXISTS sessions (
	id               INTEGER PRIMARY KEY AUTOINCREMENT,
	name             TEXT,
	session_type     TEXT NOT NULL,
	duration_seconds INTEGER NOT NULL,
	started_at       DATETIME NOT NULL,
	completed_at     DATETIME NOT NULL
);
`

type SQLiteStore struct {
	db *sql.DB
}

// DefaultDBPath returns ~/.local/share/pom/history.db.
func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local", "share", "pom")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.db"), nil
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) SaveSession(ctx context.Context, sr pomodoro.SessionResult) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (name, session_type, duration_seconds, started_at, completed_at)
		 VALUES (?, ?, ?, ?, ?)`,
		sr.Name, string(sr.SessionType), sr.Duration, sr.StartedAt, sr.CompletedAt,
	)
	return err
}

func (s *SQLiteStore) ListSessions(ctx context.Context, f QueryFilter) ([]pomodoro.SessionResult, error) {
	query := `SELECT name, session_type, duration_seconds, started_at, completed_at FROM sessions`
	where, args := buildWhere(f)
	if where != "" {
		query += " WHERE " + where
	}
	query += " ORDER BY started_at DESC"
	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", f.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []pomodoro.SessionResult
	for rows.Next() {
		var r pomodoro.SessionResult
		var st string
		if err := rows.Scan(&r.Name, &st, &r.Duration, &r.StartedAt, &r.CompletedAt); err != nil {
			return nil, err
		}
		r.SessionType = pomodoro.SessionType(st)
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *SQLiteStore) GetStatistics(ctx context.Context, f QueryFilter) (*Statistics, error) {
	query := `SELECT session_type, COUNT(*), COALESCE(SUM(duration_seconds), 0) FROM sessions`
	where, args := buildWhere(f)
	if where != "" {
		query += " WHERE " + where
	}
	query += " GROUP BY session_type"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := &Statistics{ByType: make(map[string]int)}
	var totalSeconds int64
	for rows.Next() {
		var st string
		var count int
		var seconds int64
		if err := rows.Scan(&st, &count, &seconds); err != nil {
			return nil, err
		}
		stats.ByType[st] = count
		stats.TotalSessions += count
		totalSeconds += seconds
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	stats.TotalTime = time.Duration(totalSeconds) * time.Second
	if stats.TotalSessions > 0 {
		stats.AverageDuration = stats.TotalTime / time.Duration(stats.TotalSessions)
	}
	return stats, nil
}

func buildWhere(f QueryFilter) (string, []any) {
	var clauses []string
	var args []any

	if f.Name != "" {
		clauses = append(clauses, "name = ?")
		args = append(args, f.Name)
	}
	if f.SessionType != nil {
		clauses = append(clauses, "session_type = ?")
		args = append(args, string(*f.SessionType))
	}
	if f.From != nil {
		clauses = append(clauses, "started_at >= ?")
		args = append(args, *f.From)
	}
	if f.To != nil {
		clauses = append(clauses, "completed_at <= ?")
		args = append(args, *f.To)
	}

	return strings.Join(clauses, " AND "), args
}
