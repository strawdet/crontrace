package store

import (
	"database/sql"
	"time"
)

// CommandSummary holds aggregated statistics for a single command.
type CommandSummary struct {
	Command      string
	TotalRuns    int
	SuccessCount int
	FailureCount int
	AvgDurationS float64
	LastRunAt    time.Time
}

// SummaryRepository provides aggregated command-level summaries.
type SummaryRepository struct {
	db *sql.DB
}

// NewSummaryRepository creates a SummaryRepository backed by db.
func NewSummaryRepository(db *sql.DB) *SummaryRepository {
	return &SummaryRepository{db: db}
}

// AllCommandSummaries returns one summary row per distinct command.
func (r *SummaryRepository) AllCommandSummaries() ([]CommandSummary, error) {
	const q = `
		SELECT
			command,
			COUNT(*) AS total_runs,
			SUM(CASE WHEN exit_code = 0 THEN 1 ELSE 0 END) AS success_count,
			SUM(CASE WHEN exit_code != 0 THEN 1 ELSE 0 END) AS failure_count,
			AVG(duration_ms) / 1000.0 AS avg_duration_s,
			MAX(started_at) AS last_run_at
		FROM job_runs
		GROUP BY command
		ORDER BY command ASC`

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []CommandSummary
	for rows.Next() {
		var s CommandSummary
		var lastRunAt string
		if err := rows.Scan(
			&s.Command,
			&s.TotalRuns,
			&s.SuccessCount,
			&s.FailureCount,
			&s.AvgDurationS,
			&lastRunAt,
		); err != nil {
			return nil, err
		}
		s.LastRunAt, _ = time.Parse(time.RFC3339, lastRunAt)
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}
