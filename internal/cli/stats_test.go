package cli_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/crontrace/internal/cli"
	"github.com/user/crontrace/internal/store"
)

func TestPrintStatsEmpty(t *testing.T) {
	db := openTestDB(t)
	_ = db // ensure schema is created

	var buf bytes.Buffer
	// Use the internal helper via a temp path already opened by openTestDB.
	// We call the exported path-based function indirectly by injecting via
	// a shared test helper that mirrors printStats logic.
	if err := cli.PrintStatsFromDB(db, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No job runs") {
		t.Errorf("expected empty message, got: %q", buf.String())
	}
}

func TestPrintStatsWithData(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewJobRunRepository(db)

	now := time.Now().UTC()
	end := now.Add(5 * time.Second)
	for _, code := range []int{0, 0, 1} {
		if err := repo.Insert(&store.JobRun{
			Command:   "backup.sh",
			StartedAt: now,
			EndedAt:   &end,
			ExitCode:  code,
		}); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	var buf bytes.Buffer
	if err := cli.PrintStatsFromDB(db, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "backup.sh") {
		t.Errorf("expected command in output, got: %q", out)
	}
	if !strings.Contains(out, "3") {
		t.Errorf("expected total count 3 in output, got: %q", out)
	}
	if !strings.Contains(out, "COMMAND") {
		t.Errorf("expected header row in output, got: %q", out)
	}
}
