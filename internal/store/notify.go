package store

import (
	"database/sql"
	"time"
)

// NotifyRule represents a threshold-based notification rule for a cron job.
type NotifyRule struct {
	ID            int64
	Command       string
	MaxDurationMs int64
	AlertOnFail   bool
	CreatedAt     time.Time
}

// NotifyRepository handles persistence of notification rules.
type NotifyRepository struct {
	db *sql.DB
}

// NewNotifyRepository creates a new NotifyRepository and ensures the table exists.
func NewNotifyRepository(db *sql.DB) (*NotifyRepository, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS notify_rules (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		command        TEXT NOT NULL UNIQUE,
		max_duration_ms INTEGER NOT NULL DEFAULT 0,
		alert_on_fail  INTEGER NOT NULL DEFAULT 1,
		created_at     DATETIME NOT NULL
	)`)
	if err != nil {
		return nil, err
	}
	return &NotifyRepository{db: db}, nil
}

// Upsert inserts or replaces a notification rule for the given command.
func (r *NotifyRepository) Upsert(rule NotifyRule) error {
	_, err := r.db.Exec(
		`INSERT INTO notify_rules (command, max_duration_ms, alert_on_fail, created_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(command) DO UPDATE SET
		   max_duration_ms = excluded.max_duration_ms,
		   alert_on_fail   = excluded.alert_on_fail`,
		rule.Command, rule.MaxDurationMs, rule.AlertOnFail, time.Now().UTC(),
	)
	return err
}

// Get returns the notification rule for a command, or nil if none exists.
func (r *NotifyRepository) Get(command string) (*NotifyRule, error) {
	row := r.db.QueryRow(
		`SELECT id, command, max_duration_ms, alert_on_fail, created_at FROM notify_rules WHERE command = ?`,
		command,
	)
	var rule NotifyRule
	var createdAt string
	err := row.Scan(&rule.ID, &rule.Command, &rule.MaxDurationMs, &rule.AlertOnFail, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rule.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &rule, nil
}

// Delete removes the notification rule for the given command.
func (r *NotifyRepository) Delete(command string) error {
	_, err := r.db.Exec(`DELETE FROM notify_rules WHERE command = ?`, command)
	return err
}

// List returns all notification rules.
func (r *NotifyRepository) List() ([]NotifyRule, error) {
	rows, err := r.db.Query(`SELECT id, command, max_duration_ms, alert_on_fail, created_at FROM notify_rules ORDER BY command`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []NotifyRule
	for rows.Next() {
		var rule NotifyRule
		var createdAt string
		if err := rows.Scan(&rule.ID, &rule.Command, &rule.MaxDurationMs, &rule.AlertOnFail, &createdAt); err != nil {
			return nil, err
		}
		rule.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}
