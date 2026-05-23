package store_test

import (
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func TestRetentionByAge(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)
	retRepo := store.NewRetentionRepository(db)

	old := time.Now().Add(-10 * 24 * time.Hour)
	recent := time.Now().Add(-1 * time.Hour)

	if err := repo.Insert(store.JobRun{Command: "backup", StartedAt: old, ExitCode: 0}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Insert(store.JobRun{Command: "backup", StartedAt: recent, ExitCode: 0}); err != nil {
		t.Fatal(err)
	}

	deleted, err := retRepo.Apply(store.RetentionPolicy{MaxAgeDays: 5})
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	runs, err := repo.List("")
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 remaining run, got %d", len(runs))
	}
}

func TestRetentionByCount(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)
	retRepo := store.NewRetentionRepository(db)

	base := time.Now().Add(-5 * time.Hour)
	for i := 0; i < 5; i++ {
		err := repo.Insert(store.JobRun{
			Command:   "sync",
			StartedAt: base.Add(time.Duration(i) * time.Minute),
			ExitCode:  0,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	deleted, err := retRepo.Apply(store.RetentionPolicy{MaxRunsPerCommand: 3})
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", deleted)
	}

	runs, err := repo.List("sync")
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 3 {
		t.Errorf("expected 3 remaining runs, got %d", len(runs))
	}
}

func TestRetentionNoPolicyNoOp(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)
	retRepo := store.NewRetentionRepository(db)

	_ = repo.Insert(store.JobRun{Command: "check", StartedAt: time.Now(), ExitCode: 0})

	deleted, err := retRepo.Apply(store.RetentionPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}
