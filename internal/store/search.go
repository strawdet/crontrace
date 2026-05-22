package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SearchRepository provides search/filter functionality over job runs.
type SearchRepository struct {
	db *sql.DB
}

// NewSearchRepository creates a new SearchRepository.
func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// SearchFilter holds optional filter criteria for querying job runs.
type SearchFilter struct {
	Command  string
	ExitCode *int
	Since    *time.Time
	Limit    int
}

// Search returns job runs matching the given filter.
func (r *SearchRepository) Search(f SearchFilter) ([]JobRun, error) {
	query := `SELECT id, command, started_at, finished_at, exit_code FROM job_runs WHERE 1=1`
	args := []interface{}{}

	if f.Command != "" {
		query += " AND command LIKE ?"
		args = append(args, "%"+f.Command+"%")
	}

	if f.ExitCode != nil {
		query += " AND exit_code = ?"
		args = append(args, *f.ExitCode)
	}

	if f.Since != nil {
		query += " AND started_at >= ?"
		args = append(args, f.Since.UTC().Format(time.RFC3339))
	}

	query += " ORDER BY started_at DESC"

	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", f.Limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var runs []JobRun
	for rows.Next() {
		var jr JobRun
		var finishedAt nullTime
		if err := rows.Scan(&jr.ID, &jr.Command, &jr.StartedAt, &finishedAt, &jr.ExitCode); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		if finishedAt.Valid {
			jr.FinishedAt = &finishedAt.Time
		}
		runs = append(runs, jr)
	}
	return runs, rows.Err()
}

// CommandList returns a deduplicated, sorted list of all known commands.
func (r *SearchRepository) CommandList() ([]string, error) {
	rows, err := r.db.Query(`SELECT DISTINCT command FROM job_runs ORDER BY command ASC`)
	if err != nil {
		return nil, fmt.Errorf("command list query: %w", err)
	}
	defer rows.Close()

	var cmds []string
	for rows.Next() {
		var cmd string
		if err := rows.Scan(&cmd); err != nil {
			return nil, err
		}
		cmds = append(cmds, strings.TrimSpace(cmd))
	}
	return cmds, rows.Err()
}
