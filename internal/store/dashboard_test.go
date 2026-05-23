package store_test

import (
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func TestDashboardSummaryEmpty(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewDashboardRepository(db)

	s, err := repo.GetSummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.TotalRuns != 0 {
		t.Errorf("expected 0 total runs, got %d", s.TotalRuns)
	}
	if s.UniqueCommands != 0 {
		t.Errorf("expected 0 unique commands, got %d", s.UniqueCommands)
	}
	if s.LastRunAt != nil {
		t.Errorf("expected nil LastRunAt, got %v", s.LastRunAt)
	}
}

func TestDashboardSummaryWithData(t *testing.T) {
	db := openTestDB(t)
	jobRepo := store.NewJobRunRepository(db)
	dashRepo := store.NewDashboardRepository(db)

	now := time.Now().UTC()

	runs := []store.JobRun{
		{Command: "backup.sh", StartedAt: now.Add(-2 * time.Hour), FinishedAt: now.Add(-2*time.Hour + 30*time.Second), DurationMs: 30000, ExitCode: 0},
		{Command: "backup.sh", StartedAt: now.Add(-1 * time.Hour), FinishedAt: now.Add(-1*time.Hour + 20*time.Second), DurationMs: 20000, ExitCode: 1},
		{Command: "cleanup.sh", StartedAt: now.Add(-30 * time.Minute), FinishedAt: now.Add(-30*time.Minute + 10*time.Second), DurationMs: 10000, ExitCode: 0},
	}

	for _, r := range runs {
		if err := jobRepo.Insert(r); err != nil {
			t.Fatalf("insert error: %v", err)
		}
	}

	s, err := dashRepo.GetSummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.TotalRuns != 3 {
		t.Errorf("expected 3 total runs, got %d", s.TotalRuns)
	}
	if s.SuccessRuns != 2 {
		t.Errorf("expected 2 success runs, got %d", s.SuccessRuns)
	}
	if s.FailedRuns != 1 {
		t.Errorf("expected 1 failed run, got %d", s.FailedRuns)
	}
	if s.UniqueCommands != 2 {
		t.Errorf("expected 2 unique commands, got %d", s.UniqueCommands)
	}
	if s.AvgDurationMs == 0 {
		t.Errorf("expected non-zero avg duration")
	}
	if s.LastRunAt == nil {
		t.Errorf("expected non-nil LastRunAt")
	}
}
