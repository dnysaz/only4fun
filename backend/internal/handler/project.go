package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/dns"
	idgen "only4.fun/backend/internal/id"

	"github.com/gin-gonic/gin"
)

// ProcessStopper menghentikan proses SSR yang sedang berjalan untuk subdomain.
// Dipakai saat project deleted agar proses Node.js-nya tidak tertinggal.
type ProcessStopper interface {
	Stop(subdomain string)
}

type ProjectHandler struct {
	projectStore    *database.ProjectStore
	userStore       *database.UserStore
	sessionStore    *database.SessionStore
	deploymentStore *database.DeploymentStore
	worker          *builder.Worker
	dns             *dns.Client
	processes       ProcessStopper
	sitesDir        string
	publicURL       string
}

func NewProjectHandler(
	projectStore *database.ProjectStore,
	userStore *database.UserStore,
	sessionStore *database.SessionStore,
	deploymentStore *database.DeploymentStore,
	worker *builder.Worker,
	dnsClient *dns.Client,
	processes ProcessStopper,
	sitesDir string,
	publicURL string,
) *ProjectHandler {
	return &ProjectHandler{
		projectStore:    projectStore,
		userStore:       userStore,
		sessionStore:    sessionStore,
		deploymentStore: deploymentStore,
		worker:          worker,
		dns:             dnsClient,
		processes:       processes,
		sitesDir:        sitesDir,
		publicURL:       strings.TrimRight(publicURL, "/"),
	}
}

var subdomainPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// Nama subdomain yang dipakai infrastruktur / halaman aplikasi.
var reservedSubdomains = map[string]bool{
	"www": true, "api": true, "app": true, "admin": true, "auth": true,
	"docs": true, "help": true, "profile": true, "dashboard": true, "mail": true,
}

// containsForbiddenWord checks if the name contains any forbidden word.
var forbiddenWords = []string{
	"admin", "root", "fuck", "shit", "ass", "bitch", "damn", "crap", "dick",
	"piss", "cock", "pussy", "bastard", "whore", "slut", "cunt", "twat",
	"anjing", "bangsat", "babi", "kontol", "memek", "ngentot", "tolol",
	"goblok", "bodoh", "sialan", "brengsek", "keparat", "asu", "monyet",
	"tai", "perek", "lonte", "sundal", "jalang", "pelacur", "setan", "iblis",
	"porn", "porno", "xxx", "sex", "sexy", "nude", "naked", "nsfw",
	"president", "presiden", "minister", "menteri", "governor", "gubernur",
	"mayor", "walikota", "senator", "trump", "biden", "obama", "putin",
	"jokowi", "prabowo", "megawati", "anies", "ganjar", "modi", "macron",
	"merkel", "elon", "musk", "bezos", "zuckerberg", "gates",
	"support", "helpdesk", "system", "backend", "frontend", "api", "www",
	"mail", "ftp", "smtp", "dns", "cdn", "proxy", "gateway", "router",
	"firewall", "server", "host", "localhost", "webmaster", "security",
	"database", "db", "sql", "mysql", "postgres", "mongo", "nginx",
	"apache", "docker", "kubernetes", "aws", "azure", "gcp", "cloud",
	"cloudflare", "vercel", "netlify", "github", "gitlab", "bitbucket",
	"git", "slack", "discord", "telegram", "whatsapp", "twitter",
	"facebook", "instagram", "linkedin", "youtube", "tiktok", "reddit",
	"google", "microsoft", "apple", "amazon", "meta", "openai", "stripe",
	"paypal", "visa", "mastercard", "amex", "only4", "only4fun", "fun",
	"only", "dashboard", "panel", "console", "terminal", "shell",
	"command", "login", "logout", "register", "signup", "signin",
	"password", "passwd", "credential", "credentials", "permission",
	"permissions", "role", "roles", "group", "groups", "team", "teams",
	"org", "organization", "billing", "payment", "invoice", "subscription",
	"plan", "pricing", "notification", "notifications", "alert", "alerts",
	"message", "messages", "email", "emails", "mail", "mailbox", "inbox",
	"outbox", "newsletter", "search", "find", "query", "filter", "sort",
	"order", "list", "create", "update", "delete", "remove", "add", "edit",
	"view", "show", "hide", "enable", "disable", "activate", "deactivate",
	"start", "stop", "restart", "run", "execute", "process", "job", "task",
	"schedule", "cron", "import", "export", "upload", "download", "file",
	"files", "folder", "directory", "path", "route", "routing", "redirect",
	"rewrite", "static", "assets", "public", "media", "image", "images",
	"video", "audio", "document", "documents", "attachment", "attachments",
	"json", "xml", "html", "css", "js", "javascript", "typescript",
	"node", "npm", "yarn", "pnpm", "go", "golang", "rust", "python",
	"java", "php", "ruby", "dotnet", "csharp", "swift", "kotlin",
	"docker", "kubernetes", "k8s", "helm", "terraform", "ansible",
	"aws", "azure", "gcp", "cloud", "help", "profile",
}

