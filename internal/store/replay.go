package store

import (
	"database/sql"
	"fmt"
	"time"
)

// ReplayEntry represents a stored command that can be re-executed.
type ReplayEntry struct {
	ID        int64
	Command   string
	Args      string
	LastRunAt time.Time
	ExitCode  int
}

// ReplayRepository provides access to replay-related queries.
type ReplayRepository struct {
	db *sql.DB
}

// NewReplayRepository creates a new ReplayRepository.
func NewReplayRepository(db *sql.DB) *ReplayRepository {
	return &ReplayRepository{db: db}
}

// GetReplayByID fetches a job run by ID and returns a ReplayEntry.
func (r *ReplayRepository) GetReplayByID(id int64) (*ReplayEntry, error) {
	row := r.db.QueryRow(`
		SELECT id, command, args, started_at, exit_code
		FROM job_runs
		WHERE id = ?`, id)

	var entry ReplayEntry
	var startedAt string
	err := row.Scan(&entry.ID, &entry.Command, &entry.Args, &startedAt, &entry.ExitCode)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no run found with id %d", id)
	}
	if err != nil {
		return nil, err
	}

	entry.LastRunAt, err = time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return nil, fmt.Errorf("parse time: %w", err)
	}

	return &entry, nil
}

// ListReplayable returns the most recent unique commands available for replay.
func (r *ReplayRepository) ListReplayable(limit int) ([]ReplayEntry, error) {
	rows, err := r.db.Query(`
		SELECT id, command, args, started_at, exit_code
		FROM job_runs
		WHERE id IN (
			SELECT MAX(id) FROM job_runs GROUP BY command
		)
		ORDER BY started_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ReplayEntry
	for rows.Next() {
		var e ReplayEntry
		var startedAt string
		if err := rows.Scan(&e.ID, &e.Command, &e.Args, &startedAt, &e.ExitCode); err != nil {
			return nil, err
		}
		e.LastRunAt, _ = time.Parse(time.RFC3339, startedAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
