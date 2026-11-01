package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"only4.fun/backend/internal/config"
	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

// SiteHandler melayani hasil build berdasarkan Host request,
// misal portfolio.only4.fun -> <SITES_DIR>/portfolio/index.html
// Jika project memiliki server entry point (SSR), request akan di-proxy
// ke proses Node.js yang berjalan.
func SiteHandler(cfg *config.Config, projectStore *database.ProjectStore, pm *ProcessManager) gin.HandlerFunc {
	domain := strings.ToLower(strings.TrimSpace(cfg.Cloudflare.Domain))
	sitesDir := cfg.Build.SitesDir

	return func(c *gin.Context) {
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
		var project *database.Project
		if domain != "" && strings.HasSuffix(host, "."+domain) {
			subdomain := strings.TrimSuffix(host, "."+domain)
			// Hanya satu label: portfolio.only4.fun, bukan a.b.only4.fun
			if subdomain != "" && !strings.Contains(subdomain, ".") {
				project, _ = projectStore.GetBySubdomain(subdomain)
			}
		} else {
			project, _ = projectStore.GetByCustomDomain(host)
		}
		if project == nil {
			c.String(http.StatusNotFound, "site not found")
			return
		}
		if project.Status != "deployed" {
			c.String(http.StatusServiceUnavailable, "site is being prepared (status: %s)", project.Status)
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
	candidates := []string{".next/standalone/server.js"}
	if packageHasStartScript(root) {
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
