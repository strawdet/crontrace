package store

import (
	"database/sql"
	"time"
)

// WatchEntry represents a watched command with alert thresholds.
type WatchEntry struct {
	ID             int64
	Command        string
	MaxDurationSec int
	ExpectExitCode int
	CreatedAt      time.Time
}

// WatchRepository manages watched commands.
type WatchRepository struct {
	db *sql.DB
}

// NewWatchRepository creates a new WatchRepository and ensures the table exists.
func NewWatchRepository(db *sql.DB) (*WatchRepository, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS watches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL UNIQUE,
		max_duration_sec INTEGER NOT NULL DEFAULT 0,
		expect_exit_code INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL
	)`)
	if err != nil {
		return nil, err
	}
	return &WatchRepository{db: db}, nil
}

// Upsert inserts or replaces a watch entry for the given command.
func (r *WatchRepository) Upsert(command string, maxDurationSec, expectExitCode int) error {
	_, err := r.db.Exec(
		`INSERT INTO watches (command, max_duration_sec, expect_exit_code, created_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(command) DO UPDATE SET
		   max_duration_sec=excluded.max_duration_sec,
		   expect_exit_code=excluded.expect_exit_code`,
		command, maxDurationSec, expectExitCode, time.Now().UTC(),
	)
	return err
}

// Delete removes a watch entry by command.
func (r *WatchRepository) Delete(command string) error {
	_, err := r.db.Exec(`DELETE FROM watches WHERE command = ?`, command)
	return err
}

// List returns all watch entries.
func (r *WatchRepository) List() ([]WatchEntry, error) {
	rows, err := r.db.Query(`SELECT id, command, max_duration_sec, expect_exit_code, created_at FROM watches ORDER BY command`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []WatchEntry
	for rows.Next() {
		var e WatchEntry
		if err := rows.Scan(&e.ID, &e.Command, &e.MaxDurationSec, &e.ExpectExitCode, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// Get returns a single watch entry by command, or nil if not found.
func (r *WatchRepository) Get(command string) (*WatchEntry, error) {
	row := r.db.QueryRow(`SELECT id, command, max_duration_sec, expect_exit_code, created_at FROM watches WHERE command = ?`, command)
	var e WatchEntry
	if err := row.Scan(&e.ID, &e.Command, &e.MaxDurationSec, &e.ExpectExitCode, &e.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}
