package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Schedule represents a registered cron schedule for a command.
type Schedule struct {
	ID        int64
	Command   string
	CronExpr  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ScheduleRepository handles persistence of cron schedules.
type ScheduleRepository struct {
	db *sql.DB
}

// NewScheduleRepository creates a new ScheduleRepository and ensures the table exists.
func NewScheduleRepository(db *sql.DB) (*ScheduleRepository, error) {
	const ddl = `CREATE TABLE IF NOT EXISTS schedules (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		command    TEXT NOT NULL UNIQUE,
		cron_expr  TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(ddl); err != nil {
		return nil, fmt.Errorf("schedule: create table: %w", err)
	}
	return &ScheduleRepository{db: db}, nil
}

// Upsert inserts or updates the cron expression for a command.
func (r *ScheduleRepository) Upsert(command, cronExpr string) error {
	const q = `INSERT INTO schedules (command, cron_expr, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(command) DO UPDATE SET cron_expr=excluded.cron_expr, updated_at=excluded.updated_at;`
	_, err := r.db.Exec(q, command, cronExpr, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("schedule: upsert: %w", err)
	}
	return nil
}

// Get returns the Schedule for the given command, or an error if not found.
func (r *ScheduleRepository) Get(command string) (*Schedule, error) {
	const q = `SELECT id, command, cron_expr, created_at, updated_at FROM schedules WHERE command = ?;`
	row := r.db.QueryRow(q, command)
	var s Schedule
	if err := row.Scan(&s.ID, &s.Command, &s.CronExpr, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schedule: not found: %s", command)
		}
		return nil, fmt.Errorf("schedule: get: %w", err)
	}
	return &s, nil
}

// List returns all registered schedules.
func (r *ScheduleRepository) List() ([]Schedule, error) {
	const q = `SELECT id, command, cron_expr, created_at, updated_at FROM schedules ORDER BY command;`
	rows, err := r.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("schedule: list: %w", err)
	}
	defer rows.Close()
	var out []Schedule
	for rows.Next() {
		var s Schedule
		if err := rows.Scan(&s.ID, &s.Command, &s.CronExpr, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("schedule: scan: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// Delete removes the schedule for the given command.
func (r *ScheduleRepository) Delete(command string) error {
	const q = `DELETE FROM schedules WHERE command = ?;`
	_, err := r.db.Exec(q, command)
	if err != nil {
		return fmt.Errorf("schedule: delete: %w", err)
	}
	return nil
}
