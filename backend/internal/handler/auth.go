package handler

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/secure"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// oauthStateCookie menyimpan state OAuth sementara untuk mencegah CSRF.
const oauthStateCookie = "oauth_state"

// sessionCookieName adalah cookie httpOnly berisi ID session.
const sessionCookieName = "session_id"

// oauthNextCookie menyimpan halaman tujuan setelah login (mis. halaman yang
// baru saja diblokir karena belum ada sesi). Nilainya hanya berupa path lokal.
const oauthNextCookie = "oauth_next"

// safeNextPath menerima path lokal internal dan menolak tujuan eksternal supaya
// parameter next tidak bisa dipakai untuk open redirect.
func safeNextPath(next string) string {
	next = strings.TrimSpace(next)
	if next == "" || len(next) > 512 {
		return ""
	}
	if !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		return ""
	}
	if strings.ContainsAny(next, "\\\r\n") {
		return ""
	}
	return next
}

// AuthHandler menangani alur autentikasi GitHub OAuth.
type AuthHandler struct {
	oauth2Config *oauth2.Config
	userStore    *database.UserStore
	sessionStore *database.SessionStore
	frontendURL  string
}

// NewAuthHandler membuat instance AuthHandler.
func NewAuthHandler(clientID, clientSecret, redirectURL, frontendURL string, userStore *database.UserStore, sessionStore *database.SessionStore) (*AuthHandler, error) {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"user:email", "read:user", "repo"},
		Endpoint:     github.Endpoint,
		RedirectURL:  redirectURL,
	}
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	return &AuthHandler{
		oauth2Config: cfg,
		userStore:    userStore,
		sessionStore: sessionStore,
		frontendURL:  strings.TrimRight(frontendURL, "/"),
	}, nil
}

// isHTTPS menentukan apakah request asli dari browser memakai HTTPS.
// Di belakang Cloudflare/nginx TLS berhenti di proxy, sehingga r.TLS nil;
// header X-Forwarded-Proto / CF-Visitor dipakai sebagai sumber kebenaran.
func isHTTPS(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")), "https") {
		return true
	}
	if strings.Contains(strings.ToLower(c.GetHeader("CF-Visitor")), `"scheme":"https"`) {
		return true
	}
	return false
}

// randomState membuat nilai state OAuth yang tidak bisa ditebak.
func randomState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// GetGitHubLoginURL mengembalikan URL redirect ke GitHub OAuth.
func (h *AuthHandler) GetGitHubLoginURL(c *gin.Context) {
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create OAuth state"})
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, state, 600, "/", "", isHTTPS(c), true)
	if next := safeNextPath(c.Query("next")); next != "" && next != "/dashboard" {
		c.SetCookie(oauthNextCookie, next, 600, "/", "", isHTTPS(c), true)
	}
	c.Redirect(http.StatusFound, h.oauth2Config.AuthCodeURL(state))
}

// HandleGitHubCallback menangani callback dari GitHub OAuth.
func (h *AuthHandler) HandleGitHubCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	expectedState, _ := c.Cookie(oauthStateCookie)
	c.SetCookie(oauthStateCookie, "", -1, "/", "", false, true)

	// Tujuan setelah login; cookie selalu dihapus agar tidak dipakai dua kali.
	destination := "/dashboard"
	if next, err := c.Cookie(oauthNextCookie); err == nil {
		if valid := safeNextPath(next); valid != "" {
			destination = valid
		}
	}
	c.SetCookie(oauthNextCookie, "", -1, "/", "", false, true)

	if state == "" || expectedState == "" || subtle.ConstantTimeCompare([]byte(state), []byte(expectedState)) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is missing"})
		return
	}

	sessionID, err := h.ProcessGitHubCallback(c, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Session dikirim sebagai cookie httpOnly; tidak pernah lewat URL.
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, sessionID, int((24 * time.Hour).Seconds()), "/", "", isHTTPS(c), true)

	c.Redirect(http.StatusFound, h.frontendURL+destination)
}

// ProcessGitHubCallback memproses callback GitHub OAuth.
func (h *AuthHandler) ProcessGitHubCallback(c *gin.Context, code string) (string, error) {
	token, err := h.oauth2Config.Exchange(c, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange token: %w", err)
	}

	client := h.oauth2Config.Client(c, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return "", fmt.Errorf("failed to load user data: %w", err)
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", fmt.Errorf("failed to decode user info: %w", err)
	}

	if _, ok := userInfo["id"]; !ok {
		return "", fmt.Errorf("GitHub user data is incomplete")
	}

	githubID := fmt.Sprintf("%.0f", userInfo["id"])
	username, _ := userInfo["login"].(string)
	email := ""
	if m, ok := userInfo["email"].(string); ok {
		email = m
	}
	avatarURL := ""
	if m, ok := userInfo["avatar_url"].(string); ok {
		avatarURL = m
	}
	// Jika email kosong, coba ambil dari response GitHub dengan header Accept: application/vnd.github+json
	if email == "" {
		if emailResp, err := client.Get("https://api.github.com/user/emails"); err == nil {
			defer emailResp.Body.Close()
			var emails []map[string]interface{}
			if json.NewDecoder(emailResp.Body).Decode(&emails) == nil {
				for _, e := range emails {
					if primary, ok := e["primary"].(bool); ok && primary {
						if m, ok := e["email"].(string); ok {
							email = m
							break
						}
					}
				}
				if email == "" && len(emails) > 0 {
					if m, ok := emails[0]["email"].(string); ok {
						email = m
					}
				}
			}
		}
	}

	// Simpan user di database (upsert)
	user, err := h.userStore.GetByGitHubID(githubID)
	if err != nil {
		user, err = h.userStore.Create(githubID, username, email, avatarURL)
		if err != nil {
			return "", fmt.Errorf("failed to save user: %w", err)
		}
	}

	// Buat session dengan ID acak yang tidak bisa ditebak.
	sessionID, err := secure.NewSessionID()
	if err != nil {
		return "", fmt.Errorf("failed to create session id: %w", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	if _, err := h.sessionStore.Create(user.ID, sessionID, token.AccessToken, expiresAt); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

// GetMe mengembalikan info user yang sedang login.
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	user, err := h.userStore.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"github_id":  user.GitHubID,
			"username":   user.Username,
			"email":      user.Email,
			"avatar_url": user.AvatarURL,
		},
		"ok": true,
	})
}

// Logout menghapus session user (dari cookie maupun header Authorization).
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie(sessionCookieName)
	if err != nil || sessionID == "" {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			sessionID = strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if sessionID != "" {
		h.sessionStore.Delete(sessionID)
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, "", -1, "/", "", isHTTPS(c), true)
	c.JSON(http.StatusOK, gin.H{"message": "logout ok"})
}
