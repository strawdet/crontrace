package store_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/user/crontrace/internal/store"
)

func TestExportCSVEmpty(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewExportRepository(db)

	var buf bytes.Buffer
	if err := repo.WriteCSV(&buf, ""); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected header only, got %d rows", len(records))
	}
	if records[0][0] != "id" {
		t.Errorf("unexpected header: %v", records[0])
	}
}

func TestExportCSVWithData(t *testing.T) {
	db := openTestDB(t)
	jr := store.NewJobRunRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	_ = jr.Insert("backup.sh", now, &now, 0)
	_ = jr.Insert("cleanup.sh", now, &now, 1)

	repo := store.NewExportRepository(db)
	var buf bytes.Buffer
	if err := repo.WriteCSV(&buf, ""); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	r := csv.NewReader(&buf)
	records, _ := r.ReadAll()
	// header + 2 data rows
	if len(records) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(records))
	}
}

func TestExportCSVFilterByCommand(t *testing.T) {
	db := openTestDB(t)
	jr := store.NewJobRunRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	_ = jr.Insert("backup.sh", now, &now, 0)
	_ = jr.Insert("cleanup.sh", now, &now, 1)

	repo := store.NewExportRepository(db)
	var buf bytes.Buffer
	if err := repo.WriteCSV(&buf, "backup.sh"); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	r := csv.NewReader(&buf)
	records, _ := r.ReadAll()
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d", len(records))
	}
	if !strings.Contains(records[1][1], "backup.sh") {
		t.Errorf("expected backup.sh, got %s", records[1][1])
	}
}

func TestExportJSONWithData(t *testing.T) {
	db := openTestDB(t)
	jr := store.NewJobRunRepository(db)
	now := time.Now().UTC().Truncate(time.Second)
	_ = jr.Insert("sync.sh", now, &now, 0)

	repo := store.NewExportRepository(db)
	var buf bytes.Buffer
	if err := repo.WriteJSON(&buf, ""); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var rows []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
		t.Fatalf("parse json: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["command"] != "sync.sh" {
		t.Errorf("unexpected command: %v", rows[0]["command"])
	}
}
