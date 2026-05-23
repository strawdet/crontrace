package store_test

import (
	"os"
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func openTestDB(t *testing.T) *store.DB {
	t.Helper()
	tmp, err := os.CreateTemp("", "crontrace-test-*.db")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	tmp.Close()
	t.Cleanup(func() { os.Remove(tmp.Name()) })

	db, err := store.Open(tmp.Name())
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestInsertAndList(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)

	now := time.Now().UTC().Truncate(time.Second)
	finished := now.Add(2 * time.Second)
	dur := int64(2000)
	code := 0

	run := &store.JobRun{
		Command:    "/usr/bin/backup.sh",
		StartedAt:  now,
		FinishedAt: &finished,
		DurationMs: &dur,
		ExitCode:   &code,
		Stdout:     "done",
		Stderr:     "",
	}

	if err := repo.Insert(run); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if run.ID == 0 {
		t.Error("expected non-zero ID after insert")
	}

	runs, err := repo.ListByCommand("/usr/bin/backup.sh", 10)
	if err != nil {
		t.Fatalf("ListByCommand: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	got := runs[0]
	if got.Command != run.Command {
		t.Errorf("command: got %q want %q", got.Command, run.Command)
	}
	if got.Stdout != "done" {
		t.Errorf("stdout: got %q want %q", got.Stdout, "done")
	}
	if got.ExitCode == nil || *got.ExitCode != 0 {
		t.Errorf("exit code: got %v want 0", got.ExitCode)
	}
}

func TestListReturnsEmpty(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)

	runs, err := repo.ListByCommand("nonexistent", 10)
	if err != nil {
		t.Fatalf("ListByCommand: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestListRespectsLimit(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)

	const command = "/usr/bin/backup.sh"
	const total = 5
	const limit = 3

	for i := 0; i < total; i++ {
		run := &store.JobRun{
			Command:   command,
			StartedAt: time.Now().UTC().Truncate(time.Second),
		}
		if err := repo.Insert(run); err != nil {
			t.Fatalf("Insert run %d: %v", i, err)
		}
	}

	runs, err := repo.ListByCommand(command, limit)
	if err != nil {
		t.Fatalf("ListByCommand: %v", err)
	}
	if len(runs) != limit {
		t.Errorf("expected %d runs, got %d", limit, len(runs))
	}
}
