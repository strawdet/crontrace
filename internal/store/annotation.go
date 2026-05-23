package store

import (
	"database/sql"
	"time"
)

// Annotation holds a free-text note attached to a job run.
type Annotation struct {
	ID        int64
	RunID     int64
	Note      string
	CreatedAt time.Time
}

// AnnotationRepository provides persistence for run annotations.
type AnnotationRepository struct {
	db *sql.DB
}

// NewAnnotationRepository initialises the annotations table and returns a repository.
func NewAnnotationRepository(db *sql.DB) (*AnnotationRepository, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS annotations (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id     INTEGER NOT NULL,
			note       TEXT    NOT NULL,
			created_at DATETIME NOT NULL
		)`)
	if err != nil {
		return nil, err
	}
	return &AnnotationRepository{db: db}, nil
}

// Add attaches a note to the given run ID.
func (r *AnnotationRepository) Add(runID int64, note string) error {
	_, err := r.db.Exec(
		`INSERT INTO annotations (run_id, note, created_at) VALUES (?, ?, ?)`,
		runID, note, time.Now().UTC(),
	)
	return err
}

// ListByRun returns all annotations for the given run ID, oldest first.
func (r *AnnotationRepository) ListByRun(runID int64) ([]Annotation, error) {
	rows, err := r.db.Query(
		`SELECT id, run_id, note, created_at FROM annotations WHERE run_id = ? ORDER BY created_at ASC`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Annotation
	for rows.Next() {
		var a Annotation
		if err := rows.Scan(&a.ID, &a.RunID, &a.Note, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Delete removes a single annotation by its ID.
func (r *AnnotationRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM annotations WHERE id = ?`, id)
	return err
}
