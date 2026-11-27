package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"only4.fun/backend/internal/database"
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
		{"env PATH ditolak", `{"repo_url": "https://github.com/a/b.git", "subdomain": "valid-name", "env_vars": {"PATH": "/tmp/evil"}}`},
		{"env LD_PRELOAD ditolak", `{"repo_url": "https://github.com/a/b.git", "subdomain": "valid-name", "env_vars": {"LD_PRELOAD": "/tmp/evil.so"}}`},
		{"env NODE_OPTIONS ditolak", `{"repo_url": "https://github.com/a/b.git", "subdomain": "valid-name", "env_vars": {"NODE_OPTIONS": "--require /tmp/evil.js"}}`},
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

// Kata terlarang dicek dengan exact match, bukan substring.
func TestValidateSubdomainBlockedWordsExactMatch(t *testing.T) {
	blocked := []string{"admin", "porn", "anjing", "fuck", "root"}
	for _, name := range blocked {
		if err := validateSubdomain(name); err == nil {
			t.Errorf("validateSubdomain(%q) ingin error, dapat nil", name)
		}
	}

	allowed := []string{"madmin", "adminku", "my-porn-site", "anjingku", "fun-site", "rooter"}
	for _, name := range allowed {
		if err := validateSubdomain(name); err != nil {
			t.Errorf("validateSubdomain(%q) ingin boleh, dapat %v", name, err)
		}
	}
}

func TestRedeployRejectsAlreadyRunningProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newTestStore(t)
	userID := store.createUser(t, "alice")
	projectID := store.createProject(t, userID, "https://github.com/alice/site", "busy-site", nil)
	if err := store.projects.UpdateStatus(projectID, "building"); err != nil {
		t.Fatal(err)
	}
	handler := newTestHandler(t, store)

	for _, status := range []string{"building", "pending"} {
		if err := store.projects.UpdateStatus(projectID, status); err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID+"/deploy", nil)
		c.Params = gin.Params{{Key: "id", Value: projectID}}
		c.Set("userID", userID)
		handler.Redeploy(c)
		if rec.Code != http.StatusConflict {
			t.Errorf("Redeploy on %s = %d, want 409; body = %s", status, rec.Code, rec.Body.String())
		}
	}
}

func TestClearDeploymentsCleansUnusedReleases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newTestStore(t)
	userID := store.createUser(t, "alice")
	projectID := store.createProject(t, userID, "https://github.com/alice/site", "clear-site", nil)
	deployments := database.NewDeploymentStore(store.db)
	dep, err := deployments.Create(projectID, "success", "build ok")
	if err != nil {
		t.Fatal(err)
	}
	if err := deployments.SetRelease(dep.ID, "old-release"); err != nil {
		t.Fatal(err)
	}

	sitesDir := t.TempDir()
	releasesDir := filepath.Join(sitesDir, ".releases", "clear-site")
	for _, name := range []string{"old-release", "active-release"} {
		dir := filepath.Join(releasesDir, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	target := filepath.Join(sitesDir, "clear-site")
	if err := os.Symlink(filepath.Join(releasesDir, "active-release"), target); err != nil {
		t.Fatal(err)
	}

	handler := NewProjectHandler(store.projects, store.users, nil, deployments, nil, nil, nil, sitesDir, "")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/projects/"+projectID+"/deployments", strings.NewReader(""))
	c.Params = gin.Params{{Key: "id", Value: projectID}}
	c.Set("userID", userID)
	handler.ClearDeployments(c)
	if rec.Code != http.StatusOK {
		t.Fatalf("ClearDeployments = %d: %s", rec.Code, rec.Body.String())
	}

	if list, err := deployments.GetByProjectID(projectID); err != nil || len(list) != 0 {
		t.Fatalf("history should be empty: list=%d err=%v", len(list), err)
	}
	if _, err := os.Stat(filepath.Join(releasesDir, "old-release")); !os.IsNotExist(err) {
		t.Fatalf("unused release still exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(releasesDir, "active-release", "index.html")); err != nil {
		t.Fatalf("active release was removed: %v", err)
	}
}
