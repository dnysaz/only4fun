package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"

	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

func TestRepoTreeUsesSessionToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var tokenOK int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer gh-token" {
			atomic.AddInt32(&tokenOK, 1)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/acme/private-site":
			fmt.Fprint(w, `{"default_branch":"main"}`)
		case "/repos/acme/private-site/git/trees/main":
			if r.URL.Query().Get("recursive") != "1" {
				t.Errorf("recursive query missing: %s", r.URL.RawQuery)
			}
			fmt.Fprint(w, `{"truncated":false,"tree":[
				{"path":"index.html","type":"blob"},
				{"path":"src/App.jsx","type":"blob"},
				{"path":"src","type":"tree"},
				{"path":"../escape.txt","type":"blob"}
			]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"not found"}`)
		}
	}))
	defer upstream.Close()

	store := newTestStore(t)
	handler, sessionID := newTestGitHubHandler(t, store)
	handler.apiBase = upstream.URL

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{
		{Key: "owner", Value: "acme"},
		{Key: "repo", Value: "private-site"},
	}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/github/repos/acme/private-site/tree", nil)
	c.Request.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	handler.GetRepoTree(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetRepoTree = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if got := atomic.LoadInt32(&tokenOK); got != 2 {
		t.Fatalf("requests with session token = %d, want 2", got)
	}
	var got struct {
		Paths     []string `json:"paths"`
		Truncated bool     `json:"truncated"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := []string{"index.html", "src/App.jsx"}
	if !reflect.DeepEqual(got.Paths, want) {
		t.Fatalf("paths = %v, want %v", got.Paths, want)
	}
	if got.Truncated {
		t.Fatal("truncated should be false")
	}
}

func TestRepoTreeRequiresSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newTestStore(t)
	handler := NewGitHubHandler(database.NewSessionStore(store.db))
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/github/repos/acme/private-site/tree", nil)
	handler.GetRepoTree(c)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GetRepoTree without session = %d, want 401", rec.Code)
	}
}

func TestRepoTreeExplainsGitHubForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"forbidden"}`)
	}))
	defer upstream.Close()

	store := newTestStore(t)
	handler, sessionID := newTestGitHubHandler(t, store)
	handler.apiBase = upstream.URL

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{
		{Key: "owner", Value: "acme"},
		{Key: "repo", Value: "private-site"},
	}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/github/repos/acme/private-site/tree", nil)
	c.Request.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	handler.GetRepoTree(c)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("GetRepoTree with GitHub 403 = %d, want 502", rec.Code)
	}
	var got struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Error) < 10 {
		t.Fatalf("expected a clear error message, got %q", got.Error)
	}
}
