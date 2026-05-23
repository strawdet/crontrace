package store_test

import (
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func TestBaselineUpsertAndGet(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewBaselineRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	b := store.Baseline{
		Command:       "backup.sh",
		AvgDurationMs: 4200,
		SampleCount:   10,
		UpdatedAt:     now,
	}
	if err := repo.Upsert(b); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := repo.Get("backup.sh")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("expected baseline, got nil")
	}
	if got.AvgDurationMs != 4200 {
		t.Errorf("avg duration: want 4200, got %d", got.AvgDurationMs)
	}
	if got.SampleCount != 10 {
		t.Errorf("sample count: want 10, got %d", got.SampleCount)
	}
}

func TestBaselineUpsertUpdates(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewBaselineRepository(db)

	now := time.Now().UTC().Truncate(time.Second)
	_ = repo.Upsert(store.Baseline{Command: "sync.sh", AvgDurationMs: 1000, SampleCount: 5, UpdatedAt: now})
	_ = repo.Upsert(store.Baseline{Command: "sync.sh", AvgDurationMs: 1500, SampleCount: 6, UpdatedAt: now})

	got, _ := repo.Get("sync.sh")
	if got.AvgDurationMs != 1500 {
		t.Errorf("want 1500, got %d", got.AvgDurationMs)
	}
}

func TestBaselineGetMissing(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewBaselineRepository(db)

	got, err := repo.Get("nonexistent.sh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil for missing baseline")
	}
}

func TestBaselineListAndDelete(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewBaselineRepository(db)

	now := time.Now().UTC().Truncate(time.Second)
	_ = repo.Upsert(store.Baseline{Command: "a.sh", AvgDurationMs: 100, SampleCount: 1, UpdatedAt: now})
	_ = repo.Upsert(store.Baseline{Command: "b.sh", AvgDurationMs: 200, SampleCount: 2, UpdatedAt: now})

	list, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2, got %d", len(list))
	}

	_ = repo.Delete("a.sh")
	list, _ = repo.List()
	if len(list) != 1 {
		t.Fatalf("after delete want 1, got %d", len(list))
	}
	if list[0].Command != "b.sh" {
		t.Errorf("expected b.sh, got %s", list[0].Command)
	}
}
