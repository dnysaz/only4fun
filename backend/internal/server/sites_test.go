package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"only4.fun/backend/internal/config"
	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/secure"
)

func TestSSRProxyTimeouts(t *testing.T) {
	proxy := newSSRProxy(4123)
	transport, ok := proxy.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("proxy transport = %T, want *http.Transport", proxy.Transport)
	}
	if transport.DialContext == nil {
		t.Error("proxy has no dial timeout")
	}
	if transport.TLSHandshakeTimeout <= 0 || transport.ResponseHeaderTimeout <= 0 {
		t.Errorf("proxy timeouts are not configured: %+v", transport)
	}
	if proxy.FlushInterval >= 0 {
		t.Errorf("FlushInterval = %v, want negative (flush immediately) for streaming", proxy.FlushInterval)
	}
}

func TestServerEntryForProject(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "server.js"), []byte("console.log(1)"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Tanpa script "start", server.js tidak boleh dijalankan sebagai SSR.
	if got := findServerEntry(root); got != "" {
		t.Fatalf("findServerEntry without start script = %q, want empty", got)
	}

	// Project ZIP tidak pernah boleh masuk mode SSR walaupun ada server.js.
	if got := serverEntryForProject(root, "zip://site.zip"); got != "" {
		t.Fatalf("serverEntryForProject(zip) = %q, want empty", got)
	}

	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"scripts":{"start":"node server.js"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := findServerEntry(root); !strings.HasSuffix(got, "server.js") {
		t.Fatalf("findServerEntry with start script = %q, want server.js", got)
	}
	if got := serverEntryForProject(root, "https://github.com/demo/repo"); !strings.HasSuffix(got, "server.js") {
		t.Fatalf("serverEntryForProject(repo) = %q, want server.js", got)
	}
}

func TestFindServerEntryNextMarker(t *testing.T) {
	root := t.TempDir()
	writeSiteFile := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeSiteFile("server.js", "console.log(1)")
	// Tanpa start script dan tanpa penanda builder, server.js tetap diabaikan.
	if got := findServerEntry(root); got != "" {
		t.Fatalf("findServerEntry without marker = %q, want empty", got)
	}

	writeSiteFile(nextSSRMarker, "next-start\n")
	if got := findServerEntry(root); !strings.HasSuffix(got, "server.js") {
		t.Fatalf("findServerEntry with Next.js marker = %q, want server.js", got)
	}
}

func TestSiteHandlerFallsBackToReleasesInsideSitesDir(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := secure.Init("only4.fun-test-secret"); err != nil {
		t.Fatalf("init encryption: %v", err)
	}
	db, err := database.Init(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}

	sitesDir := t.TempDir()
	release := filepath.Join(sitesDir, ".releases", "ghost", "2026-11-27-000000")
	if err := os.MkdirAll(release, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(release, "index.html"), []byte("<h1>from releases</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Cloudflare: config.CloudflareConfig{Domain: "only4.fun"},
		Build:      config.BuildConfig{SitesDir: sitesDir},
	}
	handler := SiteHandler(cfg, database.NewProjectStore(db), nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "http://ghost.only4.fun/", nil)
	c.Request.Host = "ghost.only4.fun"
	handler(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "from releases") {
		t.Fatalf("body = %q, want release index.html", rec.Body.String())
	}
}

func TestSiteHandlerRejectsDotfilesButAllowsWellKnown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := secure.Init("only4.fun-test-secret"); err != nil {
		t.Fatal(err)
	}
	db, err := database.Init(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}

	sitesDir := t.TempDir()
	root := filepath.Join(sitesDir, "dots")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<h1>ok</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("SECRET=leak"), 0o644); err != nil {
		t.Fatal(err)
	}
	wellKnown := filepath.Join(root, ".well-known")
	if err := os.MkdirAll(wellKnown, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wellKnown, "hello.txt"), []byte("verified"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Cloudflare: config.CloudflareConfig{Domain: "only4.fun"},
		Build:      config.BuildConfig{SitesDir: sitesDir},
	}
	handler := SiteHandler(cfg, database.NewProjectStore(db), nil)

	call := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "http://dots.only4.fun"+path, nil)
		c.Request.Host = "dots.only4.fun"
		handler(c)
		return rec
	}

	if rec := call("/"); rec.Code != http.StatusOK {
		t.Fatalf("index status = %d, want 200", rec.Code)
	}
	for _, path := range []string{"/.env", "/.git/config", "/.releases/dots/old/index.html"} {
		rec := call(path)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s status = %d, want 404; body = %s", path, rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "SECRET=leak") {
			t.Errorf("%s leaked the .env file", path)
		}
	}
	if rec := call("/.well-known/hello.txt"); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "verified") {
		t.Errorf(".well-known status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestSiteHandlerServesCustomDomain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := secure.Init("only4.fun-test-secret"); err != nil {
		t.Fatalf("init encryption: %v", err)
	}

	db, err := database.Init(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("init database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	projects := database.NewProjectStore(db)
	user, err := database.NewUserStore(db).Create("gh-domain-owner", "domain-owner", "owner@example.com", "")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	project, err := projects.Create(user.ID, "https://github.com/demo/fun-site", "fun-site", nil)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if err := projects.UpdateCustomDomain(project.ID, "www.example.com"); err != nil {
		t.Fatalf("save custom domain: %v", err)
	}
	if err := projects.UpdateStatus(project.ID, "deployed"); err != nil {
		t.Fatalf("publish project: %v", err)
	}

	sitesDir := t.TempDir()
	root := filepath.Join(sitesDir, project.Subdomain)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("create site dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<h1>custom domain works</h1>"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	cfg := &config.Config{
		Cloudflare: config.CloudflareConfig{Domain: "only4.fun"},
		Build:      config.BuildConfig{SitesDir: sitesDir},
	}
	handler := SiteHandler(cfg, projects, nil)

	for _, host := range []string{"www.example.com", "fun-site.only4.fun"} {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "http://"+host+"/", nil)
		c.Request.Host = host
		handler(c)

		if rec.Code != http.StatusOK {
			t.Errorf("host %s status = %d, want %d; body = %s", host, rec.Code, http.StatusOK, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "custom domain works") {
			t.Errorf("host %s body = %q, want published site", host, rec.Body.String())
		}
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "http://unknown.example.com/", nil)
	c.Request.Host = "unknown.example.com"
	handler(c)
	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown host status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestSiteHandlerNotFoundPageIncludesBrandedLayout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := secure.Init("only4.fun-test-secret"); err != nil {
		t.Fatalf("init encryption: %v", err)
	}
	db, err := database.Init(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Cloudflare: config.CloudflareConfig{Domain: "only4.fun"},
		Server:     config.ServerConfig{FrontendURL: "https://only4.fun/"},
		Build:      config.BuildConfig{SitesDir: t.TempDir()},
	}
	handler := SiteHandler(cfg, database.NewProjectStore(db), nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "http://missing.only4.fun/", nil)
	c.Request.Host = "missing.only4.fun"
	handler(c)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		"This site",
		"isn’t here.",
		"custom domain changed",
		"missing.only4.fun",
		"nav-arrow",
		"Start a new project",
		"Only4",
		"Push, build, live.",
		"cluster-spin",
		"https://only4.fun/deploy",
		"classList.add('scrolled')",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
	// Navbar, hero, dan footer masing-masing memakai empat bulatan warna warni.
	if got := strings.Count(body, "<i></i>"); got < 12 {
		t.Errorf("colored dots = %d, want at least 12", got)
	}
}
