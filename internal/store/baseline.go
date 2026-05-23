package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Baseline holds the expected duration baseline for a command.
type Baseline struct {
	Command       string
	AvgDurationMs int64
	SampleCount   int
	UpdatedAt     time.Time
}

// BaselineRepository manages duration baselines per command.
type BaselineRepository struct {
	db *sql.DB
}

// NewBaselineRepository creates the baselines table if needed and returns a repository.
func NewBaselineRepository(db *sql.DB) (*BaselineRepository, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS baselines (
		command        TEXT PRIMARY KEY,
		avg_duration_ms INTEGER NOT NULL DEFAULT 0,
		sample_count    INTEGER NOT NULL DEFAULT 0,
		updated_at      DATETIME NOT NULL
	)`)
	if err != nil {
		return nil, fmt.Errorf("create baselines table: %w", err)
	}
	return &BaselineRepository{db: db}, nil
}

// Upsert inserts or replaces the baseline for a command.
func (r *BaselineRepository) Upsert(b Baseline) error {
	_, err := r.db.Exec(
		`INSERT INTO baselines (command, avg_duration_ms, sample_count, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(command) DO UPDATE SET
			avg_duration_ms = excluded.avg_duration_ms,
			sample_count    = excluded.sample_count,
			updated_at      = excluded.updated_at`,
		b.Command, b.AvgDurationMs, b.SampleCount, b.UpdatedAt.UTC(),
	)
	return err
}

// Get returns the baseline for a command, or nil if not found.
func (r *BaselineRepository) Get(command string) (*Baseline, error) {
	row := r.db.QueryRow(
		`SELECT command, avg_duration_ms, sample_count, updated_at FROM baselines WHERE command = ?`,
		command,
	)
	var b Baseline
	var updatedAt string
	if err := row.Scan(&b.Command, &b.AvgDurationMs, &b.SampleCount, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t, err := time.Parse("2006-01-02T15:04:05Z", updatedAt)
	if err != nil {
		t, _ = time.Parse("2006-01-02 15:04:05+00:00", updatedAt)
	}
	b.UpdatedAt = t
	return &b, nil
}

// List returns all stored baselines.
func (r *BaselineRepository) List() ([]Baseline, error) {
	rows, err := r.db.Query(`SELECT command, avg_duration_ms, sample_count, updated_at FROM baselines ORDER BY command`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Baseline
	for rows.Next() {
		var b Baseline
		var updatedAt string
		if err := rows.Scan(&b.Command, &b.AvgDurationMs, &b.SampleCount, &updatedAt); err != nil {
			return nil, err
		}
		t, err := time.Parse("2006-01-02T15:04:05Z", updatedAt)
		if err != nil {
			t, _ = time.Parse("2006-01-02 15:04:05+00:00", updatedAt)
		}
		b.UpdatedAt = t
		out = append(out, b)
	}
	return out, rows.Err()
}

// Delete removes the baseline for a command.
func (r *BaselineRepository) Delete(command string) error {
	_, err := r.db.Exec(`DELETE FROM baselines WHERE command = ?`, command)
	return err
}
