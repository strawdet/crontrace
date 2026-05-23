package store_test

import (
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func TestGetReplayByIDNotFound(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewReplayRepository(db)

	_, err := repo.GetReplayByID(999)
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestGetReplayByID(t *testing.T) {
	db := openTestDB(t)
	jr := store.NewJobRunRepository(db)
	repo := store.NewReplayRepository(db)

	run := &store.JobRun{
		Command:   "echo",
		Args:      "hello",
		StartedAt: time.Now().UTC().Truncate(time.Second),
		ExitCode:  0,
	}
	if err := jr.Insert(run); err != nil {
		t.Fatalf("insert: %v", err)
	}

	runs, err := jr.List(10)
	if err != nil || len(runs) == 0 {
		t.Fatalf("list: %v", err)
	}

	entry, err := repo.GetReplayByID(runs[0].ID)
	if err != nil {
		t.Fatalf("get replay: %v", err)
	}
	if entry.Command != "echo" {
		t.Errorf("expected command 'echo', got %q", entry.Command)
	}
	if entry.Args != "hello" {
		t.Errorf("expected args 'hello', got %q", entry.Args)
	}
}

func TestListReplayable(t *testing.T) {
	db := openTestDB(t)
	jr := store.NewJobRunRepository(db)
	repo := store.NewReplayRepository(db)

	commands := []string{"backup.sh", "sync.sh", "backup.sh"}
	for _, cmd := range commands {
		run := &store.JobRun{
			Command:   cmd,
			Args:      "",
			StartedAt: time.Now().UTC().Truncate(time.Second),
			ExitCode:  0,
		}
		if err := jr.Insert(run); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	entries, err := repo.ListReplayable(10)
	if err != nil {
		t.Fatalf("list replayable: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 unique commands, got %d", len(entries))
	}
}
