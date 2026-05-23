package store_test

import (
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func TestPruneOlderThan(t *testing.T) {
	db := openTestDB(t)
	jobRepo := store.NewJobRunRepository(db)
	pruneRepo := store.NewPruneRepository(db)

	// Insert an old run (simulate by inserting then checking prune with 0 duration)
	_, err := jobRepo.Insert("echo old", 0, time.Now().Add(-48*time.Hour), time.Now().Add(-47*time.Hour))
	if err != nil {
		t.Fatalf("insert old: %v", err)
	}
	_, err = jobRepo.Insert("echo new", 0, time.Now().Add(-1*time.Minute), time.Now())
	if err != nil {
		t.Fatalf("insert new: %v", err)
	}

	n, err := pruneRepo.PruneOlderThan(24 * time.Hour)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 pruned row, got %d", n)
	}

	runs, err := jobRepo.List("", 100)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 remaining run, got %d", len(runs))
	}
}

func TestPruneByCommand(t *testing.T) {
	db := openTestDB(t)
	jobRepo := store.NewJobRunRepository(db)
	pruneRepo := store.NewPruneRepository(db)

	for i := 0; i < 3; i++ {
		_, err := jobRepo.Insert("echo hello", 0, time.Now(), time.Now())
		if err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	_, err := jobRepo.Insert("ls -la", 0, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("insert ls: %v", err)
	}

	n, err := pruneRepo.PruneByCommand("echo hello")
	if err != nil {
		t.Fatalf("prune by command: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 pruned rows, got %d", n)
	}

	runs, err := jobRepo.List("", 100)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 remaining run, got %d", len(runs))
	}
}

func TestPruneOlderThanNothingToDelete(t *testing.T) {
	db := openTestDB(t)
	pruneRepo := store.NewPruneRepository(db)

	n, err := pruneRepo.PruneOlderThan(24 * time.Hour)
	if err != nil {
		t.Fatalf("prune empty: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 pruned rows, got %d", n)
	}
}

func TestPruneByCommandNothingToDelete(t *testing.T) {
	db := openTestDB(t)
	jobRepo := store.NewJobRunRepository(db)
	pruneRepo := store.NewPruneRepository(db)

	_, err := jobRepo.Insert("echo hello", 0, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Prune a command that doesn't exist in the DB
	n, err := pruneRepo.PruneByCommand("nonexistent command")
	if err != nil {
		t.Fatalf("prune by command: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 pruned rows, got %d", n)
	}
}
