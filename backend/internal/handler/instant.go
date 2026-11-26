package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/dns"

	"github.com/gin-gonic/gin"
)

const (
	// anonymousGitHubID adalah identitas sistem untuk project instan tanpa login.
	// GitHub ID numerik, jadi nilai ini tidak mungkin bentrok dengan akun asli.
	anonymousGitHubID = "anonymous"
	anonymousUsername = "anonymous"

	// Alamat instan selalu 8 karakter huruf kecil + angka, mis. asj3jkha.only4.fun.
	instantSubdomainLength = 8
	instantAlphabet        = "abcdefghijklmnopqrstuvwxyz0123456789"

	instantTimeoutDefault = 90 * time.Second

	// Alamat hasil preview ditahan sebentar supaya dua pengunjung tidak pernah
	// melihat alamat yang sama, dan deploy memakai alamat yang benar-benar
	// ditampilkan. Database tetap menjadi sumber kebenaran terakhir.
	instantReservationTTL = 10 * time.Minute

	// Batas preview tree supaya halaman tidak memuat terlalu banyak baris.
	instantTreeLimit = 2000
)

var githubOwnerRepoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// Output build framework sering tidak menyisakan package.json atau file .jsx,
// sehingga dikenali dari struktur hasil build-nya juga.
var instantFrameworkBuildAsset = regexp.MustCompile(`(?i)(^|/)assets/[^/]+[-.][a-f0-9]{6,}\.(js|mjs)$|(^|/)static/js/[^/]+\.chunk\.js$`)

// errFrameworkProject dipakai saat source terlihat seperti aplikasi JS/framework.
// User harus login untuk endpoint build penuh.
var errFrameworkProject = errors.New("this looks like a JavaScript framework project. Sign in to deploy frameworks; instant deploy supports static HTML, CSS, and JS only")

type instantAddressReservation struct {
	token   string
	expires time.Time
}

var (
	instantReservationMu sync.Mutex
	instantReservations  = make(map[string]instantAddressReservation)
)

// sweepExpiredInstantReservations dipanggil dengan mutex sudah dipegang.
func sweepExpiredInstantReservations(now time.Time) {
	for subdomain, reservation := range instantReservations {
		if !now.Before(reservation.expires) {
			delete(instantReservations, subdomain)
		}
	}
}

func newInstantAddressToken() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// tryReserveInstantAddress hanya berhasil kalau alamat itu belum dipakai dan
// belum ditahan preview lain. Token dikembalikan supaya deploy bisa membuktikan
// bahwa alamat yang dikirimkannya berasal dari preview ini.
func tryReserveInstantAddress(subdomain string) (string, bool) {
	token, err := newInstantAddressToken()
	if err != nil {
		return "", false
	}

	instantReservationMu.Lock()
	defer instantReservationMu.Unlock()
	now := time.Now()
	sweepExpiredInstantReservations(now)
	if _, exists := instantReservations[subdomain]; exists {
		return "", false
	}
	instantReservations[subdomain] = instantAddressReservation{
		token:   token,
		expires: now.Add(instantReservationTTL),
	}
	return token, true
}

func isInstantAddressReserved(subdomain string) bool {
	instantReservationMu.Lock()
	defer instantReservationMu.Unlock()
	reservation, exists := instantReservations[subdomain]
	if !exists {
		return false
	}
	if !time.Now().Before(reservation.expires) {
		delete(instantReservations, subdomain)
		return false
	}
	return true
}

func claimInstantAddress(subdomain, token string) bool {
	instantReservationMu.Lock()
	defer instantReservationMu.Unlock()
	reservation, exists := instantReservations[subdomain]
	if !exists || reservation.token == "" || reservation.token != token {
		return false
	}
	if !time.Now().Before(reservation.expires) {
		delete(instantReservations, subdomain)
		return false
	}
	delete(instantReservations, subdomain)
	return true
}

