package store

import (
	"database/sql"
	"fmt"
)

// RetentionPolicy defines how long to keep job run records.
type RetentionPolicy struct {
	MaxAgeDays int
	MaxRunsPerCommand int
}

// RetentionRepository manages automatic retention policies.
type RetentionRepository struct {
	db *sql.DB
}

// NewRetentionRepository creates a new RetentionRepository.
func NewRetentionRepository(db *sql.DB) *RetentionRepository {
	return &RetentionRepository{db: db}
}

// Apply enforces the given retention policy, returning the number of deleted rows.
func (r *RetentionRepository) Apply(policy RetentionPolicy) (int64, error) {
	var total int64

	if policy.MaxAgeDays > 0 {
		res, err := r.db.Exec(
			`DELETE FROM job_runs WHERE started_at < datetime('now', ?)`,
			fmt.Sprintf("-%d days", policy.MaxAgeDays),
		)
		if err != nil {
			return total, fmt.Errorf("retention by age: %w", err)
		}
		n, _ := res.RowsAffected()
		total += n
	}

	if policy.MaxRunsPerCommand > 0 {
		rows, err := r.db.Query(`SELECT DISTINCT command FROM job_runs`)
		if err != nil {
			return total, fmt.Errorf("listing commands: %w", err)
		}
		defer rows.Close()

		var commands []string
		for rows.Next() {
			var cmd string
			if err := rows.Scan(&cmd); err != nil {
				return total, err
			}
			commands = append(commands, cmd)
		}

		for _, cmd := range commands {
			res, err := r.db.Exec(`
				DELETE FROM job_runs
				WHERE command = ? AND id NOT IN (
					SELECT id FROM job_runs WHERE command = ?
					ORDER BY started_at DESC LIMIT ?
				)`, cmd, cmd, policy.MaxRunsPerCommand)
			if err != nil {
				return total, fmt.Errorf("retention by count for %q: %w", cmd, err)
			}
			n, _ := res.RowsAffected()
			total += n
		}
	}

	return total, nil
}
