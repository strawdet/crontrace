package store_test

import (
	"testing"

	"github.com/user/crontrace/internal/store"
)

func openAlertTestDB(t *testing.T) *store.AlertRepository {
	t.Helper()
	db := openTestDB(t)
	repo, err := store.NewAlertRepository(db)
	if err != nil {
		t.Fatalf("NewAlertRepository: %v", err)
	}
	return repo
}

func TestAlertUpsertAndList(t *testing.T) {
	repo := openAlertTestDB(t)

	if err := repo.Upsert("/bin/backup", "duration", 30.0); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if err := repo.Upsert("/bin/backup", "exit_code", 1.0); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	alerts, err := repo.List("/bin/backup")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(alerts))
	}
}

func TestAlertUpsertUpdates(t *testing.T) {
	repo := openAlertTestDB(t)

	_ = repo.Upsert("/bin/job", "duration", 10.0)
	_ = repo.Upsert("/bin/job", "duration", 99.0)

	alerts, _ := repo.List("/bin/job")
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert after upsert, got %d", len(alerts))
	}
	if alerts[0].Threshold != 99.0 {
		t.Errorf("expected threshold 99.0, got %f", alerts[0].Threshold)
	}
}

func TestAlertDelete(t *testing.T) {
	repo := openAlertTestDB(t)

	_ = repo.Upsert("/bin/clean", "failure_rate", 0.5)
	n, err := repo.Delete("/bin/clean", "failure_rate")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 row deleted, got %d", n)
	}
	alerts, _ := repo.List("/bin/clean")
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts after delete, got %d", len(alerts))
	}
}

func TestAlertListEmpty(t *testing.T) {
	repo := openAlertTestDB(t)
	alerts, err := repo.List("")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected empty list, got %d", len(alerts))
	}
}
