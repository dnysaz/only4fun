package handler

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestCreateProjectRejectsSameRepo memastikan project baru dari repo yang sudah
// dipakai project lain gagal dengan 409 Conflict, bukan 201 Created.
func TestCreateProjectRejectsSameRepo(t *testing.T) {
	store := newTestStore(t)
	handler := newTestHandler(t, store)

	userID := store.createUser(t, "alice")
	store.createProject(t, userID, "https://github.com/alice/api.git", "old-site", nil)

	rec := callCreateProject(t, handler, userID,
		`{"repo_url": "https://github.com/alice/api.git", "subdomain": "third-site", "env_vars": {}}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("Create project repo yang sama = %d, ingin 409 Conflict", rec.Code)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("tidak bisa baca body response: %v", err)
	}
	if body.Error == "" {
		t.Error("body error seharusnya berisi pesan repository sudah di-deploy")
	}
}

// TestCreateProjectRejectsSameRepoIgnoresCase mengecek normalisasi case URL.
func TestCreateProjectRejectsSameRepoIgnoresCase(t *testing.T) {
	store := newTestStore(t)
	handler := newTestHandler(t, store)

	userID := store.createUser(t, "alice")
	store.createProject(t, userID, "https://github.com/alice/api.git", "old-site", nil)

	rec := callCreateProject(t, handler, userID,
		`{"repo_url": "HTTPS://GITHUB.COM/ALICE/API.GIT", "subdomain": "new-site", "env_vars": {}}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("Create project repo case-insensitive = %d, ingin 409 Conflict", rec.Code)
	}
}

// TestCreateProjectAllowsDifferentRepo mengecek repo berbeda tetap bisa create.
func TestCreateProjectAllowsDifferentRepo(t *testing.T) {
	store := newTestStore(t)
	handler := newTestHandler(t, store)

	userID := store.createUser(t, "alice")
	store.createProject(t, userID, "https://github.com/alice/api.git", "site-one", nil)

	rec := callCreateProject(t, handler, userID,
		`{"repo_url": "https://github.com/alice/docs.git", "subdomain": "site-docs", "env_vars": {}}`)

	if rec.Code != http.StatusCreated {
		t.Errorf("Create project repo berbeda = %d, ingin 201 Created: %s", rec.Code, rec.Body.String())
	}
}

// TestCreateProjectRejectsTakenSubdomain memastikan nama project unik.
func TestCreateProjectRejectsTakenSubdomain(t *testing.T) {
	store := newTestStore(t)
	handler := newTestHandler(t, store)

	userID := store.createUser(t, "alice")
	store.createProject(t, userID, "https://github.com/alice/api.git", "my-site", nil)

	rec := callCreateProject(t, handler, userID,
		`{"repo_url": "https://github.com/alice/other.git", "subdomain": "my-site", "env_vars": {}}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("Create project subdomain terpakai = %d, ingin 409 Conflict", rec.Code)
	}
}

// TestCreateProjectRejectsInvalidPayload mengecek validasi input.
func TestCreateProjectRejectsInvalidPayload(t *testing.T) {
	store := newTestStore(t)
	handler := newTestHandler(t, store)

	userID := store.createUser(t, "alice")

	cases := []struct {
		name string
		body string
	}{
		{"repo bukan URL", `{"repo_url": "not-a-url", "subdomain": "valid-name", "env_vars": {}}`},
		{"nama terlalu pendek", `{"repo_url": "https://github.com/a/b.git", "subdomain": "ab", "env_vars": {}}`},
		{"nama pakai spasi", `{"repo_url": "https://github.com/a/b.git", "subdomain": "bad name", "env_vars": {}}`},
		{"env key tidak valid", `{"repo_url": "https://github.com/a/b.git", "subdomain": "valid-name", "env_vars": {"1BAD": "x"}}`},
		{"payload rusak", `{`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := callCreateProject(t, handler, userID, tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("Create project %q = %d, ingin 400 Bad Request", tc.name, rec.Code)
			}
		})
	}
}

// TestCreateProjectRequiresAuth memastikan user belum login ditolak.
func TestCreateProjectRequiresAuth(t *testing.T) {
	store := newTestStore(t)
	handler := newTestHandler(t, store)

	rec := callCreateProject(t, handler, "",
		`{"repo_url": "https://github.com/a/b.git", "subdomain": "valid-name", "env_vars": {}}`)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Create project tanpa login = %d, ingin 401", rec.Code)
	}
}