func containsForbiddenWord(name string) bool {
	lower := strings.ToLower(name)
	for _, word := range forbiddenWords {
		if strings.Contains(lower, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

func validateSubdomain(subdomain string) error {
	if subdomain == "" {
		return fmt.Errorf("project name is required")
	}
	if len(subdomain) < 3 {
		return fmt.Errorf("project name must be at least 3 characters")
	}
	if len(subdomain) > 63 {
		return fmt.Errorf("project name must be at most 63 characters")
	}
	if !subdomainPattern.MatchString(subdomain) {
		return fmt.Errorf("project name may only contain lowercase letters, numbers and dashes, and must not start or end with a dash")
	}
	if containsForbiddenWord(subdomain) {
		return fmt.Errorf("project name contains a restricted word, please choose another")
	}
	if reservedSubdomains[subdomain] {
		return fmt.Errorf("project name %q is reserved, please choose another", subdomain)
	}
	return nil
}

// validateEnvKeys memastikan nama env var aman dipakai sebagai environment build.
func validateEnvKeys(envVars map[string]string) error {
	for k := range envVars {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("environment variable name must not be empty")
		}
		if !isValidEnvKey(k) {
			return fmt.Errorf("invalid environment variable name %q: use letters, numbers and underscores, starting with a letter or underscore", k)
		}
	}
	return nil
}

func isValidEnvKey(key string) bool {
	if len(key) == 0 || len(key) > 256 {
		return false
	}
	// Must start with letter or underscore
	if !((key[0] >= 'A' && key[0] <= 'Z') || (key[0] >= 'a' && key[0] <= 'z') || key[0] == '_') {
		return false
	}
	for i := 1; i < len(key); i++ {
		c := key[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

func validateRepoURL(repoURL string) error {
	if repoURL == "" {
		return fmt.Errorf("repo_url is required")
	}
	u, err := url.Parse(repoURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("repo_url must be a valid http(s) URL")
	}
	// Hanya izinkan GitHub: mencegah SSRF ke alamat internal lewat URL manual.
	host := strings.ToLower(u.Hostname())
	if host != "github.com" && host != "www.github.com" {
		return fmt.Errorf("only GitHub repository URLs are supported")
	}
	if u.User != nil {
		return fmt.Errorf("repo_url must not contain credentials")
	}
	return nil
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	projects, err := h.projectStore.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load projects"})
		return
	}
	for _, p := range projects {
		p.EnvKeys = database.ParseEnvKeys(p.EnvVars)
	}
	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req struct {
		RepoURL   string            `json:"repo_url" binding:"required"`
		Subdomain string            `json:"subdomain" binding:"required"`
		EnvVars   map[string]string `json:"env_vars"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.Subdomain = strings.ToLower(strings.TrimSpace(req.Subdomain))
	req.RepoURL = strings.TrimSpace(req.RepoURL)

	if err := validateSubdomain(req.Subdomain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateRepoURL(req.RepoURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// CreateProject sebelumnya menerima env_vars tanpa validasi, berbeda dari
	// UpdateEnv, sehingga nama key aneh bisa masuk ke environment build.
	if err := validateEnvKeys(req.EnvVars); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if existing, err := h.projectStore.GetBySubdomain(req.Subdomain); err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("project name %q is already taken", req.Subdomain)})
		return
	}
	// GetByRepoURL mengembalikan slice kosong (bukan nil) bila tidak ada hasil,
	// jadi cek panjang slice — hit[0] pada slice kosong akan panic saat request.
	if hit, err := h.projectStore.GetByRepoURL(req.RepoURL); err == nil && len(hit) > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("this repository is already deployed as %s. Delete it first, then deploy again", hit[0].Subdomain)})
		return
	}

	project, err := h.projectStore.Create(userID, req.RepoURL, req.Subdomain, req.EnvVars)
	if err != nil {
		// Jaga-jaga bila dua request membuat subdomain sama secara bersamaan.
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("project name %q is already taken", req.Subdomain)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create project"})
		return
	}
	project.EnvKeys = database.ParseEnvKeys(project.EnvVars)

	// Langsung antrikan build pertama.
	if err := h.enqueueDeploy(project, userID); err != nil {
		log.Printf("failed to queue build for project %s: %v", project.ID, err)
	}

	c.JSON(http.StatusCreated, gin.H{"project": project})
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}
	project.EnvKeys = database.ParseEnvKeys(project.EnvVars)
	if visits, err := h.projectStore.GetVisits(project.ID, 14); err == nil {
		project.Visits = visits
	}
	c.JSON(http.StatusOK, gin.H{"project": project})
}

// GetRepoInfo mengambil detail repository dari GitHub memakai token pemilik project.
func (h *ProjectHandler) GetRepoInfo(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}

	owner, repo, err := githubOwnerRepo(project.RepoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session *database.Session
	if h.sessionStore != nil {
		if s, err := h.sessionStore.GetLatestByUserID(project.UserID); err == nil {
			session = s
		}
	}
	if session == nil || session.Token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitHub token is unavailable, please sign in again"})
		return
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create the GitHub request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+session.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach GitHub"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found or not accessible with this token"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("GitHub rejected the request (status %d). Please sign in again", resp.StatusCode)})
		return
	}

	var payload struct {
		FullName      string `json:"full_name"`
		Description   string `json:"description"`
		HTMLURL       string `json:"html_url"`
		Private       bool   `json:"private"`
		Language      string `json:"language"`
		DefaultBranch string `json:"default_branch"`
		Stars         int    `json:"stargazers_count"`
		Forks         int    `json:"forks_count"`
		OpenIssues    int    `json:"open_issues_count"`
		PushedAt      string `json:"pushed_at"`
		UpdatedAt     string `json:"updated_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read repository data from GitHub"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"repo": payload})
}

// githubOwnerRepo mengambil owner dan nama repo dari URL GitHub.
func githubOwnerRepo(repoURL string) (string, string, error) {
	u, err := url.Parse(repoURL)
	if err != nil || u.Host == "" {
		return "", "", fmt.Errorf("invalid repo_url")
	}
	if !strings.EqualFold(u.Host, "github.com") && !strings.EqualFold(u.Host, "www.github.com") {
		return "", "", fmt.Errorf("repository details are only available for GitHub repositories")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repo_url")
	}
	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid repo_url")
	}
	return owner, repo, nil
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}
	// Hentikan server SSR yang masih berjalan; tanpa ini proses Node.js
	// tetap hidup sampai idle cleanup (10 menit) setelah project deleted.
	if h.processes != nil {
		h.processes.Stop(project.Subdomain)
	}
	if err := h.projectStore.Delete(project.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
		return
	}

	// Hapus juga record DNS subdomain (best effort).
	if h.dns != nil && h.dns.Enabled() {
		if err := h.dns.DeleteRecord(c.Request.Context(), project.Subdomain); err != nil {
			log.Printf("failed to delete DNS record %s: %v", project.Subdomain, err)
		}
	}

	// Hapus file situs hasil build. Folder aktif berupa symlink, jadi rilis
	// lama di .releases/<subdomain> juga harus dibersihkan.
	if h.sitesDir != "" && validateSubdomain(project.Subdomain) == nil {
		if err := os.RemoveAll(filepath.Join(h.sitesDir, project.Subdomain)); err != nil {
			log.Printf("failed to delete site files %s: %v", project.Subdomain, err)
		}
		if err := os.RemoveAll(filepath.Join(h.sitesDir, ".releases", project.Subdomain)); err != nil {
			log.Printf("failed to delete site releases %s: %v", project.Subdomain, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}

func (h *ProjectHandler) UpdateEnv(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}
	var req struct {
		EnvVars map[string]string `json:"env_vars"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	if req.EnvVars == nil {
		req.EnvVars = map[string]string{}
	}
	if err := validateEnvKeys(req.EnvVars); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.projectStore.UpdateEnvVars(project.ID, req.EnvVars); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save environment variables"})
		return
	}
	// Ambil key terbaru setelah merge.
	merged, _ := h.projectStore.GetEnvVars(project.ID)
	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	c.JSON(http.StatusOK, gin.H{
		"message":  "environment variables updated",
		"env_keys": keys,
	})
}

// GetEnv mengembalikan nilai env vars hanya untuk pemilik project.
func (h *ProjectHandler) GetEnv(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}
	// GetEnv mengembalikan daftar nama key env (tanpa nilai) untuk pemilik project.
	// Nilai env tidak pernah dikirim ke browser demi keamanan.
	envVars, err := h.projectStore.GetEnvVars(project.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load environment variables"})
		return
	}
	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	c.JSON(http.StatusOK, gin.H{"env_keys": keys})
}

// CheckSubdomain checks if a subdomain is available.
func (h *ProjectHandler) CheckSubdomain(c *gin.Context) {
	name := strings.ToLower(strings.TrimSpace(c.Query("name")))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project name is required"})
		return
	}
	if err := validateSubdomain(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"available": false, "error": err.Error()})
		return
	}
	if existing, err := h.projectStore.GetBySubdomain(name); err == nil && existing != nil {
		c.JSON(http.StatusOK, gin.H{"available": false, "error": "project name is already taken"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"available": true})
}

func (h *ProjectHandler) Redeploy(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}
	if strings.HasPrefix(project.RepoURL, zipRepoPrefix) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this project was created from a ZIP file; create a new project with the latest ZIP to publish an update"})
		return
	}
	if err := h.projectStore.UpdateStatus(project.ID, "pending"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start redeploy"})
		return
	}
	if err := h.enqueueDeploy(project, project.UserID); err != nil {
		_ = h.projectStore.UpdateStatus(project.ID, "failed")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "redeploy started"})
}

// ListDeployments mengembalikan riwayat deployment beserta log build milik project.
func (h *ProjectHandler) ListDeployments(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}
	if h.deploymentStore == nil {
		c.JSON(http.StatusOK, gin.H{"deployments": []any{}})
		return
	}
	deployments, err := h.deploymentStore.GetByProjectID(project.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load deployment history"})
		return
	}
	if len(deployments) > 20 {
		deployments = deployments[:20]
	}
	c.JSON(http.StatusOK, gin.H{"deployments": deployments})
}

// RollbackDeployment mempromosikan folder rilis dari deployment lama menjadi
// versi aktif. Tiga rilis terbaru disimpan, jadi deployment yang lebih lama
// dari itu tidak bisa dipulihkan.
func (h *ProjectHandler) RollbackDeployment(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}
	if h.deploymentStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "deployment history is unavailable"})
		return
	}
	deploymentID, err := parseID(c.Param("deploymentID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deployment id"})
		return
	}
	dep, err := h.deploymentStore.GetByID(deploymentID)
	if err != nil || dep == nil || dep.ProjectID != project.ID {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}
	if dep.Status != "success" || dep.Release == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this deployment has no release that can be restored"})
		return
	}

	target := filepath.Join(h.sitesDir, project.Subdomain)
	if err := builder.ActivateRelease(target, dep.Release); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Hentikan server SSR supaya versi lama tidak terus disajikan; request
	// berikutnya menyalakan proses baru dari rilis yang baru diaktifkan.
	if h.processes != nil {
		h.processes.Stop(project.Subdomain)
	}
	if err := h.projectStore.UpdateStatus(project.ID, "deployed"); err != nil {
		log.Printf("failed to update project status %s after rollback: %v", project.ID, err)
	}

	if record, err := h.deploymentStore.Create(project.ID, "success", fmt.Sprintf("rollback to release %s (deployment %s)", dep.Release, dep.ID)); err == nil {
		_ = h.deploymentStore.SetRelease(record.ID, dep.Release)
	}

	c.JSON(http.StatusOK, gin.H{"message": "rollback succeeded", "release": dep.Release})
}

func (h *ProjectHandler) ClearDeployments(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}
	if h.deploymentStore == nil {
		c.JSON(http.StatusOK, gin.H{"cleared": true})
		return
	}
	if err := h.deploymentStore.ClearByProjectID(project.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear deployment history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cleared": true})
}

// CheckSubdomain checks if a subdomain is available.
// enqueueDeploy mengirim job build ke worker dengan env vars dan token GitHub pemilik.
func (h *ProjectHandler) enqueueDeploy(project *database.Project, userID string) error {
	if h.worker == nil {
		return fmt.Errorf("build worker is not running")
	}

	envVars, err := h.projectStore.GetEnvVars(project.ID)
	if err != nil {
		envVars = map[string]string{}
	}

	token := ""
	if h.sessionStore != nil {
		if session, err := h.sessionStore.GetLatestByUserID(userID); err == nil && session != nil {
			token = session.Token
		}
	}

	return h.worker.Enqueue(builder.Job{
		ProjectID: project.ID,
		UserID:    userID,
		Subdomain: project.Subdomain,
		RepoURL:   project.RepoURL,
		EnvVars:   envVars,
		Token:     token,
	})
}

// ownedProject mengambil project dari :id dan memastikan milik user yang login.
// Project milik user lain dijawab 404 agar keberadaannya tidak terbocor.
func (h *ProjectHandler) ownedProject(c *gin.Context) (*database.Project, bool) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return nil, false
	}
	projectID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return nil, false
	}
	project, err := h.projectStore.GetByID(projectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return nil, false
	}
	if project.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return nil, false
	}
	return project, true
}

func parseID(s string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(s))
	if !idgen.Valid(value) {
		return "", fmt.Errorf("invalid id")
	}
	return value, nil
}
