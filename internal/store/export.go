package store

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ExportRepository handles exporting job run data.
type ExportRepository struct {
	db *sql.DB
}

// NewExportRepository creates a new ExportRepository.
func NewExportRepository(db *sql.DB) *ExportRepository {
	return &ExportRepository{db: db}
}

// ExportRow represents a single row for export.
type ExportRow struct {
	ID        int64     `json:"id"`
	Command   string    `json:"command"`
	StartedAt time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	ExitCode  int       `json:"exit_code"`
	DurationMs int64    `json:"duration_ms"`
}

func (r *ExportRepository) fetchRows(command string) ([]ExportRow, error) {
	query := `SELECT id, command, started_at, finished_at, exit_code,
		CAST((julianday(COALESCE(finished_at, started_at)) - julianday(started_at)) * 86400000 AS INTEGER)
		FROM job_runs`
	args := []interface{}{}
	if command != "" {
		query += " WHERE command = ?"
		args = append(args, command)
	}
	query += " ORDER BY started_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("export query: %w", err)
	}
	defer rows.Close()

	var result []ExportRow
	for rows.Next() {
		var row ExportRow
		var fin nullTime
		if err := rows.Scan(&row.ID, &row.Command, &row.StartedAt, &fin, &row.ExitCode, &row.DurationMs); err != nil {
			return nil, err
		}
		if fin.Valid {
			row.FinishedAt = &fin.Time
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// WriteCSV writes job runs as CSV to the provided writer.
func (r *ExportRepository) WriteCSV(w io.Writer, command string) error {
	rows, err := r.fetchRows(command)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "command", "started_at", "finished_at", "exit_code", "duration_ms"})
	for _, row := range rows {
		finStr := ""
		if row.FinishedAt != nil {
			finStr = row.FinishedAt.Format(time.RFC3339)
		}
		_ = cw.Write([]string{
			fmt.Sprintf("%d", row.ID),
			row.Command,
			row.StartedAt.Format(time.RFC3339),
			finStr,
			fmt.Sprintf("%d", row.ExitCode),
			fmt.Sprintf("%d", row.DurationMs),
		})
	}
	cw.Flush()
	return cw.Error()
}

// WriteJSON writes job runs as JSON to the provided writer.
func (r *ExportRepository) WriteJSON(w io.Writer, command string) error {
	rows, err := r.fetchRows(command)
	if err != nil {
		return err
	}
	if rows == nil {
		rows = []ExportRow{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}
