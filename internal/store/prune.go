package store

import (
	"database/sql"
	"fmt"
	"time"
)

// PruneRepository handles deletion of old job run records.
type PruneRepository struct {
	db *sql.DB
}

// NewPruneRepository creates a new PruneRepository.
func NewPruneRepository(db *sql.DB) *PruneRepository {
	return &PruneRepository{db: db}
}

// PruneOlderThan deletes job runs whose started_at is older than the given duration.
// Returns the number of rows deleted.
func (r *PruneRepository) PruneOlderThan(age time.Duration) (int64, error) {
	cutoff := time.Now().Add(-age)
	res, err := r.db.Exec(
		`DELETE FROM job_runs WHERE started_at < ?`,
		cutoff.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("prune: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("prune rows affected: %w", err)
	}
	return n, nil
}

// PruneByCommand deletes all job runs for a specific command.
// Returns the number of rows deleted.
func (r *PruneRepository) PruneByCommand(command string) (int64, error) {
	res, err := r.db.Exec(`DELETE FROM job_runs WHERE command = ?`, command)
	if err != nil {
		return 0, fmt.Errorf("prune by command: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("prune by command rows affected: %w", err)
	}
	return n, nil
}
