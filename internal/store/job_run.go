package store

import (
	"database/sql"
	"fmt"
	"time"
)

// JobRun represents a single recorded execution of a cron job.
type JobRun struct {
	ID         int64
	Command    string
	StartedAt  time.Time
	FinishedAt *time.Time
	DurationMs *int64
	ExitCode   *int
	Stdout     string
	Stderr     string
}

// JobRunRepository handles persistence of JobRun records.
type JobRunRepository struct {
	db *sql.DB
}

// NewJobRunRepository creates a new repository backed by db.
func NewJobRunRepository(db *DB) *JobRunRepository {
	return &JobRunRepository{db: db.Conn()}
}

// Insert persists a new JobRun and sets its ID.
func (r *JobRunRepository) Insert(run *JobRun) error {
	const q = `
	INSERT INTO job_runs (command, started_at, finished_at, duration_ms, exit_code, stdout, stderr)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(q,
		run.Command,
		run.StartedAt.UTC(),
		nullTime(run.FinishedAt),
		run.DurationMs,
		run.ExitCode,
		run.Stdout,
		run.Stderr,
	)
	if err != nil {
		return fmt.Errorf("inserting job run: %w", err)
	}
	run.ID, _ = res.LastInsertId()
	return nil
}

// ListByCommand returns the most recent `limit` runs for the given command.
func (r *JobRunRepository) ListByCommand(command string, limit int) ([]JobRun, error) {
	const q = `
	SELECT id, command, started_at, finished_at, duration_ms, exit_code, stdout, stderr
	FROM job_runs WHERE command = ?
	ORDER BY started_at DESC LIMIT ?`

	rows, err := r.db.Query(q, command, limit)
	if err != nil {
		return nil, fmt.Errorf("querying job runs: %w", err)
	}
	defer rows.Close()

	var runs []JobRun
	for rows.Next() {
		var run JobRun
		var finishedAt sql.NullTime
		if err := rows.Scan(
			&run.ID, &run.Command, &run.StartedAt, &finishedAt,
			&run.DurationMs, &run.ExitCode, &run.Stdout, &run.Stderr,
		); err != nil {
			return nil, fmt.Errorf("scanning job run: %w", err)
		}
		if finishedAt.Valid {
			run.FinishedAt = &finishedAt.Time
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.UTC()
}
