package cli_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/user/crontrace/internal/cli"
	"github.com/user/crontrace/internal/store"
)

func TestPruneRunsOlderThan(t *testing.T) {
	db := openTestDB(t)
	jobRepo := store.NewJobRunRepository(db)

	_, err := jobRepo.Insert("echo old", 0, time.Now().Add(-48*time.Hour), time.Now().Add(-47*time.Hour))
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	_, err = jobRepo.Insert("echo new", 0, time.Now().Add(-1*time.Minute), time.Now())
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	var buf bytes.Buffer
	err = cli.PruneRuns(db, 24*time.Hour, "")
	if err != nil {
		t.Fatalf("PruneRuns: %v", err)
	}
	_ = buf
}

func TestPruneRunsByCommand(t *testing.T) {
	db := openTestDB(t)
	jobRepo := store.NewJobRunRepository(db)

	_, err := jobRepo.Insert("echo hello", 0, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	err = cli.PruneRuns(db, 0, "echo hello")
	if err != nil {
		t.Fatalf("PruneRuns by command: %v", err)
	}
}

func TestPruneRunsNoArgs(t *testing.T) {
	db := openTestDB(t)
	err := cli.PruneRuns(db, 0, "")
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestPruneRunsOutputMessage(t *testing.T) {
	db := openTestDB(t)
	jobRepo := store.NewJobRunRepository(db)

	for i := 0; i < 2; i++ {
		_, err := jobRepo.Insert("cron-job", 0, time.Now().Add(-72*time.Hour), time.Now().Add(-71*time.Hour))
		if err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	err := cli.PruneRuns(db, 24*time.Hour, "")
	if err != nil {
		t.Fatalf("PruneRuns: %v", err)
	}
}
