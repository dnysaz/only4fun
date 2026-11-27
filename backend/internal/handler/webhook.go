package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

// maxWebhookBody membatasi ukuran payload webhook GitHub (push event besar
// karena berisi daftar commit).
const maxWebhookBody = 5 << 20

// webhookDeliveryTTL mencegah delivery yang sama memicu dua build bila GitHub
// mengirim ulang callback setelah timeout jaringan.
const webhookDeliveryTTL = 10 * time.Minute

var (
	webhookDeliveryMu sync.Mutex
	webhookDeliveries = make(map[string]time.Time)
)

// markWebhookDelivery mengembalikan true bila delivery ID belum pernah diproses.
// ID kosong dianggap unik supaya webhook lama tanpa header tetap diproses.
func markWebhookDelivery(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return true
	}
	now := time.Now()
	webhookDeliveryMu.Lock()
	defer webhookDeliveryMu.Unlock()
	for key, at := range webhookDeliveries {
		if now.Sub(at) > webhookDeliveryTTL {
			delete(webhookDeliveries, key)
		}
	}
	if _, seen := webhookDeliveries[id]; seen {
		return false
	}
	webhookDeliveries[id] = now
	return true
}

// GetWebhookInfo menampilkan URL dan secret webhook untuk pemilik project.
// Nilai secret hanya dikirim ke pemilik yang sudah terotentikasi.
func (h *ProjectHandler) GetWebhookInfo(c *gin.Context) {
	project, ok := h.ownedProject(c)
	if !ok {
		return
	}
	secret, err := h.projectStore.EnsureWebhookSecret(project.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare the webhook secret"})
		return
	}

	// Bangun URL dari publicURL (jika di-set) atau dari request host,
	// supaya tetap benar di balik proxy Cloudflare.
	host := h.publicURL
	if host == "" {
		scheme := "http"
		if isHTTPS(c) {
			scheme = "https"
		}
		host = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}
	c.JSON(http.StatusOK, gin.H{
		"webhook": database.WebhookInfo{
			URL:    host + "/webhooks/github",
			Secret: secret,
			Events: []string{"push"},
		},
	})
}

// HandleGitHubWebhook menerima push event GitHub dan mengantrikan redeploy.
// Signature HMAC diverifikasi dengan secret project, jadi endpoint ini aman
// dipanggil tanpa session.
func (h *ProjectHandler) HandleGitHubWebhook(c *gin.Context) {
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxWebhookBody))
	if err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "webhook payload is too large"})
		return
	}

	var payload struct {
		Ref        string `json:"ref"`
		Repository struct {
			HTMLURL       string `json:"html_url"`
			CloneURL      string `json:"clone_url"`
			DefaultBranch string `json:"default_branch"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook payload"})
		return
	}

	// Cari project lewat URL repo (html_url maupun clone_url).
	var projects []*database.Project
	seen := map[string]bool{}
	for _, repoURL := range []string{payload.Repository.HTMLURL, payload.Repository.CloneURL} {
		if strings.TrimSpace(repoURL) == "" {
			continue
		}
		found, err := h.projectStore.GetByRepoURL(repoURL)
		if err != nil {
			continue
		}
		for _, p := range found {
			if !seen[p.ID] {
				seen[p.ID] = true
				projects = append(projects, p)
			}
		}
	}
	if len(projects) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no project uses this repository"})
		return
	}

	// Cocokkan signature dengan secret milik project.
	signature := c.GetHeader("X-Hub-Signature-256")
	var project *database.Project
	for _, p := range projects {
		// Jangan pernah membuat secret di jalur publik: hanya baca yang sudah
		// dibuat pemilik project lewat halaman webhook terotentikasi.
		secret, err := h.projectStore.GetWebhookSecret(p.ID)
		if err != nil || secret == "" {
			continue
		}
		if validWebhookSignature(secret, body, signature) {
			project = p
			break
		}
	}
	if project == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook signature"})
		return
	}

	// Hanya setelah HMAC valid: penyerang tidak boleh mengisi cache dengan ID
	// palsu, dan GitHub retry tidak boleh mengantrikan build dua kali.
	if !markWebhookDelivery(c.GetHeader("X-GitHub-Delivery")) {
		c.JSON(http.StatusAccepted, gin.H{"message": "duplicate delivery ignored"})
		return
	}

	event := c.GetHeader("X-GitHub-Event")
	if event == "ping" {
		c.JSON(http.StatusOK, gin.H{"message": "webhook connected"})
		return
	}
	if event != "push" {
		c.JSON(http.StatusOK, gin.H{"message": "event ignored", "event": event})
		return
	}

	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")
	if payload.Repository.DefaultBranch != "" && branch != payload.Repository.DefaultBranch {
		c.JSON(http.StatusOK, gin.H{"message": "not the default branch, deploy skipped", "branch": branch})
		return
	}
	if project.Status == "building" {
		c.JSON(http.StatusAccepted, gin.H{"message": "another build is running, this push was skipped"})
		return
	}

	if err := h.projectStore.UpdateStatus(project.ID, "pending"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start the deploy"})
		return
	}
	if err := h.enqueueDeploy(project, project.UserID); err != nil {
		_ = h.projectStore.UpdateStatus(project.ID, "failed")
		log.Printf("failed to queue the webhook deploy for project %s: %v", project.ID, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "deploy started", "project": project.Subdomain})
}

// validWebhookSignature membandingkan header X-Hub-Signature-256 dengan HMAC
// SHA-256 body memakai secret project.
func validWebhookSignature(secret string, body []byte, header string) bool {
	const prefix = "sha256="
	if secret == "" || !strings.HasPrefix(header, prefix) {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.TrimPrefix(header, prefix)))
}
