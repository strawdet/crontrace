package store

import (
	"database/sql"
)

// JobStats holds aggregate statistics for a specific command.
type JobStats struct {
	Command      string
	TotalRuns    int
	SuccessRuns  int
	FailureRuns  int
	AvgDurationS float64
	LastExitCode int
}

// StatsRepository computes aggregate statistics over job runs.
type StatsRepository struct {
	db *sql.DB
}

// NewStatsRepository creates a StatsRepository backed by db.
func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// ByCommand returns per-command statistics for all recorded commands.
func (r *StatsRepository) ByCommand() ([]JobStats, error) {
	const q = `
		SELECT
			command,
			COUNT(*) AS total_runs,
			SUM(CASE WHEN exit_code = 0 THEN 1 ELSE 0 END) AS success_runs,
			SUM(CASE WHEN exit_code != 0 THEN 1 ELSE 0 END) AS failure_runs,
			AVG(CASE WHEN ended_at IS NOT NULL
				THEN (julianday(ended_at) - julianday(started_at)) * 86400.0
				ELSE NULL END) AS avg_duration_s,
			MAX(exit_code) AS last_exit_code
		FROM job_runs
		GROUP BY command
		ORDER BY total_runs DESC`

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []JobStats
	for rows.Next() {
		var s JobStats
		var avgDur sql.NullFloat64
		if err := rows.Scan(&s.Command, &s.TotalRuns, &s.SuccessRuns, &s.FailureRuns, &avgDur, &s.LastExitCode); err != nil {
			return nil, err
		}
		if avgDur.Valid {
			s.AvgDurationS = avgDur.Float64
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
