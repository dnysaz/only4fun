package database

import (
	"testing"
)

// TestGetByRepoURLCaseInsensitive memastikan pencarian project per repo URL
// case-insensitive, sesuai konsistensi handler.
func TestGetByRepoURLCaseInsensitive(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	store := NewProjectStore(db)
	userID := "user-uuid"

	// Buat dua project, URL yang sama tapi dengan casing berbeda.
	store.Create(userID, "https://github.com/alice/API.git", "upper-api", nil)
	store.Create(userID, "https://github.com/alice/api.git", "lower-api", nil)

	got, err := store.GetByRepoURL("https://github.com/alice/api.git")
	if err != nil {
		t.Fatalf("GetByRepoURL error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("GetByRepoURL returns %d project, ingin 2", len(got))
	}

	gotLower, err := store.GetByRepoURL("https://github.com/alice/API.git")
	if err != nil {
		t.Fatalf("GetByRepoURL case-insensitive error: %v", err)
	}
	if len(gotLower) != 2 {
		t.Fatalf("GetByRepoURL case-insensitive returns %d project, ingin 2", len(gotLower))
	}
}
