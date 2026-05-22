package runner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/crontrace/internal/runner"
	"github.com/user/crontrace/internal/store"
)

func openTestRepo(t *testing.T) *store.JobRunRepository {
	t.Helper()
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return store.NewJobRunRepository(db)
}

func TestRunSuccess(t *testing.T) {
	repo := openTestRepo(t)

	result, err := runner.Run(repo, "echo", []string{"hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Finished.Before(result.Started) {
		t.Error("finished time is before started time")
	}

	runs, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].ExitCode != 0 {
		t.Errorf("stored exit code mismatch: got %d", runs[0].ExitCode)
	}
}

func TestRunFailure(t *testing.T) {
	repo := openTestRepo(t)

	// 'false' exits with code 1 on all POSIX systems.
	falseCmd := "false"
	if _, err := os.Stat("/usr/bin/false"); err != nil {
		t.Skip("'false' binary not available")
		_ = falseCmd
	}

	result, err := runner.Run(repo, "false", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}

	runs, _ := repo.List()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run stored, got %d", len(runs))
	}
	if runs[0].ExitCode == 0 {
		t.Error("stored exit code should be non-zero")
	}
}
