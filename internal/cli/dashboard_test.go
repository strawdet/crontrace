package cli_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/crontrace/internal/cli"
	"github.com/user/crontrace/internal/store"
)

func TestPrintDashboardEmpty(t *testing.T) {
	db := openTestDB(t)
	repo := store.NewDashboardRepository(db)

	var buf bytes.Buffer
	cli.PrintDashboardWriter(&buf, repo)

	out := buf.String()
	if !strings.Contains(out, "Total runs      : 0") {
		t.Errorf("expected zero total runs, got: %s", out)
	}
	if !strings.Contains(out, "Last run at     : n/a") {
		t.Errorf("expected n/a last run, got: %s", out)
	}
}

func TestPrintDashboardWithData(t *testing.T) {
	db := openTestDB(t)
	jobRepo := store.NewJobRunRepository(db)
	dashRepo := store.NewDashboardRepository(db)

	now := time.Now().UTC()
	runs := []store.JobRun{
		{Command: "nightly.sh", StartedAt: now.Add(-3 * time.Hour), FinishedAt: now.Add(-3*time.Hour + 60*time.Second), DurationMs: 60000, ExitCode: 0},
		{Command: "nightly.sh", StartedAt: now.Add(-1 * time.Hour), FinishedAt: now.Add(-1*time.Hour + 45*time.Second), DurationMs: 45000, ExitCode: 2},
	}
	for _, r := range runs {
		if err := jobRepo.Insert(r); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	var buf bytes.Buffer
	cli.PrintDashboardWriter(&buf, dashRepo)

	out := buf.String()
	if !strings.Contains(out, "Total runs      : 2") {
		t.Errorf("expected 2 total runs, got: %s", out)
	}
	if !strings.Contains(out, "Successful      : 1") {
		t.Errorf("expected 1 success, got: %s", out)
	}
	if !strings.Contains(out, "Failed          : 1") {
		t.Errorf("expected 1 failure, got: %s", out)
	}
	if !strings.Contains(out, "Unique commands : 1") {
		t.Errorf("expected 1 unique command, got: %s", out)
	}
	if strings.Contains(out, "Last run at     : n/a") {
		t.Errorf("expected real last run time, got n/a")
	}
}
