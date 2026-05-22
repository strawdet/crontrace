package store

import (
	"database/sql"
	"fmt"
	"time"
)

// TagRepository handles tagging of job runs for easier filtering and grouping.
type TagRepository struct {
	db *sql.DB
}

// NewTagRepository creates a new TagRepository and ensures the tags table exists.
func NewTagRepository(db *sql.DB) (*TagRepository, error) {
	schema := `
	CREATE TABLE IF NOT EXISTS job_run_tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		tag TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (run_id) REFERENCES job_runs(id) ON DELETE CASCADE,
		UNIQUE(run_id, tag)
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("create job_run_tags table: %w", err)
	}
	return &TagRepository{db: db}, nil
}

// AddTag attaches a tag to a specific job run.
func (r *TagRepository) AddTag(runID int64, tag string) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO job_run_tags (run_id, tag, created_at) VALUES (?, ?, ?)`,
		runID, tag, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("add tag: %w", err)
	}
	return nil
}

// ListTags returns all distinct tags associated with a given run ID.
func (r *TagRepository) ListTags(runID int64) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT tag FROM job_run_tags WHERE run_id = ? ORDER BY tag ASC`, runID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// RemoveTag removes a specific tag from a job run.
func (r *TagRepository) RemoveTag(runID int64, tag string) error {
	_, err := r.db.Exec(
		`DELETE FROM job_run_tags WHERE run_id = ? AND tag = ?`, runID, tag,
	)
	if err != nil {
		return fmt.Errorf("remove tag: %w", err)
	}
	return nil
}

// RunIDsByTag returns all run IDs that have the given tag.
func (r *TagRepository) RunIDsByTag(tag string) ([]int64, error) {
	rows, err := r.db.Query(
		`SELECT run_id FROM job_run_tags WHERE tag = ? ORDER BY run_id DESC`, tag,
	)
	if err != nil {
		return nil, fmt.Errorf("runs by tag: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
