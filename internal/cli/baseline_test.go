package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/crontrace/internal/store"
)

func newBaselineRepo(t *testing.T) *store.BaselineRepository {
	t.Helper()
	db := openTestDB(t)
	repo, err := store.NewBaselineRepository(db)
	if err != nil {
		t.Fatalf("new baseline repo: %v", err)
	}
	return repo
}

func TestBaselineSetAndList(t *testing.T) {
	repo := newBaselineRepo(t)
	var buf bytes.Buffer

	manageBaseline(repo, []string{"set", "deploy.sh", "3500", "7"}, &buf)
	if !strings.Contains(buf.String(), "deploy.sh") {
		t.Errorf("expected deploy.sh in output, got: %s", buf.String())
	}

	buf.Reset()
	manageBaseline(repo, []string{"list"}, &buf)
	if !strings.Contains(buf.String(), "deploy.sh") {
		t.Errorf("expected deploy.sh in list, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "3500") {
		t.Errorf("expected 3500 in list, got: %s", buf.String())
	}
}

func TestBaselineSetInvalidMs(t *testing.T) {
	repo := newBaselineRepo(t)
	var buf bytes.Buffer
	manageBaseline(repo, []string{"set", "deploy.sh", "notanumber"}, &buf)
	if !strings.Contains(buf.String(), "invalid avg_ms") {
		t.Errorf("expected error message, got: %s", buf.String())
	}
}

func TestBaselineListEmpty(t *testing.T) {
	repo := newBaselineRepo(t)
	var buf bytes.Buffer
	manageBaseline(repo, []string{"list"}, &buf)
	if !strings.Contains(buf.String(), "no baselines") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}

func TestBaselineDel(t *testing.T) {
	repo := newBaselineRepo(t)
	var buf bytes.Buffer
	manageBaseline(repo, []string{"set", "clean.sh", "1000", "3"}, &buf)
	buf.Reset()
	manageBaseline(repo, []string{"del", "clean.sh"}, &buf)
	if !strings.Contains(buf.String(), "deleted") {
		t.Errorf("expected deleted message, got: %s", buf.String())
	}
	buf.Reset()
	manageBaseline(repo, []string{"list"}, &buf)
	if strings.Contains(buf.String(), "clean.sh") {
		t.Errorf("clean.sh should be gone, got: %s", buf.String())
	}
}

func TestBaselineUnknownSubcommand(t *testing.T) {
	repo := newBaselineRepo(t)
	var buf bytes.Buffer
	manageBaseline(repo, []string{"bogus"}, &buf)
	if !strings.Contains(buf.String(), "unknown") {
		t.Errorf("expected unknown sub-command message, got: %s", buf.String())
	}
}