var instantFrameworkEntries = map[string]bool{
	// Manifest / lockfile / config build.
	"package.json": true, "package-lock.json": true, "yarn.lock": true,
	"pnpm-lock.yaml": true, "bun.lock": true, "bun.lockb": true,
	"deno.json": true, "tsconfig.json": true,
	"next.config.js": true, "next.config.mjs": true, "next.config.ts": true,
	"vite.config.js": true, "vite.config.mjs": true, "vite.config.ts": true,
	"nuxt.config.js": true, "nuxt.config.mjs": true, "nuxt.config.ts": true,
	"astro.config.js": true, "astro.config.mjs": true, "astro.config.ts": true,
	"svelte.config.js": true, "angular.json": true, "gatsby-config.js": true,
	"gatsby-config.ts": true, "webpack.config.js": true, "webpack.config.mjs": true,
	"rollup.config.js": true,
	// Entry server yang akan dijalankan SiteHandler sebagai SSR.
	"server.js": true, "server.mjs": true, "server.cjs": true,
	// Folder framework yang menandakan proyek butuh build.
	".next": true, ".nuxt": true, ".svelte-kit": true, ".astro": true,
	"_next": true, "__next": true, "_astro": true,
	".output": true, ".turbo": true, "asset-manifest.json": true,
	"node_modules": true,
}

var instantFrameworkExts = map[string]bool{
	".vue": true, ".svelte": true, ".jsx": true, ".tsx": true, ".ts": true,
}

// InstantHandler melayani deploy tanpa sesi. Project yang dibuat selalu menjadi
// milik akun sistem "anonymous", sehingga tidak muncul di launchpad siapa pun
// dan tidak bisa di-edit, dihapus, atau dilihat statistiknya.
type InstantHandler struct {
	userStore       *database.UserStore
	projectStore    *database.ProjectStore
	deploymentStore *database.DeploymentStore
	dns             *dns.Client
	sitesDir        string
	cacheDir        string
	domain          string
	timeout         time.Duration
	anonymousUserID string
	githubAPI       string
	httpClient      *http.Client
}

func NewInstantHandler(
	userStore *database.UserStore,
	projectStore *database.ProjectStore,
	deploymentStore *database.DeploymentStore,
	dnsClient *dns.Client,
	sitesDir string,
	cacheDir string,
	domain string,
	timeout time.Duration,
) (*InstantHandler, error) {
	if timeout <= 0 {
		timeout = instantTimeoutDefault
	}
	if domain == "" {
		domain = "only4.fun"
	}

	anonymous, err := userStore.GetByGitHubID(anonymousGitHubID)
	if err != nil {
		anonymous, err = userStore.Create(anonymousGitHubID, anonymousUsername, "", "")
		if err != nil {
			// Dua instance yang start bersamaan bisa bentrok membuat akun sistem;
			// baca ulang sebelum menyerah.
			anonymous, err = userStore.GetByGitHubID(anonymousGitHubID)
			if err != nil {
				return nil, fmt.Errorf("failed to prepare the anonymous account: %w", err)
			}
		}
	}

	return &InstantHandler{
		userStore:       userStore,
		projectStore:    projectStore,
		deploymentStore: deploymentStore,
		dns:             dnsClient,
		sitesDir:        sitesDir,
		cacheDir:        cacheDir,
		domain:          domain,
		timeout:         timeout,
		anonymousUserID: anonymous.ID,
		githubAPI:       "https://api.github.com",
		httpClient:      &http.Client{Timeout: 12 * time.Second},
	}, nil
}

// NewAddress menghasilkan alamat acak yang tersedia tanpa membuat project.
// Dipakai halaman instant supaya alamat langsung terlihat saat input diisi.
func (h *InstantHandler) NewAddress(c *gin.Context) {
	subdomain, token, err := h.newAvailableAddress()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate an address, please try again"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"subdomain": subdomain,
		"url":       fmt.Sprintf("https://%s.%s", subdomain, h.domain),
		"token":     token,
	})
}

