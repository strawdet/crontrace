package store_test

import (
	"testing"

	"github.com/user/crontrace/internal/store"
)

func TestWatchUpsertAndList(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewWatchRepository(db)
	if err != nil {
		t.Fatalf("NewWatchRepository: %v", err)
	}

	if err := repo.Upsert("/bin/backup.sh", 120, 0); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if err := repo.Upsert("/bin/sync.sh", 60, 0); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	entries, err := repo.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestWatchUpsertUpdates(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewWatchRepository(db)
	if err != nil {
		t.Fatalf("NewWatchRepository: %v", err)
	}

	if err := repo.Upsert("/bin/job.sh", 30, 0); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if err := repo.Upsert("/bin/job.sh", 90, 1); err != nil {
		t.Fatalf("Upsert update: %v", err)
	}

	e, err := repo.Get("/bin/job.sh")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if e == nil {
		t.Fatal("expected entry, got nil")
	}
	if e.MaxDurationSec != 90 {
		t.Errorf("expected MaxDurationSec=90, got %d", e.MaxDurationSec)
	}
	if e.ExpectExitCode != 1 {
		t.Errorf("expected ExpectExitCode=1, got %d", e.ExpectExitCode)
	}
}

func TestWatchDelete(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewWatchRepository(db)
	if err != nil {
		t.Fatalf("NewWatchRepository: %v", err)
	}

	_ = repo.Upsert("/bin/job.sh", 60, 0)
	if err := repo.Delete("/bin/job.sh"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	e, err := repo.Get("/bin/job.sh")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if e != nil {
		t.Error("expected nil after delete")
	}
}

func TestWatchGetMissing(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewWatchRepository(db)
	if err != nil {
		t.Fatalf("NewWatchRepository: %v", err)
	}

	e, err := repo.Get("/bin/nonexistent.sh")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if e != nil {
		t.Error("expected nil for missing entry")
	}
}
