package store_test

import (
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func TestStatsByCommandEmpty(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewStatsRepository(db)

	stats, err := repo.ByCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("expected 0 stats, got %d", len(stats))
	}
}

func TestStatsByCommandCounts(t *testing.T) {
	db := openTestDB(t)
	runs := store.NewJobRunRepository(db)
	statsRepo := store.NewStatsRepository(db)

	now := time.Now().UTC()
	end := now.Add(2 * time.Second)

	for i := 0; i < 3; i++ {
		r := &store.JobRun{
			Command:   "echo hello",
			StartedAt: now,
			EndedAt:   &end,
			ExitCode:  0,
		}
		if err := runs.Insert(r); err != nil {
			t.Fatalf("insert error: %v", err)
		}
	}
	// one failure
	failRun := &store.JobRun{
		Command:   "echo hello",
		StartedAt: now,
		EndedAt:   &end,
		ExitCode:  1,
	}
	if err := runs.Insert(failRun); err != nil {
		t.Fatalf("insert error: %v", err)
	}

	stats, err := statsRepo.ByCommand()
	if err != nil {
		t.Fatalf("stats error: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 command stat, got %d", len(stats))
	}
	s := stats[0]
	if s.TotalRuns != 4 {
		t.Errorf("expected 4 total runs, got %d", s.TotalRuns)
	}
	if s.SuccessRuns != 3 {
		t.Errorf("expected 3 success runs, got %d", s.SuccessRuns)
	}
	if s.FailureRuns != 1 {
		t.Errorf("expected 1 failure run, got %d", s.FailureRuns)
	}
	if s.AvgDurationS <= 0 {
		t.Errorf("expected positive avg duration, got %f", s.AvgDurationS)
	}
}
