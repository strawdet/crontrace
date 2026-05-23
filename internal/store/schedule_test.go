package store_test

import (
	"testing"

	"github.com/user/crontrace/internal/store"
)

func TestScheduleUpsertAndGet(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewScheduleRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	if err := repo.Upsert("backup.sh", "0 2 * * *"); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	s, err := repo.Get("backup.sh")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if s.Command != "backup.sh" {
		t.Errorf("expected command backup.sh, got %s", s.Command)
	}
	if s.CronExpr != "0 2 * * *" {
		t.Errorf("expected cron expr '0 2 * * *', got %s", s.CronExpr)
	}
}

func TestScheduleUpsertUpdates(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewScheduleRepository(db)

	_ = repo.Upsert("sync.sh", "*/5 * * * *")
	_ = repo.Upsert("sync.sh", "*/10 * * * *")

	s, err := repo.Get("sync.sh")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if s.CronExpr != "*/10 * * * *" {
		t.Errorf("expected updated expr, got %s", s.CronExpr)
	}
}

func TestScheduleGetMissing(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewScheduleRepository(db)

	_, err := repo.Get("nonexistent.sh")
	if err == nil {
		t.Fatal("expected error for missing schedule")
	}
}

func TestScheduleList(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewScheduleRepository(db)

	_ = repo.Upsert("a.sh", "@daily")
	_ = repo.Upsert("b.sh", "@hourly")

	list, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 schedules, got %d", len(list))
	}
}

func TestScheduleDelete(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewScheduleRepository(db)

	_ = repo.Upsert("cleanup.sh", "@weekly")
	if err := repo.Delete("cleanup.sh"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := repo.Get("cleanup.sh")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestScheduleListEmpty(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewScheduleRepository(db)

	list, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}
