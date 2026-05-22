package cli_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/user/crontrace/internal/cli"
	"github.com/user/crontrace/internal/store"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestListRunsEmpty(t *testing.T) {
	repo := store.NewJobRunRepository(openTestDB(t))

	// Redirect stdout is complex; just ensure no error is returned.
	opts := cli.ListOptions{Limit: 10}
	if err := cli.ListRuns(repo, opts); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestListRunsWithData(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)

	// Insert a completed run directly via the runner path.
	_, err := db.Exec(
		`INSERT INTO job_runs (job_name, command, started_at, finished_at, exit_code)
		 VALUES (?, ?, datetime('now','-1 minute'), datetime('now'), ?)`,
		"backup", "tar -czf /tmp/backup.tar.gz /data", 0,
	)
	if err != nil {
		t.Fatalf("insert seed row: %v", err)
	}

	// Capture output by temporarily redirecting os.Stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := cli.ListOptions{JobName: "backup", Limit: 5}
	if err := cli.ListRuns(repo, opts); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if output == "" {
		t.Error("expected non-empty output")
	}
	if got := output; len(got) == 0 {
		t.Error("output should contain table header")
	}
}
