package database

import (
	"strings"
	"testing"
)

func TestWALAndBusyTimeoutAreEnabled(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	var mode string
	if err := db.DB().Get(&mode, "PRAGMA journal_mode"); err != nil {
		t.Fatalf("read journal_mode: %v", err)
	}
	if !strings.EqualFold(mode, "wal") {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}
	var timeout int
	if err := db.DB().Get(&timeout, "PRAGMA busy_timeout"); err != nil {
		t.Fatalf("read busy_timeout: %v", err)
	}
	if timeout != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", timeout)
	}
}

func TestDeploymentHistoryIndexExists(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	var count int
	if err := db.DB().Get(&count, `SELECT count(*) FROM sqlite_master WHERE type = 'index' AND name = 'idx_deployments_project_created'`); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("deployment history index count = %d, want 1", count)
	}
}

func TestForeignKeysAreEnabled(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	var enabled int
	if err := db.DB().Get(&enabled, "PRAGMA foreign_keys"); err != nil {
		t.Fatalf("failed to read foreign_keys pragma: %v", err)
	}
	if enabled != 1 {
		t.Fatalf("foreign_keys = %d, want 1", enabled)
	}

	_, err := db.DB().Exec(`INSERT INTO projects (id, user_id, repo_url, subdomain) VALUES (?, ?, ?, ?)`,
		"orphan-project", "missing-user", "https://github.com/alice/example", "orphan-site")
	if err == nil {
		t.Fatal("project with a missing user was accepted; foreign keys are not enforced")
	}
}
