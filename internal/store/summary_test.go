package store_test

import (
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func TestAllCommandSummariesEmpty(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewSummaryRepository(db)

	summaries, err := repo.AllCommandSummaries()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestAllCommandSummariesWithData(t *testing.T) {
	db := openTestDB(t)
	jobRepo := store.NewJobRunRepository(db)
	sumRepo := store.NewSummaryRepository(db)

	now := time.Now().UTC()

	runs := []struct {
		cmd      string
		exit     int
		duration int64
	}{
		{"backup.sh", 0, 2000},
		{"backup.sh", 0, 4000},
		{"backup.sh", 1, 1000},
		{"cleanup.sh", 0, 500},
	}

	for _, r := range runs {
		if err := jobRepo.Insert(store.JobRun{
			Command:    r.cmd,
			StartedAt:  now,
			FinishedAt: now.Add(time.Duration(r.duration) * time.Millisecond),
			DurationMS: r.duration,
			ExitCode:   r.exit,
		}); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
	}

	summaries, err := sumRepo.AllCommandSummaries()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}

	// summaries are ordered by command ASC: backup.sh first
	backup := summaries[0]
	if backup.Command != "backup.sh" {
		t.Errorf("expected backup.sh, got %s", backup.Command)
	}
	if backup.TotalRuns != 3 {
		t.Errorf("expected 3 total runs, got %d", backup.TotalRuns)
	}
	if backup.SuccessCount != 2 {
		t.Errorf("expected 2 successes, got %d", backup.SuccessCount)
	}
	if backup.FailureCount != 1 {
		t.Errorf("expected 1 failure, got %d", backup.FailureCount)
	}
	// avg duration: (2000+4000+1000)/3 / 1000 = 2.333...
	if backup.AvgDurationS < 2.0 || backup.AvgDurationS > 3.0 {
		t.Errorf("unexpected avg duration: %f", backup.AvgDurationS)
	}

	cleanup := summaries[1]
	if cleanup.Command != "cleanup.sh" {
		t.Errorf("expected cleanup.sh, got %s", cleanup.Command)
	}
	if cleanup.TotalRuns != 1 {
		t.Errorf("expected 1 total run, got %d", cleanup.TotalRuns)
	}
}
