package handler

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/secure"
)

// testStore membungkus database sementara (SQLite di folder temp) beserta
// store yang dipakai handler, supaya test tidak perlu menyalakan server penuh.
type testStore struct {
	db       *database.Database
	projects *database.ProjectStore
	users    *database.UserStore
}

func newTestStore(t *testing.T) *testStore {
	t.Helper()
	gin.SetMode(gin.TestMode)

	if err := secure.Init("only4.fun-test-secret"); err != nil {
		t.Fatalf("init encryption: %v", err)
	}

	db, err := database.Init(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("init test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return &testStore{
		db:       db,
		projects: database.NewProjectStore(db),
		users:    database.NewUserStore(db),
	}
}

func (s *testStore) createUser(t *testing.T, username string) string {
	t.Helper()
	user, err := s.users.Create(
		"gh-"+username,
		username,
		username+"@example.com",
		"",
	)
	if err != nil {
		t.Fatalf("create user %q: %v", username, err)
	}
	return user.ID
}

func (s *testStore) createProject(t *testing.T, userID, repoURL, subdomain string, envVars map[string]string) string {
	t.Helper()
	if envVars == nil {
		envVars = map[string]string{}
	}
	project, err := s.projects.Create(userID, repoURL, subdomain, envVars)
	if err != nil {
		t.Fatalf("create project %q: %v", subdomain, err)
	}
	return project.ID
}

// newTestHandler menyusun ProjectHandler dengan store asli dari testStore.
// Worker, DNS, dan process manager sengaja nil karena tidak dipakai test ini.
func newTestHandler(t *testing.T, store *testStore) *ProjectHandler {
	t.Helper()
	return NewProjectHandler(
		store.projects,
		store.users,
		nil, // sessionStore
		nil, // deploymentStore
		nil, // worker
		nil, // dnsClient
		nil, // processes
		"",  // sitesDir
	)
}

// callCreateProject menjalankan handler CreateProject seperti request HTTP biasa.
func callCreateProject(t *testing.T, h *ProjectHandler, userID, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", userID)
	h.CreateProject(c)
	return rec
}
