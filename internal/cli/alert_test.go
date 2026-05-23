package cli_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/user/crontrace/internal/cli"
)

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestManageAlertsSet(t *testing.T) {
	db := openTestDB(t)
	err := cli.ManageAlerts(db, []string{"set", "/bin/backup", "duration", "60"})
	if err != nil {
		t.Fatalf("ManageAlerts set: %v", err)
	}
}

func TestManageAlertsSetInvalidMetric(t *testing.T) {
	db := openTestDB(t)
	err := cli.ManageAlerts(db, []string{"set", "/bin/backup", "latency", "60"})
	if err == nil {
		t.Fatal("expected error for invalid metric")
	}
}

func TestManageAlertsSetInvalidThreshold(t *testing.T) {
	db := openTestDB(t)
	err := cli.ManageAlerts(db, []string{"set", "/bin/backup", "duration", "abc"})
	if err == nil {
		t.Fatal("expected error for non-numeric threshold")
	}
}

func TestManageAlertsDel(t *testing.T) {
	db := openTestDB(t)
	_ = cli.ManageAlerts(db, []string{"set", "/bin/clean", "exit_code", "1"})
	err := cli.ManageAlerts(db, []string{"del", "/bin/clean", "exit_code"})
	if err != nil {
		t.Fatalf("ManageAlerts del: %v", err)
	}
}

func TestManageAlertsListEmpty(t *testing.T) {
	db := openTestDB(t)
	out := captureOutput(t, func() {
		_ = cli.ManageAlerts(db, []string{"list"})
	})
	if out == "" {
		t.Error("expected some output for empty list")
	}
}

func TestManageAlertsListWithData(t *testing.T) {
	db := openTestDB(t)
	_ = cli.ManageAlerts(db, []string{"set", "/bin/job", "duration", "45"})
	out := captureOutput(t, func() {
		_ = cli.ManageAlerts(db, []string{"list", "/bin/job"})
	})
	if out == "" {
		t.Error("expected output listing alert")
	}
}

func TestManageAlertsUnknownSubcommand(t *testing.T) {
	db := openTestDB(t)
	err := cli.ManageAlerts(db, []string{"bogus"})
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
}
