package store_test

import (
	"testing"

	"github.com/user/crontrace/internal/store"
)

func TestAddAndListTags(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewTagRepository(db)
	if err != nil {
		t.Fatalf("NewTagRepository: %v", err)
	}

	if err := repo.AddTag(1, "nightly"); err != nil {
		t.Fatalf("AddTag: %v", err)
	}
	if err := repo.AddTag(1, "backup"); err != nil {
		t.Fatalf("AddTag: %v", err)
	}

	tags, err := repo.ListTags(1)
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "backup" || tags[1] != "nightly" {
		t.Errorf("unexpected tags order: %v", tags)
	}
}

func TestAddTagIdempotent(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewTagRepository(db)
	if err != nil {
		t.Fatalf("NewTagRepository: %v", err)
	}

	if err := repo.AddTag(2, "critical"); err != nil {
		t.Fatalf("first AddTag: %v", err)
	}
	if err := repo.AddTag(2, "critical"); err != nil {
		t.Fatalf("duplicate AddTag should not error: %v", err)
	}

	tags, err := repo.ListTags(2)
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}
}

func TestRemoveTag(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewTagRepository(db)
	if err != nil {
		t.Fatalf("NewTagRepository: %v", err)
	}

	_ = repo.AddTag(3, "weekly")
	_ = repo.AddTag(3, "report")

	if err := repo.RemoveTag(3, "weekly"); err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}

	tags, _ := repo.ListTags(3)
	if len(tags) != 1 || tags[0] != "report" {
		t.Errorf("expected only 'report' tag, got %v", tags)
	}
}

func TestRunIDsByTag(t *testing.T) {
	db := openTestDB(t)
	repo, err := store.NewTagRepository(db)
	if err != nil {
		t.Fatalf("NewTagRepository: %v", err)
	}

	_ = repo.AddTag(10, "prod")
	_ = repo.AddTag(11, "prod")
	_ = repo.AddTag(12, "staging")

	ids, err := repo.RunIDsByTag("prod")
	if err != nil {
		t.Fatalf("RunIDsByTag: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 run IDs for tag 'prod', got %d", len(ids))
	}
}

func TestListTagsEmpty(t *testing.T) {
	db := openTestDB(t)
	repo, _ := store.NewTagRepository(db)

	tags, err := repo.ListTags(999)
	if err != nil {
		t.Fatalf("ListTags on unknown run: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected empty tags, got %v", tags)
	}
}
