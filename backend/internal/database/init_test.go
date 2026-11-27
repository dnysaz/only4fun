package database

import "testing"

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
