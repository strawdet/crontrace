package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS job_runs (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	command     TEXT    NOT NULL,
	started_at  DATETIME NOT NULL,
	finished_at DATETIME,
	duration_ms INTEGER,
	exit_code   INTEGER,
	stdout      TEXT,
	stderr      TEXT
);

CREATE INDEX IF NOT EXISTS idx_job_runs_command ON job_runs(command);
CREATE INDEX IF NOT EXISTS idx_job_runs_started_at ON job_runs(started_at);
`

// DB wraps a SQLite database connection for crontrace.
type DB struct {
	conn *sql.DB
}

// Open opens (or creates) the SQLite database at the given path.
func Open(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("opening sqlite db: %w", err)
	}

	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("applying schema: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the raw *sql.DB for use by repository types.
func (db *DB) Conn() *sql.DB {
	return db.conn
}
