package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

// newTestGitHubHandler menyiapkan handler dengan session nyata dan API GitHub
// palsu, supaya pagination bisa diuji tanpa jaringan.
func newTestGitHubHandler(t *testing.T, store *testStore) (*GitHubHandler, string) {
	t.Helper()

	sessionStore := database.NewSessionStore(store.db)
	userID := store.createUser(t, "alice")
	// Session disimpan sebagai hash, jadi token asli harus dipakai di request.
	if _, err := sessionStore.Create(userID, "test-session-id", "gh-token", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create session: %v", err)
	}

	return NewGitHubHandler(sessionStore), "test-session-id"
}

// callListRepos menjalankan ListRepos dengan session di cookie seperti browser.
func callListRepos(t *testing.T, h *GitHubHandler, sessionID, query string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/github/repos"+query, nil)
	c.Request.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	c.Set("userID", "user-id")
	h.ListRepos(c)
	return rec
}

func TestListReposPagination(t *testing.T) {
	var requestedPage, requestedPerPage string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPage = r.URL.Query().Get("page")
		requestedPerPage = r.URL.Query().Get("per_page")
		if r.Header.Get("Authorization") != "Bearer gh-token" {
			t.Errorf("Authorization = %q, ingin memakai token session", r.Header.Get("Authorization"))
		}

		// Halaman 1 penuh (2 repo), halaman 2 hanya 1 repo → has_more false.
		count := 2
		if requestedPage == "2" {
			count = 1
		}
		repos := make([]map[string]any, 0, count)
		for i := 0; i < count; i++ {
			repos = append(repos, map[string]any{
				"name":      fmt.Sprintf("repo-%d", i),
				"full_name": fmt.Sprintf("alice/repo-%d", i),
				"html_url":  fmt.Sprintf("https://github.com/alice/repo-%d", i),
				"private":   i == 0,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(repos)
	}))
	defer upstream.Close()

	store := newTestStore(t)
	handler, sessionID := newTestGitHubHandler(t, store)
	handler.apiBase = upstream.URL

	// per_page dibatasi maksimum 100 walaupun klien meminta lebih.
	rec := callListRepos(t, handler, sessionID, "?page=1&per_page=500")
	if rec.Code != http.StatusOK {
		t.Fatalf("ListRepos = %d, ingin 200: %s", rec.Code, rec.Body.String())
	}
	if requestedPerPage != "100" {
		t.Errorf("per_page diteruskan = %q, ingin 100", requestedPerPage)
	}
	var first struct {
		Repos   []struct{ FullName string } `json:"repos"`
		Page    int                         `json:"page"`
		HasMore bool                        `json:"has_more"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &first); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if requestedPage != "1" || first.Page != 1 {
		t.Errorf("page = %q/%d, ingin 1", requestedPage, first.Page)
	}
	// per_page=100 tetapi upstream hanya mengirim 2, jadi has_more false.
	if first.HasMore {
		t.Errorf("has_more = true, ingin false karena halaman tidak penuh")
	}

	// Halaman berikutnya diteruskan apa adanya.
	rec = callListRepos(t, handler, sessionID, "?page=2")
	if rec.Code != http.StatusOK {
		t.Fatalf("ListRepos halaman 2 = %d, ingin 200", rec.Code)
	}
	if requestedPage != "2" {
		t.Errorf("page diteruskan = %q, ingin 2", requestedPage)
	}

	// page tidak valid jatuh ke 1, bukan error/panic.
	rec = callListRepos(t, handler, sessionID, "?page=abc")
	if rec.Code != http.StatusOK || requestedPage != "1" {
		t.Errorf("page=abc → status %d, page %q; ingin 200 dan page 1", rec.Code, requestedPage)
	}
}

func TestListReposRequiresSession(t *testing.T) {
	store := newTestStore(t)
	gin.SetMode(gin.TestMode)

	handler := NewGitHubHandler(database.NewSessionStore(store.db))
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/github/repos", nil)
	handler.ListRepos(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("ListRepos tanpa session = %d, ingin 401", rec.Code)
	}
}
