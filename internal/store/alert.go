package store

import (
	"database/sql"
	"time"
)

// Alert represents a threshold-based alert rule for a cron job command.
type Alert struct {
	ID          int64
	Command     string
	Metric      string // "duration", "exit_code", "failure_rate"
	Threshold   float64
	CreatedAt   time.Time
}

// AlertRepository handles persistence of alert rules.
type AlertRepository struct {
	db *sql.DB
}

// NewAlertRepository creates a new AlertRepository and ensures the schema exists.
func NewAlertRepository(db *sql.DB) (*AlertRepository, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS alerts (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		command    TEXT NOT NULL,
		metric     TEXT NOT NULL,
		threshold  REAL NOT NULL,
		created_at DATETIME NOT NULL,
		UNIQUE(command, metric)
	)`)
	if err != nil {
		return nil, err
	}
	return &AlertRepository{db: db}, nil
}

// Upsert inserts or replaces an alert rule for the given command and metric.
func (r *AlertRepository) Upsert(command, metric string, threshold float64) error {
	_, err := r.db.Exec(
		`INSERT INTO alerts (command, metric, threshold, created_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(command, metric) DO UPDATE SET threshold=excluded.threshold, created_at=excluded.created_at`,
		command, metric, threshold, time.Now().UTC(),
	)
	return err
}

// List returns all alert rules, optionally filtered by command.
func (r *AlertRepository) List(command string) ([]Alert, error) {
	query := `SELECT id, command, metric, threshold, created_at FROM alerts`
	args := []interface{}{}
	if command != "" {
		query += ` WHERE command = ?`
		args = append(args, command)
	}
	query += ` ORDER BY command, metric`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var alerts []Alert
	for rows.Next() {
		var a Alert
		if err := rows.Scan(&a.ID, &a.Command, &a.Metric, &a.Threshold, &a.CreatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// Delete removes an alert rule by command and metric.
func (r *AlertRepository) Delete(command, metric string) (int64, error) {
	res, err := r.db.Exec(`DELETE FROM alerts WHERE command = ? AND metric = ?`, command, metric)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