// RepoTree mengembalikan daftar file publik dari repo GitHub tanpa meng-clone
// atau menjalankan kode. Dipakai halaman instant untuk menampilkan hierarchy
// file sebelum user menyetujui terms.
func (h *InstantHandler) RepoTree(c *gin.Context) {
	var req struct {
		RepoURL string `json:"repo_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	owner, repo, err := instantGithubOwnerRepo(strings.TrimSpace(req.RepoURL))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 12*time.Second)
	defer cancel()

	var meta struct {
		DefaultBranch string `json:"default_branch"`
	}
	base := h.githubAPI
	if base == "" {
		base = "https://api.github.com"
	}
	if err := h.githubJSON(ctx, base+"/repos/"+owner+"/"+repo, &meta); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if meta.DefaultBranch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this repository has no default branch to preview"})
		return
	}

	var tree struct {
		Truncated bool `json:"truncated"`
		Tree      []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}
	treeURL := base + "/repos/" + owner + "/" + repo + "/git/trees/" + url.PathEscape(meta.DefaultBranch) + "?recursive=1"
	if err := h.githubJSON(ctx, treeURL, &tree); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paths := make([]string, 0, len(tree.Tree))
	truncated := tree.Truncated
	for _, item := range tree.Tree {
		if item.Type != "blob" {
			continue
		}
		path := cleanInstantTreePath(item.Path)
		if path == "" {
			continue
		}
		if len(paths) >= instantTreeLimit {
			truncated = true
			break
		}
		paths = append(paths, path)
	}
	if len(paths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files were found in this repository"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"paths": paths, "truncated": truncated})
}

// githubJSON mengambil JSON dari GitHub API dengan batas ukuran respons supaya
// endpoint publik tidak bisa dipakai untuk memuat data besar.
func (h *InstantHandler) githubJSON(ctx context.Context, endpoint string, out any) error {
	client := h.httpClient
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return errors.New("invalid GitHub repository")
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "only4fun-instant")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := client.Do(req)
	if err != nil {
		return errors.New("could not reach GitHub to preview this repository")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return errors.New("repository not found or not public")
		}
		if res.StatusCode == http.StatusForbidden {
			return errors.New("GitHub rate limit reached. Please try again in a minute")
		}
		return errors.New("could not read this repository right now")
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(out); err != nil {
		return errors.New("could not read the GitHub response")
	}
	return nil
}

// instantGithubOwnerRepo mengubah URL github.com/owner/repo menjadi pasangan
// owner/repo. Hanya hostname GitHub yang diterima untuk mencegah SSRF.
func instantGithubOwnerRepo(repoURL string) (string, string, error) {
	if err := validateRepoURL(repoURL); err != nil {
		return "", "", err
	}
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", errors.New("invalid GitHub repository URL")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("use a GitHub URL like https://github.com/owner/repo")
	}
	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")
	if owner == "." || owner == ".." || repo == "." || repo == ".." ||
		!githubOwnerRepoPattern.MatchString(owner) || !githubOwnerRepoPattern.MatchString(repo) {
		return "", "", errors.New("use a GitHub URL like https://github.com/owner/repo")
	}
	return owner, repo, nil
}

// cleanInstantTreePath membuang path yang tidak bisa ditampilkan dengan aman.
func cleanInstantTreePath(raw string) string {
	p := strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/")
	p = strings.TrimPrefix(p, "./")
	if p == "" || strings.HasPrefix(p, "/") {
		return ""
	}
	for _, segment := range strings.Split(p, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return ""
		}
	}
	if len(p) > 500 {
		p = p[:500]
	}
	return p
}

// DeployRepo menerima URL repo GitHub publik dan hanya menyajikan file statisnya.
// Tidak ada perintah npm/package manager yang dijalankan di jalur ini.
func (h *InstantHandler) DeployRepo(c *gin.Context) {
	var req struct {
		RepoURL      string `json:"repo_url" binding:"required"`
		Subdomain    string `json:"subdomain"`
		AddressToken string `json:"address_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	req.RepoURL = strings.TrimSpace(req.RepoURL)
	if err := validateRepoURL(req.RepoURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.sitesDir == "" || h.cacheDir == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "deployment folders are not configured"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	workDir, err := os.MkdirTemp(h.cacheDir, "instant-")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare the workspace"})
		return
	}
	defer os.RemoveAll(workDir)

	if err := cloneStaticRepo(ctx, req.RepoURL, workDir); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateStaticSite(workDir); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.finish(c, req.RepoURL, workDir, "GitHub repo", req.Subdomain, req.AddressToken)
}

// DeployZip menerima arsip ZIP statis. Subdomain hanya dipakai bila disertai
// token reservasi dari endpoint /address; selain itu server yang memilih alamat.
func (h *InstantHandler) DeployZip(c *gin.Context) {
	if h.sitesDir == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sites folder is not configured"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)
	if err := c.Request.ParseMultipartForm(16 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read the upload (maximum 100 MB)"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ZIP file not found in the \"file\" field"})
		return
	}
	defer file.Close()
	if !strings.EqualFold(filepath.Ext(header.Filename), ".zip") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "the file must have a .zip extension"})
		return
	}

	tmpZip, err := os.CreateTemp("", "instant-*.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare a temporary file"})
		return
	}
	defer os.Remove(tmpZip.Name())
	defer tmpZip.Close()

	if _, err := io.Copy(tmpZip, file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to store the upload"})
		return
	}

	extractDir, err := os.MkdirTemp("", "instant-site-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare the extraction folder"})
		return
	}
	defer os.RemoveAll(extractDir)

	if err := extractZip(tmpZip.Name(), extractDir); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	root := stripSingleRoot(extractDir)
	if err := validateStaticSite(root); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.finish(
		c,
		zipRepoPrefix+header.Filename,
		root,
		fmt.Sprintf("ZIP upload: %s", header.Filename),
		c.PostForm("subdomain"),
		c.PostForm("address_token"),
	)
}

