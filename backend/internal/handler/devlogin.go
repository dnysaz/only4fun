package handler

import (
	"log"
	"net"
	"net/http"
	"time"

	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/secure"

	"github.com/gin-gonic/gin"
)

// devDemoGitHubID menandai akun dummy yang hanya dipakai lewat /auth/dev-login.
const devDemoGitHubID = "dev-demo-account"

// devDemoProjects adalah isi dashboard akun dummy: dua situs live dan satu
// yang masih dibangun, supaya halaman tidak tampil kosong.
var devDemoProjects = []struct {
	Repo      string
	Subdomain string
	Status    string
	Visits    []int64
}{
	{
		Repo:      "https://github.com/demo-account/rocket-shop",
		Subdomain: "rocket-shop",
		Status:    "deployed",
		Visits:    []int64{4, 7, 5, 9, 12, 8, 14, 11, 18, 15, 22, 19, 26, 31},
	},
	{
		Repo:      "https://github.com/demo-account/doodle-zoo",
		Subdomain: "doodle-zoo",
		Status:    "deployed",
		Visits:    []int64{12, 10, 14, 9, 11, 16, 13, 17, 15, 20, 18, 24, 21, 28},
	},
	{
		Repo:      "https://github.com/demo-account/meme-lab",
		Subdomain: "meme-lab",
		Status:    "failed",
		Visits:    []int64{2, 1, 3, 0, 4, 2, 6, 3, 5, 1, 7, 4, 9, 6},
	},
}

// DevLoginHandler membuat sesi untuk akun dummy tanpa GitHub OAuth.
// Route ini hanya didaftarkan bila DEV_LOGIN=true dan menolak request dari
// luar loopback, jadi tidak boleh dipakai di production.
type DevLoginHandler struct {
	userStore    *database.UserStore
	projectStore *database.ProjectStore
	sessionStore *database.SessionStore
}

func NewDevLoginHandler(userStore *database.UserStore, projectStore *database.ProjectStore, sessionStore *database.SessionStore) *DevLoginHandler {
	return &DevLoginHandler{
		userStore:    userStore,
		projectStore: projectStore,
		sessionStore: sessionStore,
	}
}

// Login membuat (atau memakai ulang) akun demo, mengisi project contoh, lalu
// memasang cookie session dan mengarahkan ke halaman tujuan.
func (h *DevLoginHandler) Login(c *gin.Context) {
	if !isLoopback(c.ClientIP()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "dev login is only available from localhost"})
		return
	}

	user, err := h.userStore.GetByGitHubID(devDemoGitHubID)
	if err != nil {
		user, err = h.userStore.Create(devDemoGitHubID, "demo", "demo@only4.fun", "")
		if err != nil {
			log.Printf("dev login: failed to create demo user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create the demo user"})
			return
		}
	}
	h.seedProjects(user.ID)

	sessionID, err := secure.NewSessionID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session id"})
		return
	}
	if _, err := h.sessionStore.Create(user.ID, sessionID, "", time.Now().Add(24*time.Hour)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, sessionID, int((24 * time.Hour).Seconds()), "/", "", isHTTPS(c), true)

	// Endpoint ini sekarang POST: form dikirim lewat body, bukan query string.
	next := safeNextPath(c.PostForm("next"))
	if next == "" {
		next = "/launchpad"
	}
	c.Redirect(http.StatusFound, next)
}

// seedProjects mengisi project contoh bila belum ada, lalu menyegarkan data
// kunjungan dummy supaya chart traffic tidak kosong. Tidak menyentuh DNS atau
// builder, jadi aman dijalankan berulang.
func (h *DevLoginHandler) seedProjects(userID string) {
	existing, err := h.projectStore.GetByUserID(userID)
	if err != nil || len(existing) == 0 {
		for _, demo := range devDemoProjects {
			project, err := h.projectStore.Create(userID, demo.Repo, demo.Subdomain, nil)
			if err != nil {
				log.Printf("dev login: failed to seed project %s: %v", demo.Subdomain, err)
				continue
			}
			if demo.Status != "pending" {
				if err := h.projectStore.UpdateStatus(project.ID, demo.Status); err != nil {
					log.Printf("dev login: failed to set status for %s: %v", demo.Subdomain, err)
				}
			}
		}
	}

	for _, demo := range devDemoProjects {
		project, err := h.projectStore.GetBySubdomain(demo.Subdomain)
		if err != nil || project == nil || project.UserID != userID {
			continue
		}
		if err := h.projectStore.SeedVisits(project.ID, demo.Visits); err != nil {
			log.Printf("dev login: failed to seed visits for %s: %v", demo.Subdomain, err)
		}
	}
}

func isLoopback(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.IsLoopback()
}
