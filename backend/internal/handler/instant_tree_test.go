package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInstantGithubOwnerRepo(t *testing.T) {
	owner, repo, err := instantGithubOwnerRepo("https://github.com/acme/static-site.git")
	if err != nil || owner != "acme" || repo != "static-site" {
		t.Fatalf("instantGithubOwnerRepo = %q, %q, %v", owner, repo, err)
	}

	for _, raw := range []string{
		"https://evil.example/acme/static-site",
		"https://github.com/acme",
		"https://user:pass@github.com/acme/static-site",
		"https://github.com/acme/../static-site",
	} {
		if _, _, err := instantGithubOwnerRepo(raw); err == nil {
			t.Fatalf("instantGithubOwnerRepo(%q) should fail", raw)
		}
	}
}

func TestCleanInstantTreePath(t *testing.T) {
	cases := map[string]string{
		"index.html":           "index.html",
		"./css/style.css":      "css/style.css",
		"src\\app\\main.js":    "src/app/main.js",
		"/etc/passwd":          "",
		"src/../../secret.txt": "",
		"":                     "",
	}
	for input, want := range cases {
		if got := cleanInstantTreePath(input); got != want {
			t.Fatalf("cleanInstantTreePath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestRepoTreeReturnsPublicFileList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/acme/static-site":
			fmt.Fprint(w, `{"default_branch":"main"}`)
		case "/repos/acme/static-site/git/trees/main":
			if r.URL.Query().Get("recursive") != "1" {
				t.Errorf("recursive query missing: %s", r.URL.RawQuery)
			}
			fmt.Fprint(w, `{"truncated":false,"tree":[
				{"path":"index.html","type":"blob"},
				{"path":"css/style.css","type":"blob"},
				{"path":"src","type":"tree"},
				{"path":"src/App.jsx","type":"blob"},
				{"path":"../escape.txt","type":"blob"}
			]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"not found"}`)
		}
	}))
	defer srv.Close()

	h := &InstantHandler{githubAPI: srv.URL, httpClient: srv.Client()}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/instant/repo-tree", strings.NewReader(`{"repo_url":"https://github.com/acme/static-site"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.RepoTree(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	var got struct {
		Paths     []string `json:"paths"`
		Truncated bool     `json:"truncated"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := []string{"index.html", "css/style.css", "src/App.jsx"}
	if !reflect.DeepEqual(got.Paths, want) {
		t.Fatalf("paths = %v, want %v", got.Paths, want)
	}
	if got.Truncated {
		t.Fatal("truncated should be false")
	}
}
