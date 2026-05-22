package store_test

import (
	"testing"

	"github.com/user/crontrace/internal/store"
)

func TestNotifyUpsertAndGet(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewNotifyRepository(db)
	if err != nil {
		t.Fatalf("NewNotifyRepository: %v", err)
	}

	rule := store.NotifyRule{
		Command:       "backup.sh",
		MaxDurationMs: 5000,
		AlertOnFail:   true,
	}
	if err := repo.Upsert(rule); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := repo.Get("backup.sh")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("expected rule, got nil")
	}
	if got.MaxDurationMs != 5000 {
		t.Errorf("MaxDurationMs: want 5000, got %d", got.MaxDurationMs)
	}
	if !got.AlertOnFail {
		t.Error("expected AlertOnFail to be true")
	}
}

func TestNotifyUpsertUpdates(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewNotifyRepository(db)

	_ = repo.Upsert(store.NotifyRule{Command: "sync.sh", MaxDurationMs: 1000, AlertOnFail: true})
	_ = repo.Upsert(store.NotifyRule{Command: "sync.sh", MaxDurationMs: 9999, AlertOnFail: false})

	got, _ := repo.Get("sync.sh")
	if got.MaxDurationMs != 9999 {
		t.Errorf("expected updated MaxDurationMs 9999, got %d", got.MaxDurationMs)
	}
	if got.AlertOnFail {
		t.Error("expected AlertOnFail false after update")
	}
}

func TestNotifyGetMissing(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewNotifyRepository(db)

	got, err := repo.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != nil {
		t.Error("expected nil for missing rule")
	}
}

func TestNotifyDeleteAndList(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewNotifyRepository(db)

	_ = repo.Upsert(store.NotifyRule{Command: "a.sh", MaxDurationMs: 100, AlertOnFail: true})
	_ = repo.Upsert(store.NotifyRule{Command: "b.sh", MaxDurationMs: 200, AlertOnFail: false})

	if err := repo.Delete("a.sh"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	rules, err := repo.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Command != "b.sh" {
		t.Errorf("expected b.sh, got %s", rules[0].Command)
	}
}
