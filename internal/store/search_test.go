package store_test

import (
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func TestSearchByCommand(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)
	search := store.NewSearchRepository(db)

	now := time.Now().UTC()
	_ = repo.Insert(store.JobRun{Command: "backup.sh", StartedAt: now, ExitCode: 0})
	_ = repo.Insert(store.JobRun{Command: "backup.sh", StartedAt: now, ExitCode: 1})
	_ = repo.Insert(store.JobRun{Command: "cleanup.sh", StartedAt: now, ExitCode: 0})

	runs, err := search.Search(store.SearchFilter{Command: "backup"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

func TestSearchByExitCode(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)
	search := store.NewSearchRepository(db)

	now := time.Now().UTC()
	_ = repo.Insert(store.JobRun{Command: "job.sh", StartedAt: now, ExitCode: 0})
	_ = repo.Insert(store.JobRun{Command: "job.sh", StartedAt: now, ExitCode: 1})
	_ = repo.Insert(store.JobRun{Command: "job.sh", StartedAt: now, ExitCode: 1})

	exitOne := 1
	runs, err := search.Search(store.SearchFilter{ExitCode: &exitOne})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 failed runs, got %d", len(runs))
	}
}

func TestSearchSince(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)
	search := store.NewSearchRepository(db)

	old := time.Now().UTC().Add(-48 * time.Hour)
	recent := time.Now().UTC()
	_ = repo.Insert(store.JobRun{Command: "old.sh", StartedAt: old, ExitCode: 0})
	_ = repo.Insert(store.JobRun{Command: "new.sh", StartedAt: recent, ExitCode: 0})

	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	runs, err := search.Search(store.SearchFilter{Since: &cutoff})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 recent run, got %d", len(runs))
	}
}

func TestCommandList(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)
	search := store.NewSearchRepository(db)

	now := time.Now().UTC()
	_ = repo.Insert(store.JobRun{Command: "alpha.sh", StartedAt: now, ExitCode: 0})
	_ = repo.Insert(store.JobRun{Command: "beta.sh", StartedAt: now, ExitCode: 0})
	_ = repo.Insert(store.JobRun{Command: "alpha.sh", StartedAt: now, ExitCode: 1})

	cmds, err := search.CommandList()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmds) != 2 {
		t.Errorf("expected 2 distinct commands, got %d", len(cmds))
	}
	if cmds[0] != "alpha.sh" || cmds[1] != "beta.sh" {
		t.Errorf("unexpected command order: %v", cmds)
	}
}