// finish membuat project anonymous, mempublikasikan file, lalu mengembalikan
// alamat yang dihasilkan server. Project dihapus lagi bila publish gagal.
func (h *InstantHandler) finish(c *gin.Context, repoURL string, root string, logLabel string, preferred string, addressToken string) {
	project, err := h.createProject(repoURL, preferred, addressToken)
	if err != nil {
		log.Printf("failed to create an instant project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create the site, please try again"})
		return
	}

	target := filepath.Join(h.sitesDir, project.Subdomain)
	if err := builder.Publish(root, target); err != nil {
		log.Printf("failed to publish instant project %s: %v", project.ID, err)
		_ = h.projectStore.Delete(project.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish the files"})
		return
	}
	release, _ := builder.CurrentReleaseName(target)

	if err := h.projectStore.UpdateStatus(project.ID, "deployed"); err != nil {
		log.Printf("failed to update instant project %s: %v", project.ID, err)
	}
	if h.dns != nil && h.dns.Enabled() {
		if err := h.dns.EnsureRecord(c.Request.Context(), project.Subdomain); err != nil {
			log.Printf("failed to ensure DNS for instant site %s: %v", project.Subdomain, err)
		}
	}
	if h.deploymentStore != nil {
		record, err := h.deploymentStore.Create(project.ID, "success", fmt.Sprintf("%s (%d files)", logLabel, countFiles(root)))
		if err == nil && release != "" {
			_ = h.deploymentStore.SetRelease(record.ID, release)
		}
	}

	url := fmt.Sprintf("https://%s.%s", project.Subdomain, h.domain)
	c.JSON(http.StatusCreated, gin.H{
		"url": url,
		"project": gin.H{
			"subdomain":   project.Subdomain,
			"status":      "deployed",
			"url":         url,
			"static_only": true,
		},
	})
}

func (h *InstantHandler) createProject(repoURL string, preferred string, addressToken string) (*database.Project, error) {
	var lastErr error
	for attempt := 0; attempt < 6; attempt++ {
		var subdomain string

		// Alamat preview dipakai hanya bila tokennya masih cocok. Claim bersifat
		// atomik, jadi dua deploy tidak pernah memakai reservasi yang sama.
		if attempt == 0 && isValidInstantSubdomain(preferred) {
			if addressToken != "" {
				if claimInstantAddress(preferred, addressToken) {
					subdomain = preferred
				}
			} else if !isInstantAddressReserved(preferred) {
				// Klien lama yang belum mengirim token tetap didukung.
				if existing, err := h.projectStore.GetBySubdomain(preferred); err == nil && existing == nil {
					subdomain = preferred
				}
			}
		}

		if subdomain == "" {
			var err error
			subdomain, err = h.randomAvailableSubdomain()
			if err != nil {
				return nil, err
			}
		}

		// UNIQUE constraint di database adalah pengaman terakhir. Bila ada race
		// dengan request lain, ulangi dengan alamat baru yang belum dipakai.
		project, err := h.projectStore.Create(h.anonymousUserID, repoURL, subdomain, nil)
		if err == nil {
			return project, nil
		}
		lastErr = err
		if !strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			return nil, err
		}
	}
	return nil, fmt.Errorf("could not generate an available address: %w", lastErr)
}

