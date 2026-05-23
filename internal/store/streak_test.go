package store_test

import (
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func TestGetStreakEmpty(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewStreakRepository(db)

	result, err := repo.GetStreak("nonexistent-command")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result for unknown command, got %+v", result)
	}
}

func TestGetStreakSuccess(t *testing.T) {
	db := openTestDB(t)
	runRepo := NewJobRunRepository(db)
	streakRepo := store.NewStreakRepository(db)

	cmd := "echo hello"
	for i := 0; i < 3; i++ {
		start := time.Now().UTC().Add(-time.Duration(i) * time.Minute)
		end := start.Add(2 * time.Second)
		if err := runRepo.Insert(cmd, start, end, 0); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
	}

	result, err := streakRepo.GetStreak(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.CurrentStreak != 3 {
		t.Errorf("expected streak 3, got %d", result.CurrentStreak)
	}
	if result.StreakType != "success" {
		t.Errorf("expected streak type 'success', got %q", result.StreakType)
	}
}

func TestGetStreakFailureBroken(t *testing.T) {
	db := openTestDB(t)
	runRepo := NewJobRunRepository(db)
	streakRepo := store.NewStreakRepository(db)

	cmd := "backup.sh"
	now := time.Now().UTC()
	// Two failures, then one success (oldest)
	inserts := []struct {
		offset   time.Duration
		exitCode int
	}{
		{0, 1},
		{-1 * time.Minute, 1},
		{-2 * time.Minute, 0},
	}
	for _, ins := range inserts {
		start := now.Add(ins.offset)
		end := start.Add(time.Second)
		if err := runRepo.Insert(cmd, start, end, ins.exitCode); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
	}

	result, err := streakRepo.GetStreak(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.CurrentStreak != 2 {
		t.Errorf("expected streak 2, got %d", result.CurrentStreak)
	}
	if result.StreakType != "failure" {
		t.Errorf("expected streak type 'failure', got %q", result.StreakType)
	}
}
