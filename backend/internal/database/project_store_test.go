package database

import (
	"testing"
)

// TestGetByRepoURLCaseInsensitive memastikan pencarian project per repo URL
// case-insensitive, sesuai konsistensi handler.
func TestCustomDomainUniqueIsCaseInsensitive(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	store := NewProjectStore(db)
	userID := "user-uuid"
	first, err := store.Create(userID, "https://github.com/alice/one", "one-site", nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Create(userID, "https://github.com/alice/two", "two-site", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.UpdateCustomDomain(first.ID, "www.example.com"); err != nil {
		t.Fatalf("save first domain: %v", err)
	}
	if err := store.UpdateCustomDomain(second.ID, "WWW.EXAMPLE.COM"); err == nil {
		t.Fatal("case-only duplicate custom domain must be rejected")
	}

	got, err := store.GetByCustomDomain("www.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != first.ID {
		t.Fatalf("GetByCustomDomain returned %q, want %q", got.ID, first.ID)
	}
}

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
