package database

import (
	"strings"
	"testing"
)

func TestAuditStoreRecordsAndClipsSensitiveFields(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	store := NewAuditStore(db)

	user, err := NewUserStore(db).Create("gh-audit", "audit-user", "", "")
	if err != nil {
		t.Fatal(err)
	}
	longTarget := strings.Repeat("x", 400)
	if err := store.Record(user.ID, "project.env.update", longTarget, "203.0.113.9", 200); err != nil {
		t.Fatalf("Record: %v", err)
	}

	logs, err := store.ListByUser(user.ID, 10)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	got := logs[0]
	if got.Action != "project.env.update" || got.IP != "203.0.113.9" || got.Status != 200 {
		t.Fatalf("unexpected audit entry: %+v", got)
	}
	if len(got.Target) != 256 {
		t.Fatalf("target length = %d, want 256", len(got.Target))
	}
	if len(got.Target) >= len(longTarget) {
		t.Fatal("audit log did not clip the long target")
	}
}

func TestAuditLogsCascadeWithUser(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	store := NewAuditStore(db)
	users := NewUserStore(db)

	user, err := users.Create("gh-cascade", "cascade-user", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Record(user.ID, "project.delete", "p1", "127.0.0.1", 200); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB().Exec(`DELETE FROM users WHERE id = ?`, user.ID); err != nil {
		t.Fatal(err)
	}

	var remaining int
	if err := db.DB().Get(&remaining, `SELECT count(*) FROM audit_logs WHERE id IS NOT NULL`); err != nil {
		t.Fatal(err)
	}
	if remaining != 1 {
		t.Fatalf("audit rows after user deletion = %d, want 1 (user_id becomes NULL)", remaining)
	}
}
