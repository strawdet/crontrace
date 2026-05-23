package store

import (
	"database/sql"
	"time"
)

// StreakResult holds consecutive success/failure streak info for a command.
type StreakResult struct {
	Command        string
	CurrentStreak  int
	StreakType     string // "success" or "failure"
	LastRun        time.Time
}

type StreakRepository struct {
	db *sql.DB
}

func NewStreakRepository(db *sql.DB) *StreakRepository {
	return &StreakRepository{db: db}
}

// GetStreak returns the current consecutive streak for a given command.
// It walks recent runs from newest to oldest and counts unbroken same-exit-code runs.
func (r *StreakRepository) GetStreak(command string) (*StreakResult, error) {
	rows, err := r.db.Query(`
		SELECT exit_code, started_at
		FROM job_runs
		WHERE command = ?
		ORDER BY started_at DESC
		LIMIT 100
	`, command)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var streak int
	var streakType string
	var lastRun time.Time
	first := true
	baseCode := -1

	for rows.Next() {
		var exitCode int
		var startedAt time.Time
		if err := rows.Scan(&exitCode, &startedAt); err != nil {
			return nil, err
		}
		if first {
			baseCode = exitCode
			lastRun = startedAt
			if exitCode == 0 {
				streakType = "success"
			} else {
				streakType = "failure"
			}
			first = false
		}
		if exitCode == baseCode {
			streak++
		} else {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if streak == 0 {
		return nil, nil
	}
	return &StreakResult{
		Command:       command,
		CurrentStreak: streak,
		StreakType:    streakType,
		LastRun:       lastRun,
	}, nil
}
