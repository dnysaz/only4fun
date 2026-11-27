package server

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"time"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/config"
	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/dns"
	"only4.fun/backend/internal/handler"
	"only4.fun/backend/internal/proxytrust"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Stores mengumpulkan semua store database agar bisa dipakai router dan worker.
type Stores struct {
	User       *database.UserStore
	Session    *database.SessionStore
	Project    *database.ProjectStore
	Deployment *database.DeploymentStore
	Audit      *database.AuditStore
}

func NewStores(db *database.Database) *Stores {
	return &Stores{
		User:       database.NewUserStore(db),
		Session:    database.NewSessionStore(db),
		Project:    database.NewProjectStore(db),
		Deployment: database.NewDeploymentStore(db),
		Audit:      database.NewAuditStore(db),
	}
}

// NewRouter membuat Gin engine dengan semua route terdaftar.
func NewRouter(cfg *config.Config, db *database.Database, worker *builder.Worker, pm *ProcessManager) *gin.Engine {
	r := gin.Default()

	// Hanya proxy yang dikonfigurasi yang boleh menentukan IP asli dan skema
	// HTTPS lewat header. Default-nya loopback; di belakang Cloudflare, isi
	// TRUSTED_PROXIES dengan rentang IP Cloudflare.
	trust, err := proxytrust.New(cfg.Server.TrustedProxies)
	if err != nil {
		panic(err)
	}
	if err := r.SetTrustedProxies(trust.Entries()); err != nil {
		panic(err)
	}
	r.RemoteIPHeaders = []string{"CF-Connecting-IP", "X-Forwarded-For", "X-Real-IP"}
	r.Use(TrustedProxyMiddleware(trust))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Inisialisasi handler dengan stores dari DB
	stores := NewStores(db)
	userStore := stores.User
	sessionStore := stores.Session
	projectStore := stores.Project
	authHandler, err := handler.NewAuthHandler(
		cfg.GitHub.ClientID,
		cfg.GitHub.ClientSecret,
		cfg.GitHub.RedirectURL,
		cfg.Server.FrontendURL,
		userStore,
		sessionStore,
	)
	if err != nil {
		panic(err)
	}
	// ProcessManager yang sama dipakai handler project, SiteHandler, dan signal
	// handler di Server, supaya proses SSR benar-benar dimatikan saat shutdown.
	dnsClient := dns.New(cfg.Cloudflare)
	projectHandler := handler.NewProjectHandler(projectStore, userStore, sessionStore, stores.Deployment, worker, dnsClient, pm, cfg.Build.SitesDir, cfg.Server.FrontendURL)
	githubHandler := handler.NewGitHubHandler(sessionStore)

	// Deploy instan tanpa login. Route ini sengaja didaftarkan di luar grup /api
	// supaya tidak kena AuthMiddleware, tetapi tetap sebelum NoRoute.
	instantHandler, err := handler.NewInstantHandler(
		userStore,
		projectStore,
		stores.Deployment,
		dnsClient,
		cfg.Build.SitesDir,
		cfg.Build.CacheDir,
		cfg.Cloudflare.Domain,
		time.Duration(cfg.Build.TimeoutSeconds)*time.Second,
	)
	if err != nil {
		log.Printf("instant deploy is unavailable: %v", err)
	}

	// Logging utama: durasi request, status, error count.
	r.Use(RequestLogger())

	// Health check (tanpa auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Metrics endpoint hanya untuk pemantauan lokal.
	metricsHandler := MetricsHandler()
	r.GET("/metrics", func(c *gin.Context) {
		if !metricsAllowed(c.ClientIP()) {
			c.Status(http.StatusNotFound)
			return
		}
		metricsHandler(c)
	})

	// Webhook GitHub (tanpa session; diverifikasi lewat HMAC signature).
	r.POST("/webhooks/github", RateLimit(120, time.Minute), projectHandler.HandleGitHubWebhook)

	// Instant deploy (tanpa session, static HTML/CSS/JS saja, alamat acak 8 char).
	// Origin guard menutup celah CSRF lewat multipart form lintas origin, dan
	// rate limit memakai IP (bukan cookie yang bisa dipalsukan).
	originGuard := TrustedOriginGuard(cfg.Server.CORSOrigins)
	if instantHandler != nil {
		r.POST("/api/instant/address", originGuard, RateLimitByIP(30, time.Minute), instantHandler.NewAddress)
		r.POST("/api/instant/repo-tree", originGuard, RateLimitByIP(20, time.Minute), instantHandler.RepoTree)
		r.POST("/api/instant/repo", originGuard, RateLimitByIP(6, time.Minute), instantHandler.DeployRepo)
		r.POST("/api/instant/zip", originGuard, RateLimitByIP(6, time.Minute), instantHandler.DeployZip)
	}

	// Auth handler routes
	auth := r.Group("/auth")
	{
		auth.GET("/github/login", RateLimit(30, time.Minute), authHandler.GetGitHubLoginURL)
		// Callback dipanggil GitHub tanpa session. Kuncinya IP, bukan cookie yang
		// bisa diganti penyerang untuk mendapatkan kuota baru.
		auth.GET("/github/callback", RateLimitByIP(20, time.Minute), authHandler.HandleGitHubCallback)
	}

	// Login dummy untuk pengembangan lokal; nonaktif kecuali DEV_LOGIN=true.
	if cfg.Server.DevLogin {
		log.Printf("DEV_LOGIN aktif: /auth/dev-login membuat sesi akun demo (jangan dipakai di production)")
		devLogin := handler.NewDevLoginHandler(userStore, projectStore, sessionStore)
		// GET hanya menampilkan form (tanpa efek samping); POST yang membuat sesi
		// dilindungi origin guard dan rate limit.
		r.GET("/auth/dev-login", func(c *gin.Context) {
			next := html.EscapeString(c.Query("next"))
			page := fmt.Sprintf(`<!doctype html><html lang="en"><head><meta charset="utf-8"><title>Dev login</title></head>
<body><h1>Dev login</h1><p>Local development only.</p>
<form method="post" action="/auth/dev-login"><input type="hidden" name="next" value="%s"><button type="submit">Log in as demo</button></form>
</body></html>`, next)
			c.Data(200, "text/html; charset=utf-8", []byte(page))
		})
		r.POST("/auth/dev-login", DevOriginGuard(), RateLimit(10, time.Minute), devLogin.Login)
	}

	// API routes (butuh auth middleware). Origin guard menutup CSRF lewat
	// multipart form dari halaman pada subdomain yang dibuat user.
	api := r.Group("/api")
	api.Use(originGuard)
	api.Use(handler.AuthMiddleware(sessionStore))
	api.Use(AuditMiddleware(stores.Audit))
	{
		api.GET("/auth/me", authHandler.GetMe)
		api.POST("/auth/logout", authHandler.Logout)

		// Projects
		api.GET("/projects", projectHandler.ListProjects)
		api.POST("/projects", RateLimit(10, time.Minute), projectHandler.CreateProject)
		api.GET("/check-subdomain", projectHandler.CheckSubdomain)
		api.GET("/projects/:id", projectHandler.GetProject)
		api.DELETE("/projects/:id", RateLimit(20, time.Minute), projectHandler.DeleteProject)
		api.GET("/projects/:id/env", projectHandler.GetEnv)
		api.POST("/projects/:id/env", RateLimit(20, time.Minute), projectHandler.UpdateEnv)
		api.PATCH("/projects/:id/env", RateLimit(20, time.Minute), projectHandler.UpdateEnv)
		api.PUT("/projects/:id/env", RateLimit(20, time.Minute), projectHandler.UpdateEnv)
		api.POST("/projects/:id/deploy", RateLimit(6, time.Minute), projectHandler.Redeploy)
		api.POST("/projects/upload", RateLimit(6, time.Minute), projectHandler.CreateFromZip)
		api.GET("/projects/:id/webhook", projectHandler.GetWebhookInfo)
		api.PUT("/projects/:id/domain", RateLimit(20, time.Minute), projectHandler.UpdateCustomDomain)
		api.GET("/projects/:id/domain/status", RateLimit(60, time.Minute), projectHandler.GetDomainStatus)
		api.GET("/projects/:id/deployments", projectHandler.ListDeployments)
		api.POST("/projects/:id/deployments/:deploymentID/rollback", RateLimit(10, time.Minute), projectHandler.RollbackDeployment)
		api.DELETE("/projects/:id/deployments", projectHandler.ClearDeployments)
		api.GET("/projects/:id/repo", projectHandler.GetRepoInfo)

		// GitHub
		api.GET("/github/repos", githubHandler.ListRepos)
		api.GET("/github/repos/:owner/:repo/contents", githubHandler.GetRepoContents)
		api.GET("/github/repos/:owner/:repo/readme", githubHandler.GetRepoReadme)
		api.GET("/github/profile", RateLimit(60, time.Minute), githubHandler.GetProfile)
	}

	// Sajikan situs hasil build berdasarkan Host (subdomain.only4.fun).
	siteHandler := SiteHandler(cfg, projectStore, pm)
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/auth/") {
			c.JSON(404, gin.H{"error": "endpoint not found"})
			return
		}
		siteHandler(c)
	})

	return r
}
