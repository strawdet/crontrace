package store

import (
	"database/sql"
	"time"
)

// JobRun represents a single execution record of a cron job.
type JobRun struct {
	ID        int64
	Command   string
	StartedAt time.Time
	EndedAt   *time.Time
	ExitCode  int
	Output    string
}

// JobRunRepository provides persistence for JobRun records.
type JobRunRepository struct {
	db *sql.DB
}

// NewJobRunRepository creates a new repository backed by db.
func NewJobRunRepository(db *sql.DB) *JobRunRepository {
	return &JobRunRepository{db: db}
}

// Insert persists a new JobRun and sets its ID.
func (r *JobRunRepository) Insert(run *JobRun) error {
	const q = `INSERT INTO job_runs (command, started_at, ended_at, exit_code, output)
		VALUES (?, ?, ?, ?, ?)`
	res, err := r.db.Exec(q,
		run.Command,
		run.StartedAt.UTC().Format(time.RFC3339Nano),
		nullTime(run.EndedAt),
		run.ExitCode,
		run.Output,
	)
	if err != nil {
		return err
	}
	run.ID, err = res.LastInsertId()
	return err
}

// List returns the most recent limit job runs, newest first.
func (r *JobRunRepository) List(limit int) ([]JobRun, error) {
	const q = `SELECT id, command, started_at, ended_at, exit_code, output
		FROM job_runs ORDER BY started_at DESC LIMIT ?`
	rows, err := r.db.Query(q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []JobRun
	for rows.Next() {
		var run JobRun
		var startedStr string
		var endedStr sql.NullString
		if err := rows.Scan(&run.ID, &run.Command, &startedStr, &endedStr, &run.ExitCode, &run.Output); err != nil {
			return nil, err
		}
		run.StartedAt, _ = time.Parse(time.RFC3339Nano, startedStr)
		if endedStr.Valid {
			t, _ := time.Parse(time.RFC3339Nano, endedStr.String)
			run.EndedAt = &t
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}
