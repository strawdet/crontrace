package store_test

import (
	"testing"

	"github.com/user/crontrace/internal/store"
)

func openAnnotationTestDB(t *testing.T) (*store.AnnotationRepository, func()) {
	t.Helper()
	db := openTestDB(t)
	repo, err := store.NewAnnotationRepository(db)
	if err != nil {
		t.Fatalf("NewAnnotationRepository: %v", err)
	}
	return repo, func() { db.Close() }
}

func TestAddAndListAnnotations(t *testing.T) {
	repo, cleanup := openAnnotationTestDB(t)
	defer cleanup()

	const runID = int64(42)

	if err := repo.Add(runID, "first note"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := repo.Add(runID, "second note"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	anns, err := repo.ListByRun(runID)
	if err != nil {
		t.Fatalf("ListByRun: %v", err)
	}
	if len(anns) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(anns))
	}
	if anns[0].Note != "first note" {
		t.Errorf("expected 'first note', got %q", anns[0].Note)
	}
	if anns[1].Note != "second note" {
		t.Errorf("expected 'second note', got %q", anns[1].Note)
	}
}

func TestAnnotationsListEmpty(t *testing.T) {
	repo, cleanup := openAnnotationTestDB(t)
	defer cleanup()

	anns, err := repo.ListByRun(999)
	if err != nil {
		t.Fatalf("ListByRun: %v", err)
	}
	if len(anns) != 0 {
		t.Errorf("expected empty slice, got %d items", len(anns))
	}
}

func TestDeleteAnnotation(t *testing.T) {
	repo, cleanup := openAnnotationTestDB(t)
	defer cleanup()

	const runID = int64(7)
	if err := repo.Add(runID, "to be deleted"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	anns, _ := repo.ListByRun(runID)
	if len(anns) != 1 {
		t.Fatalf("expected 1 annotation before delete, got %d", len(anns))
	}

	if err := repo.Delete(anns[0].ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	anns, _ = repo.ListByRun(runID)
	if len(anns) != 0 {
		t.Errorf("expected 0 annotations after delete, got %d", len(anns))
	}
}