// newAvailableAddress membuat alamat baru sekaligus menahannya. Pengecekan
// database dan reservasi dilakukan sebelum alamat dikembalikan ke halaman.
func (h *InstantHandler) newAvailableAddress() (string, string, error) {
	for attempt := 0; attempt < 40; attempt++ {
		subdomain, err := randomInstantSubdomain()
		if err != nil {
			return "", "", err
		}
		if validateSubdomain(subdomain) != nil {
			continue
		}
		if h.projectStore != nil {
			if existing, err := h.projectStore.GetBySubdomain(subdomain); err == nil && existing != nil {
				continue
			}
		}
		if token, ok := tryReserveInstantAddress(subdomain); ok {
			return subdomain, token, nil
		}
	}
	return "", "", errors.New("could not generate a unique address, please try again")
}

func (h *InstantHandler) randomAvailableSubdomain() (string, error) {
	for attempt := 0; attempt < 40; attempt++ {
		subdomain, err := randomInstantSubdomain()
		if err != nil {
			return "", err
		}
		// Pakai aturan yang sama dengan subdomain buatan user agar alamat acak
		// tidak mengandung kata terlarang.
		if validateSubdomain(subdomain) != nil {
			continue
		}
		if existing, err := h.projectStore.GetBySubdomain(subdomain); err == nil && existing != nil {
			continue
		}
		if isInstantAddressReserved(subdomain) {
			continue
		}
		return subdomain, nil
	}
	return "", errors.New("could not generate a unique address, please try again")
}

// isValidInstantSubdomain menerima hanya alamat acak 8 karakter yang dihasilkan
// server, bukan nama pilihan user (mis. "my-site" tetap ditolak).
func isValidInstantSubdomain(subdomain string) bool {
	if len(subdomain) != instantSubdomainLength {
		return false
	}
	for _, c := range subdomain {
		if !strings.ContainsRune(instantAlphabet, c) {
			return false
		}
	}
	return validateSubdomain(subdomain) == nil
}

func randomInstantSubdomain() (string, error) {
	out := make([]byte, instantSubdomainLength)
	max := big.NewInt(int64(len(instantAlphabet)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = instantAlphabet[n.Int64()]
	}
	return string(out), nil
}

func cloneStaticRepo(ctx context.Context, repoURL string, workDir string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("git is not available on the server")
	}
	// "--" mencegah URL dibaca git sebagai opsi, bukan argumen.
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--quiet", "--", repoURL, workDir)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Output git tidak dipercaya: buang control character/newline supaya
		// tidak bisa dipakai untuk menyisipkan markup di halaman error frontend.
		detail := strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == '\t' || r < 32 || r == 127 {
				return ' '
			}
			return r
		}, string(out))
		detail = strings.Join(strings.Fields(detail), " ")
		if len(detail) > 300 {
			detail = detail[:300] + "…"
		}
		if detail == "" {
			detail = "make sure the repository is public"
		}
		return fmt.Errorf("could not read that repository: %s", detail)
	}
	return nil
}

// validateStaticSite memastikan tidak ada manifest/build/entry server yang
// membuat jalur instan menjalankan kode JS selain file statis.
func validateStaticSite(root string) error {
	if !instantFileExists(filepath.Join(root, "index.html")) {
		return errors.New("the site must have index.html at its root (or inside a single folder)")
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return errors.New("could not read the site files")
	}
	for _, entry := range entries {
		name := strings.ToLower(entry.Name())
		if instantFrameworkEntries[name] {
			return errFrameworkProject
		}
	}
	// SiteHandler menjalankan Node bila ada entry server di output build.
	for _, rel := range []string{"dist/server.js", "dist/server.mjs", "build/server.js", "build/server.mjs"} {
		if instantFileExists(filepath.Join(root, rel)) {
			return errFrameworkProject
		}
	}

	// File sumber framework bisa berada di subfolder (mis. src/App.jsx).
	checked := 0
	err = filepath.Walk(root, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info == nil {
			return nil
		}
		if info.IsDir() && (info.Name() == ".git" || info.Name() == "node_modules") {
			return filepath.SkipDir
		}
		if instantFrameworkEntries[strings.ToLower(info.Name())] {
			return errFrameworkProject
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		checked++
		if checked > 5000 {
			return filepath.SkipAll
		}
		if instantFrameworkExts[strings.ToLower(filepath.Ext(info.Name()))] {
			return errFrameworkProject
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr == nil && instantFrameworkBuildAsset.MatchString(filepath.ToSlash(rel)) {
			return errFrameworkProject
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func instantFileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.Mode().IsRegular()
}
