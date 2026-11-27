package server

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"only4.fun/backend/internal/config"
	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

// subdomainLabel memastikan nama subdomain yang dipakai untuk lookup dan
// path folder hanya berisi karakter DNS yang sah (mencegah path traversal).
var subdomainLabel = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// nextSSRMarker ditulis builder bersama wrapper server.js Next.js.
const nextSSRMarker = ".only4-next-ssr"

// SiteHandler melayani hasil build berdasarkan Host request,
// misal portfolio.only4.fun -> <SITES_DIR>/portfolio/index.html
// Jika project memiliki server entry point (SSR), request akan di-proxy
// ke proses Node.js yang berjalan.
func SiteHandler(cfg *config.Config, projectStore *database.ProjectStore, pm *ProcessManager) gin.HandlerFunc {
	domain := strings.ToLower(strings.TrimSpace(cfg.Cloudflare.Domain))
	sitesDir := cfg.Build.SitesDir
	frontendURL := strings.TrimRight(cfg.Server.FrontendURL, "/")
	if frontendURL == "" {
		frontendURL = "https://only4.fun"
	}

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.String(http.StatusInternalServerError, errorPage(500, "Server Error", "Something went wrong. Please try again later.", "Go to homepage", frontendURL))
			}
		}()

		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead &&
			c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut &&
			c.Request.Method != http.MethodPatch && c.Request.Method != http.MethodDelete {
			c.Status(http.StatusMethodNotAllowed)
			return
		}

		host := strings.ToLower(strings.TrimSpace(c.Request.Host))
		if i := strings.LastIndex(host, ":"); i > -1 {
			host = host[:i]
		}
		host = strings.TrimSuffix(host, ".")

		// Subdomain platform (project.only4.fun) dicari lewat nama project,
		// sedangkan host lain dicari sebagai domain kustom project.
		// Ambil subdomain dari host (*.only4.fun).
		var project *database.Project
		var subdomain string
		if domain != "" && strings.HasSuffix(host, "."+domain) {
			subdomain = strings.TrimSuffix(host, "."+domain)
			// Hanya satu label yang sah: portfolio.only4.fun, bukan a.b.only4.fun
			// atau nama berisi "../" yang bisa keluar dari folder situs.
			if !subdomainLabel.MatchString(subdomain) {
				subdomain = ""
			} else if found, err := projectStore.GetBySubdomain(subdomain); err == nil {
				project = found
			}
		} else if found, err := projectStore.GetByCustomDomain(host); err == nil {
			project = found
		}

		// Fallback: serve static site langsung dari disk jika project tidak ditemukan di DB
		if project == nil && subdomain != "" {
			subdomainFolder := filepath.Join(sitesDir, subdomain)
			if _, err := os.Stat(filepath.Join(subdomainFolder, "index.html")); err == nil {
				serveStatic(c, subdomainFolder, nil, projectStore)
				return
			}
			// Release disimpan di <sitesDir>/.releases/<subdomain>/<release>,
			// sama seperti yang dibuat builder.publish().
			releasesDir := filepath.Join(sitesDir, ".releases", subdomain)
			if entries, err := os.ReadDir(releasesDir); err == nil && len(entries) > 0 {
				latest := entries[len(entries)-1].Name()
				releaseRoot := filepath.Join(releasesDir, latest)
				if _, err := os.Stat(filepath.Join(releaseRoot, "index.html")); err == nil {
					serveStatic(c, releaseRoot, nil, projectStore)
					return
				}
			}
		}

		if project == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			message := fmt.Sprintf(
				"We couldn’t find a project for <strong>%s</strong>. It may have been removed, renamed, or its custom domain changed.",
				html.EscapeString(host),
			)
			c.String(http.StatusNotFound, errorPage(404, "This site isn’t here.", message, "Start a new project", frontendURL))
			return
		}
		if project.Status != "deployed" {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Header("Retry-After", "30")
			c.String(http.StatusServiceUnavailable, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Site is being prepared</title>
<meta http-equiv="refresh" content="30">
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; background: #fafafa; color: #171717; }
  .card { text-align: center; max-width: 480px; padding: 40px; }
  .spinner { width: 40px; height: 40px; border: 3px solid #e2e2e2; border-top-color: #2563eb; border-radius: 50%%; animation: spin 0.8s linear infinite; margin: 0 auto 24px; }
  @keyframes spin { to { transform: rotate(360deg); } }
  h1 { font-size: 20px; margin: 0 0 8px; }
  p { color: #8a8a8a; font-size: 14px; line-height: 1.6; margin: 0 0 16px; }
  .status { display: inline-block; padding: 4px 12px; border-radius: 999px; background: #f0f7ff; color: #2563eb; font-size: 12px; font-weight: 600; margin-bottom: 16px; }
</style>
</head>
<body>
<div class="card">
  <div class="spinner"></div>
  <span class="status">%s</span>
  <h1>Site is being prepared</h1>
  <p>Your site <strong>%s.only4.fun</strong> is currently being built. This page will automatically refresh every 30 seconds. Please check again shortly.</p>
  <p style="font-size:12px;color:#aaa;">Status: %s</p>
</div>
<script>
  (function check() {
    fetch(window.location.href, { method: 'HEAD' })
      .then(r => { if (r.ok) window.location.reload(); })
      .catch(() => {});
    setTimeout(check, 10000);
  })();
</script>
</body>
</html>`, html.EscapeString(project.Status), html.EscapeString(project.Subdomain), html.EscapeString(project.Status))
			return
		}

		root := filepath.Join(sitesDir, project.Subdomain)

		// Project dari ZIP selalu statis. Jangan pernah menjalankan server.js
		// dari arsip pengguna, walaupun file tersebut muncul setelah build.
		serverEntry := serverEntryForProject(root, project.RepoURL)
		if serverEntry != "" {
			// SSR mode: proxy ke Node.js process. Env var project ikut
			// diteruskan supaya API routes Next.js punya DB/API key saat runtime.
			envVars, _ := projectStore.GetEnvVars(project.ID)
			port, err := pm.Start(project.Subdomain, serverEntry, envVars)
			if err != nil {
				log.Printf("SSR start failed for %s: %v", project.Subdomain, err)
				c.String(http.StatusServiceUnavailable, "server is starting, please retry")
				return
			}
			proxyToPort(c, port)
			return
		}

		// Static mode: sajikan file statis
		serveStatic(c, root, project, projectStore)
	}
}

// serverEntryForProject menentukan apakah project boleh berjalan dalam mode SSR.
// Untuk project ZIP selalu mengembalikan "" agar tidak pernah ada proses node server.js dari arsip user.
func serverEntryForProject(root, repoURL string) string {
	if strings.HasPrefix(repoURL, "zip://") {
		return ""
	}
	return findServerEntry(root)
}

// findServerEntry mencari entry point server yang dinyatakan secara eksplisit.
// Repo statis yang kebetulan punya index.js/start.js tidak boleh dijalankan
// sebagai server, jadi SSR hanya aktif bila package.json punya script "start",
// tersedia output standalone Next.js, atau builder menulis penanda wrapper
// Next.js (project tanpa script "start").
func findServerEntry(root string) string {
	candidates := []string{".next/standalone/server.js"}
	if packageHasStartScript(root) || hasNextSSRMarker(root) {
		candidates = append(candidates,
			"server.js",
			"server.mjs",
			"build/server.js",
			"dist/server.js",
		)
	}
	for _, c := range candidates {
		p := filepath.Join(root, c)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

// hasNextSSRMarker melaporkan apakah builder membuat wrapper Next.js di root.
func hasNextSSRMarker(root string) bool {
	info, err := os.Stat(filepath.Join(root, nextSSRMarker))
	return err == nil && !info.IsDir()
}

// packageHasStartScript melaporkan apakah package.json mendefinisikan script "start".
func packageHasStartScript(root string) bool {
	raw, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return false
	}
	_, ok := pkg.Scripts["start"]
	return ok
}

// newSSRProxy menyiapkan reverse proxy ke proses SSR dengan timeout yang
// jelas. FlushInterval penting untuk response streaming (SSE/Next.js).
func newSSRProxy(port int) *httputil.ReverseProxy {
	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	proxy := httputil.NewSingleHostReverseProxy(target)
	transport := &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 2 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   16,
	}
	proxy.Transport = transport
	proxy.FlushInterval = -1 // flush segera, cocok untuk SSE/chunked response.
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("server error"))
	}
	return proxy
}

// proxyToPort mem-proxy request ke port tertentu.
func proxyToPort(c *gin.Context, port int) {
	newSSRProxy(port).ServeHTTP(c.Writer, c.Request)
}

// serveStatic melayani file statis dari folder project.
func serveStatic(c *gin.Context, root string, project *database.Project, projectStore *database.ProjectStore) {
	rel := strings.TrimPrefix(path.Clean(c.Request.URL.Path), "/")
	if rel == "" || rel == "." {
		rel = "index.html"
	}

	// Jangan pernah menyajikan dotfile dari hasil build: .env, .git, dan
	// folder .releases tidak boleh bisa diakses publik. .well-known dikecualikan
	// karena dipakai validasi SSL/DCV.
	for _, segment := range strings.Split(rel, "/") {
		if strings.HasPrefix(segment, ".") && segment != ".well-known" {
			c.String(http.StatusNotFound, "page not found")
			return
		}
	}

	target := filepath.Join(root, filepath.FromSlash(rel))
	// Cegah path traversal keluar dari folder situs.
	if !strings.HasPrefix(target, filepath.Clean(root)+string(os.PathSeparator)) {
		c.String(http.StatusNotFound, "page not found")
		return
	}

	if info, statErr := os.Stat(target); statErr != nil || info.IsDir() {
		if statErr == nil && info.IsDir() {
			index := filepath.Join(target, "index.html")
			if _, e := os.Stat(index); e == nil {
				target = index
			} else {
				c.String(http.StatusNotFound, "page not found")
				return
			}
		} else if isDocumentRequest(rel) {
			// SPA fallback: rute client-side tetap dilayani index.html.
			index := filepath.Join(root, "index.html")
			if _, e := os.Stat(index); e != nil {
				c.String(http.StatusNotFound, "page not found")
				return
			}
			target = index
		} else {
			// Aset yang tidak ada tetap 404, bukan HTML.
			c.String(http.StatusNotFound, "page not found")
			return
		}
	}

	c.Header("X-Content-Type-Options", "nosniff")
	// Hitung kunjungan hanya untuk permintaan halaman, bukan aset statis.
	if project != nil && isDocumentRequest(rel) {
		_ = projectStore.RecordVisit(project.ID)
	}
	http.ServeFile(c.Writer, c.Request, target)
}

func isDocumentRequest(rel string) bool {
	if rel == "" || rel == "index.html" {
		return true
	}
	switch strings.ToLower(filepath.Ext(rel)) {
	case "", ".html", ".htm":
		return true
	default:
		return false
	}
}

// errorPage generates a branded error page that mirrors the public site:
// floating pill navbar, oversized hero copy, four colored dots and the footer.
func errorPage(statusCode int, title string, message string, ctaText string, baseURL string) string {
	base := html.EscapeString(strings.TrimRight(baseURL, "/"))
	if base == "" {
		base = "https://only4.fun"
	}
	safeTitle := html.EscapeString(title)
	safeCTA := html.EscapeString(ctaText)

	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">
  <meta name="color-scheme" content="light dark">
  <title>` + safeTitle + ` — only4.fun</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Gabarito:wght@400;500;600;700;800&family=DM+Mono:wght@400;500&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg: #ffffff; --soft: #f1f3f4; --ink: #202124; --muted: #5f6368;
      --line: #dadce0; --blue: #4285f4; --red: #ea4335; --yellow: #fbbc04;
      --green: #34a853; --nav-h: 92px; --page-x: clamp(28px, 4vw, 72px);
      color-scheme: light;
    }
    @media (prefers-color-scheme: dark) {
      :root:not([data-theme='light']) {
        --bg: #202124; --soft: #2d2e31; --ink: #e8eaed; --muted: #9aa0a6;
        --line: #3c4043; color-scheme: dark;
      }
    }
    *, *::before, *::after { box-sizing: border-box; }
    html { scroll-behavior: smooth; }
    body {
      margin: 0;
      padding-top: calc(var(--nav-h) + env(safe-area-inset-top, 0px));
      background: var(--bg); color: var(--ink);
      font-family: 'Gabarito', system-ui, sans-serif;
      font-size: 20px; line-height: 1.5;
      -webkit-font-smoothing: antialiased;
    }
    :focus-visible { outline: 3px solid var(--blue); outline-offset: 3px; }
    .wrap { width: 100%; padding: 0 var(--page-x); }

    nav {
      position: fixed; top: calc(env(safe-area-inset-top, 0px) + 12px);
      left: var(--page-x); right: var(--page-x); z-index: 50;
      display: flex; align-items: center; justify-content: space-between;
      height: calc(var(--nav-h) - 24px); padding: 0 28px;
      background: var(--bg); border: 1px solid var(--line); border-radius: 999px;
    }
    .logo {
      display: flex; align-items: center; gap: 10px;
      font-weight: 800; font-size: 28px; letter-spacing: -0.045em;
      white-space: nowrap; text-decoration: none; color: var(--ink);
    }
    .logo .fun, .brandline .fun { color: var(--blue); }
    .dots4 {
      display: grid; grid-template-columns: 1fr 1fr; gap: 3px;
      width: var(--dot-size, 26px); height: var(--dot-size, 26px);
    }
    .dots4 i { border-radius: 50%; }
    .dots4 i:nth-child(1) { background: var(--blue); }
    .dots4 i:nth-child(2) { background: var(--red); }
    .dots4 i:nth-child(3) { background: var(--yellow); }
    .dots4 i:nth-child(4) { background: var(--green); }
    nav ul { display: flex; align-items: center; gap: 32px; list-style: none; margin: 0; padding: 0; }
    nav a { color: var(--muted); font-weight: 500; text-decoration: none; }
    nav a:hover { color: var(--ink); }
    nav a.btn {
      display: inline-block; padding: 12px 24px; border-radius: 999px;
      background: var(--blue); color: #fff; font-size: 17px; white-space: nowrap;
    }
    nav a.btn:hover { background: #2f72e0; color: #fff; }

    main { min-height: calc(100vh - var(--nav-h)); display: flex; flex-direction: column; }
    .hero { position: relative; padding: 64px 0 88px; overflow: hidden; }
    .hero-inner { position: relative; z-index: 1; }
    .eyebrow {
      display: inline-flex; align-items: center; gap: 10px;
      margin: 0 0 26px; padding: 8px 16px;
      border: 1px solid var(--line); border-radius: 999px;
      font: 500 14px/1 'DM Mono', monospace; letter-spacing: 0.04em;
      text-transform: uppercase; color: var(--muted);
    }
    .pulse { width: 8px; height: 8px; border-radius: 50%; background: var(--blue); }
    h1 {
      margin: 0; max-width: 15ch; font-size: clamp(52px, 8vw, 112px);
      line-height: 0.93; letter-spacing: -0.045em; font-weight: 800;
    }
    .hl { background: linear-gradient(transparent 62%, rgba(251, 188, 4, 0.55) 62%); padding: 0 0.04em; }
    .lead {
      max-width: 30em; margin: 26px 0 0;
      font-size: clamp(19px, 1.7vw, 24px); line-height: 1.45; color: var(--muted);
    }
    .lead strong { color: var(--ink); font-weight: 700; }
    .actions { display: flex; align-items: center; flex-wrap: wrap; gap: 18px; margin-top: 32px; }
    .btn {
      display: inline-flex; align-items: center; gap: 9px;
      padding: 15px 30px; border: 0; border-radius: 999px;
      background: var(--blue); color: #fff; font: 600 19px/1 'Gabarito', sans-serif;
      text-decoration: none; cursor: pointer; transition: background 0.15s;
    }
    .btn:hover { background: #2f72e0; color: #fff; }
    .secondary {
      font-size: 16px; color: var(--muted); text-decoration: underline;
      text-decoration-thickness: 1.5px; text-underline-offset: 0.2em;
    }
    .secondary:hover { color: var(--blue); }
    .hero-dots { position: absolute; top: 190px; right: calc(var(--page-x) + 40px); pointer-events: none; }
    .hero-dots .dots4 {
      grid-template-columns: repeat(2, 72px); grid-template-rows: repeat(2, 72px);
      width: auto; height: auto; gap: 10px;
      animation: dots-float 5.6s ease-in-out infinite;
    }
    .hero-dots .dots4 i:nth-child(2) { animation-delay: 0.12s; }
    .hero-dots .dots4 i:nth-child(3) { animation-delay: 0.24s; }
    .hero-dots .dots4 i:nth-child(4) { animation-delay: 0.36s; }
    @keyframes dots-float { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-14px); } }
    @media (prefers-reduced-motion: reduce) { .hero-dots .dots4 { animation: none; } }

    footer {
      --footer-bg: #202124; --footer-ink: #f1f3f4; --footer-muted: #9aa0a6;
      margin-top: auto; padding: 88px 0 0; overflow: hidden;
      background: var(--footer-bg); color: var(--footer-ink);
    }
    .footer-inner {
      display: grid; grid-template-columns: minmax(0, 1.2fr) minmax(0, 2fr);
      gap: 72px; padding-bottom: 72px;
    }
    .brandline {
      display: flex; align-items: center; gap: 12px;
      font-size: 30px; font-weight: 800; letter-spacing: -0.045em; white-space: nowrap;
    }
    .intro p { max-width: 34ch; margin: 20px 0 0; font-size: 18px; line-height: 1.5; color: var(--footer-muted); }
    .groups { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 32px; }
    .group h3 {
      margin: 0 0 20px; font: 500 13px/1 'DM Mono', monospace;
      letter-spacing: 0.1em; text-transform: uppercase; color: rgba(255, 255, 255, 0.45);
    }
    .group ul { display: grid; gap: 12px; list-style: none; margin: 0; padding: 0; }
    .group a { font-size: 16px; color: var(--footer-ink); text-decoration: none; }
    .group a:hover { color: var(--yellow); }
    .watermark-clip {
      display: flex; align-items: flex-end; justify-content: center;
      height: clamp(150px, 24vw, 340px); overflow: hidden;
    }
    .watermark {
      font-size: clamp(120px, 24vw, 380px); font-weight: 800; line-height: 0.9;
      letter-spacing: -0.06em; white-space: nowrap; user-select: none;
      color: rgba(255, 255, 255, 0.055); transform: translateY(4%);
    }
    .watermark .fun { color: rgba(66, 133, 244, 0.38); }

    @media (max-width: 1050px) { .hero-dots { display: none; } }
    @media (max-width: 900px) {
      nav ul li:not(.actions) { display: none; }
      .footer-inner { grid-template-columns: 1fr; gap: 48px; padding-bottom: 56px; }
      .groups { grid-template-columns: repeat(2, minmax(0, 1fr)); row-gap: 40px; }
    }
    @media (max-width: 700px) {
      :root { --page-x: 18px; }
      nav { padding: 0 18px; }
      .logo { font-size: 24px; gap: 8px; }
      nav a.btn { padding: 10px 18px; font-size: 15px; }
      .hero { padding: 26px 0 72px; }
      footer { padding-top: 64px; }
      .groups { grid-template-columns: 1fr; }
      .watermark-clip { height: 130px; }
      .watermark { font-size: 120px; }
    }
  </style>
</head>
<body>
  <nav aria-label="Main navigation">
    <a class="logo" href="` + base + `/"><span class="dots4" style="--dot-size:26px" aria-hidden="true"><i></i><i></i><i></i><i></i></span><span>Only4<span class="fun">fun</span></span></a>
    <ul>
      <li><a href="` + base + `/#how">How it works</a></li>
      <li><a href="` + base + `/#features">Features</a></li>
      <li><a href="` + base + `/instant">Instant Launch</a></li>
      <li class="actions"><a class="btn" href="` + base + `/launchpad">Get started</a></li>
    </ul>
  </nav>

  <main>
    <section class="hero wrap">
      <div class="hero-inner">
        <p class="eyebrow"><span class="pulse" aria-hidden="true"></span>Error ` + fmt.Sprintf("%d", statusCode) + `</p>
        <h1>` + safeTitle + `</h1>
        <p class="lead">` + message + `</p>
        <div class="actions">
          <a class="btn" href="` + base + `/deploy">` + safeCTA + `</a>
          <a class="secondary" href="` + base + `/">Go to homepage</a>
        </div>
      </div>
      <div class="hero-dots" aria-hidden="true"><span class="dots4"><i></i><i></i><i></i><i></i></span></div>
    </section>
  </main>

  <footer>
    <div class="wrap footer-inner">
      <div class="intro">
        <div class="brandline"><span class="dots4" style="--dot-size:26px" aria-hidden="true"><i></i><i></i><i></i><i></i></span><span>Only4<span class="fun">fun</span></span></div>
        <p>Push, build, live. Free hosting for the fun stuff — no config, no limits, no credit card.</p>
      </div>
      <nav class="groups" aria-label="Footer links">
        <div class="group">
          <h3>Explore</h3>
          <ul>
            <li><a href="` + base + `/">Home</a></li>
            <li><a href="` + base + `/#how">How it works</a></li>
            <li><a href="` + base + `/#features">Features</a></li>
            <li><a href="` + base + `/instant">Instant Launch</a></li>
            <li><a href="` + base + `/launchpad">Launchpad</a></li>
            <li><a href="` + base + `/deploy">Deploy new</a></li>
          </ul>
        </div>
        <div class="group">
          <h3>Resources</h3>
          <ul>
            <li><a href="` + base + `/docs">Docs</a></li>
            <li><a href="` + base + `/status">Status</a></li>
            <li><a href="` + base + `/support">Support</a></li>
          </ul>
        </div>
        <div class="group">
          <h3>Legal</h3>
          <ul>
            <li><a href="` + base + `/terms">Terms</a></li>
            <li><a href="` + base + `/privacy">Privacy</a></li>
          </ul>
        </div>
      </nav>
    </div>
    <div class="watermark-clip" aria-hidden="true"><div class="watermark">Only4<span class="fun">fun</span></div></div>
  </footer>
</body>
</html>`
}
