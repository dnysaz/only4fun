package server

import (
	"encoding/json"
	"fmt"
	"html"
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

// SiteHandler melayani hasil build berdasarkan Host request,
// misal portfolio.only4.fun -> <SITES_DIR>/portfolio/index.html
// Jika project memiliki server entry point (SSR), request akan di-proxy
// ke proses Node.js yang berjalan.
func SiteHandler(cfg *config.Config, projectStore *database.ProjectStore, pm *ProcessManager) gin.HandlerFunc {
	domain := strings.ToLower(strings.TrimSpace(cfg.Cloudflare.Domain))
	sitesDir := cfg.Build.SitesDir

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.String(http.StatusInternalServerError, errorPage(500, "Server Error", "Something went wrong. Please try again later.", "Go to homepage"))
			}
		}()

		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead &&
			c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut &&
			c.Request.Method != http.MethodPatch && c.Request.Method != http.MethodDelete {
			c.Status(http.StatusMethodNotAllowed)
			return
		}

		host := strings.ToLower(c.Request.Host)
		if i := strings.LastIndex(host, ":"); i > -1 {
			host = host[:i]
		}

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
			} else {
				project, _ = projectStore.GetBySubdomain(subdomain)
			}
		} else {
			project, _ = projectStore.GetByCustomDomain(host)
		}

		// Fallback: serve static site langsung dari disk jika project tidak ditemukan di DB
		if project == nil && subdomain != "" {
			subdomainFolder := filepath.Join(sitesDir, subdomain)
			indexPath := filepath.Join(subdomainFolder, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				c.File(indexPath)
				return
			}
			releasesDir := filepath.Join(filepath.Dir(sitesDir), ".releases", subdomain)
			if entries, err := os.ReadDir(releasesDir); err == nil && len(entries) > 0 {
				latest := entries[len(entries)-1].Name()
				indexPath = filepath.Join(releasesDir, latest, "index.html")
				if _, err := os.Stat(indexPath); err == nil {
					c.File(indexPath)
					return
				}
			}
		}

		if project == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusNotFound, errorPage(404, "Subdomain not found", "Want to use this subdomain?", "Register your project at only4.fun"))
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

		// Cek apakah ada server entry point (SSR mode)
		serverEntry := findServerEntry(root)
		if serverEntry != "" {
			// SSR mode: proxy ke Node.js process
			port, err := pm.Start(project.Subdomain, serverEntry)
			if err != nil {
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

// findServerEntry mencari entry point server yang dinyatakan secara eksplisit.
// Repo statis yang kebetulan punya index.js/start.js tidak boleh dijalankan
// sebagai server, jadi SSR hanya aktif bila package.json punya script "start"
// atau tersedia output standalone Next.js.
func findServerEntry(root string) string {
	candidates := []string{".next/standalone/server.js", "server.js"}
	if packageHasStartScript(root) {
		candidates = append(candidates,
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

// proxyToPort mem-proxy request ke port tertentu.
func proxyToPort(c *gin.Context, port int) {
	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 30 * time.Second,
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("server error"))
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

// serveStatic melayani file statis dari folder project.
func serveStatic(c *gin.Context, root string, project *database.Project, projectStore *database.ProjectStore) {
	rel := strings.TrimPrefix(path.Clean(c.Request.URL.Path), "/")
	if rel == "" || rel == "." {
		rel = "index.html"
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
	if isDocumentRequest(rel) {
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

// errorPage generates a custom HTML error page.
func errorPage(statusCode int, title string, message string, ctaText string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>` + title + ` — only4.fun</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; 
           display: flex; align-items: center; justify-content: center; min-height: 100vh; 
           margin: 0; background: #fafafa; color: #171717; }
    .card { text-align: center; max-width: 480px; padding: 48px 32px; border: 1px solid #e2e2e2; 
            border-radius: 12px; background: #fff; }
    .icon { width: 96px; height: 96px; margin: 0 auto 24px; background: #f0f0f0; 
            border-radius: 50%; display: flex; align-items: center; justify-content: center; }
    .icon svg { width: 48px; height: 48px; fill: #8a8a8a; }
    h1 { font-size: 28px; margin: 0 0 12px; color: #171717; font-weight: 700; }
    p { color: #666; font-size: 15px; line-height: 1.6; margin: 0 0 24px; }
    .cta { display: inline-block; padding: 12px 24px; margin-top: 8px; 
           background: #171717; color: #fff; text-decoration: none; 
           border-radius: 8px; font-size: 14px; font-weight: 600; }
    .cta:hover { background: #2563eb; }
    .status { display: inline-block; padding: 4px 12px; margin-bottom: 16px; 
              border-radius: 999px; background: #f0f7ff; color: #2563eb; 
              font-size: 12px; font-weight: 600; }
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
    </div>
    <span class="status">Error ` + fmt.Sprintf("%d", statusCode) + `</span>
    <h1>` + title + `</h1>
    <p>` + message + `</p>
    <a href="/" class="cta">` + ctaText + `</a>
  </div>
</body>
</html>`
}
