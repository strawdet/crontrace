package store

import (
	"database/sql"
	"time"
)

// DashboardSummary holds aggregated metrics for the dashboard view.
type DashboardSummary struct {
	TotalRuns      int
	SuccessRuns    int
	FailedRuns     int
	UniqueCommands int
	AvgDurationMs  float64
	LastRunAt      *time.Time
}

// DashboardRepository provides dashboard-level aggregated queries.
type DashboardRepository struct {
	db *sql.DB
}

// NewDashboardRepository creates a new DashboardRepository.
func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// GetSummary returns an aggregated summary of all job runs.
func (r *DashboardRepository) GetSummary() (DashboardSummary, error) {
	var s DashboardSummary

	row := r.db.QueryRow(`
		SELECT
			COUNT(*) AS total_runs,
			SUM(CASE WHEN exit_code = 0 THEN 1 ELSE 0 END) AS success_runs,
			SUM(CASE WHEN exit_code != 0 THEN 1 ELSE 0 END) AS failed_runs,
			COUNT(DISTINCT command) AS unique_commands,
			AVG(duration_ms) AS avg_duration_ms,
			MAX(started_at) AS last_run_at
		FROM job_runs
	`)

	var lastRunAt sql.NullString
	var avgDuration sql.NullFloat64

	if err := row.Scan(
		&s.TotalRuns,
		&s.SuccessRuns,
		&s.FailedRuns,
		&s.UniqueCommands,
		&avgDuration,
		&lastRunAt,
	); err != nil {
		return s, err
	}

	if avgDuration.Valid {
		s.AvgDurationMs = avgDuration.Float64
	}

	if lastRunAt.Valid {
		t, err := time.Parse(time.RFC3339, lastRunAt.String)
		if err == nil {
			s.LastRunAt = &t
		}
	}

	return s, nil
}
